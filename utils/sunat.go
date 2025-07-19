package utils

import (
    "archive/zip"
    "bytes"
    "encoding/base64"
    "encoding/xml"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
)

// Paso 3: Crear ZIP con el XML firmado
func ZipXML(rutaXML string) (string, error) {
    zipName := removeExtension(rutaXML) + ".ZIP"
    zipFile, err := os.Create(zipName)
    if err != nil {
        return "", err
    }
    defer zipFile.Close()

    zipWriter := zip.NewWriter(zipFile)
    defer zipWriter.Close()

    xmlFile, err := os.Open(rutaXML)
    if err != nil {
        return "", err
    }
    defer xmlFile.Close()

    w, err := zipWriter.Create(fmt.Sprintf("%s.XML", removeExtension(filepath.Base(rutaXML))))
    if err != nil {
        return "", err
    }
    if _, err = io.Copy(w, xmlFile); err != nil {
        return "", err
    }

    return zipName, nil
}

// Paso 4: Preparar mensaje SOAP
func BuildSOAP(ruc, usuario, clave, zipPath string) (string, error) {
    content, err := ioutil.ReadFile(zipPath)
    if err != nil {
        return "", err
    }
    encoded := base64.StdEncoding.EncodeToString(content)
    zipName := filepath.Base(zipPath)

    soap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:ser="http://service.sunat.gob.pe"
    xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
  <soapenv:Header>
    <wsse:Security>
      <wsse:UsernameToken>
        <wsse:Username>%s%s</wsse:Username>
        <wsse:Password>%s</wsse:Password>
      </wsse:UsernameToken>
    </wsse:Security>
  </soapenv:Header>
  <soapenv:Body>
    <ser:sendBill>
      <fileName>%s</fileName>
      <contentFile>%s</contentFile>
    </ser:sendBill>
  </soapenv:Body>
</soapenv:Envelope>`, ruc, usuario, clave, zipName, encoded)

    return soap, nil
}

// Paso 5 y 6: Enviar a SUNAT y procesar respuesta
func SendToSunat(endpoint, soap, xmlZipName, baseCDRDir string) (string, error) {
    client := &http.Client{}
    req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(soap))
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
    req.Header.Set("SOAPAction", "")

    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    type Envelope struct {
        XMLName             xml.Name `xml:"Envelope"`
        ApplicationResponse string   `xml:"Body>sendBillResponse>applicationResponse"`
        FaultCode           string   `xml:"Body>Fault>faultcode"`
        FaultString         string   `xml:"Body>Fault>faultstring"`
    }

    var envelope Envelope
    if err := xml.Unmarshal(bodyBytes, &envelope); err != nil {
        return "", fmt.Errorf("error al parsear respuesta XML: %v", err)
    }

    if envelope.FaultCode != "" {
        return fmt.Sprintf(`{"estado":"error","codigo":"%s","mensaje":"%s"}`, envelope.FaultCode, envelope.FaultString), nil
    }

    decodedZip, err := base64.StdEncoding.DecodeString(envelope.ApplicationResponse)
    if err != nil {
        return "", fmt.Errorf("error al decodificar base64: %v", err)
    }

    zipBaseName := removeExtension(filepath.Base(xmlZipName)) 
    cdrDir := filepath.Join(baseCDRDir, zipBaseName)

    if err := os.MkdirAll(cdrDir, 0755); err != nil {
        return "", fmt.Errorf("error al crear carpeta CDR: %v", err)
    }

    zipFileName := "CDR-" + filepath.Base(xmlZipName)
    zipFilePath := filepath.Join(cdrDir, zipFileName)
    if err := os.WriteFile(zipFilePath, decodedZip, 0644); err != nil {
        return "", fmt.Errorf("error al guardar ZIP de respuesta: %v", err)
    }

    zipReader, err := zip.NewReader(bytes.NewReader(decodedZip), int64(len(decodedZip)))
    if err != nil {
        return "", fmt.Errorf("error al leer ZIP: %v", err)
    }

    for _, file := range zipReader.File {
        if ext := filepath.Ext(file.Name); ext == ".XML" || ext == ".xml" {
            rc, err := file.Open()
            if err != nil {
                return "", err
            }
            defer rc.Close()

            content, err := io.ReadAll(rc)
            if err != nil {
                return "", err
            }

            cdrXmlPath := filepath.Join(cdrDir, file.Name)
            if err := os.WriteFile(cdrXmlPath, content, 0644); err != nil {
                return "", fmt.Errorf("error al guardar XML del CDR: %v", err)
            }

            type CDR struct {
                ResponseCode string `xml:"DocumentResponse>Response>ResponseCode"`
                Description  string `xml:"DocumentResponse>Response>Description"`
            }

            var cdr CDR
            if err := xml.Unmarshal(content, &cdr); err != nil {
                return "", fmt.Errorf("error al parsear CDR: %v", err)
            }

            estado := "rechazada"
            if cdr.ResponseCode == "0" {
                estado = "aprobada"
            }

            return fmt.Sprintf(`{
  "estado": "%s",
  "codigo": "%s",
  "mensaje": "%s"
}`, estado, cdr.ResponseCode, cdr.Description), nil
        }
    }

    return "", fmt.Errorf("no se encontró XML dentro del ZIP del CDR")
}


func removeExtension(file string) string {
    return file[:len(file)-len(filepath.Ext(file))]
}

