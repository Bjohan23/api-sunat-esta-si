/*
Utilidades para Comunicación con SUNAT
====================================

Este paquete maneja la comunicación directa con los webservices de SUNAT
para el envío de comprobantes electrónicos y recepción de CDR (Comprobante de Recepción).

Funcionalidades principales:
1. ZipXML() - Compresión del XML firmado según especificaciones SUNAT
2. BuildSOAP() - Construcción de mensaje SOAP con autenticación
3. SendToSunat() - Envío HTTP al webservice y procesamiento de respuesta
4. Procesamiento de CDR (Comprobante de Recepción) de SUNAT

Cumple con:
- Protocolo SOAP 1.1 con WS-Security
- Compresión ZIP según formato SUNAT
- Codificación Base64 para transferencia binaria
- Procesamiento de respuestas XML estructuradas

El flujo típico es: XML firmado → ZIP → SOAP → HTTP → CDR
*/
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
    "ubl-go-conversor/models"
)

/*
ZipXML comprime el archivo XML firmado en formato ZIP según especificaciones SUNAT.

SUNAT requiere que el XML firmado se envíe comprimido en un archivo ZIP que contenga:
- Un solo archivo XML con el mismo nombre base que el ZIP
- Extensión .ZIP en mayúsculas
- Sin carpetas internas, solo el archivo XML en la raíz

Proceso:
1. Crear archivo ZIP con nombre basado en el XML
2. Abrir el XML firmado para lectura
3. Comprimir el XML dentro del ZIP
4. Retornar la ruta del archivo ZIP creado

Parámetros:
- rutaXML: Ruta del archivo XML firmado a comprimir

Retorna:
- string: Ruta del archivo ZIP creado
- error: Error si falla el proceso de compresión
*/
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

/*
BuildSOAP construye el mensaje SOAP requerido para enviar comprobantes a SUNAT.

SUNAT utiliza un webservice SOAP que requiere:
1. Autenticación WS-Security con usuario y contraseña
2. Método sendBill con fileName y contentFile
3. contentFile debe ser el ZIP en formato Base64
4. Usuario formado por RUC + usuario secundario

Estructura del mensaje:
- Header: Contiene autenticación WS-Security
- Body: Contiene el método sendBill con parámetros

Parámetros:
- ruc: RUC del emisor (20123456789)
- usuario: Usuario secundario del certificado
- clave: Contraseña del usuario
- zipPath: Ruta del archivo ZIP a enviar

Retorna:
- string: Mensaje SOAP completo listo para envío HTTP
- error: Error si no puede leer el archivo ZIP
*/
func BuildSOAP(ruc, usuario, clave, zipPath string) (string, error) {
    // Leer contenido del archivo ZIP
    content, err := ioutil.ReadFile(zipPath)
    if err != nil {
        return "", err
    }
    
    // Codificar ZIP en Base64 para transmisión SOAP
    encoded := base64.StdEncoding.EncodeToString(content)
    
    // Extraer solo el nombre del archivo ZIP (sin ruta)
    zipName := filepath.Base(zipPath)

    // Construir mensaje SOAP según especificaciones SUNAT
    // El usuario debe ser RUC + usuario secundario (sin separador)
    // Ejemplo: "20123456789MODDATOS" donde "MODDATOS" es el usuario secundario
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

/*
SendToSunat es la función legacy que retorna la respuesta en formato JSON string.
Se mantiene para compatibilidad con código existente.

Esta función es un wrapper de SendToSunatStructured que convierte
la respuesta estructurada a JSON string con los campos básicos.

Retorna JSON con formato:
{
  "estado": "aprobada|rechazada|observada",
  "codigo": "0|4xxx|error_code", 
  "mensaje": "descripción del resultado"
}
*/
func SendToSunat(endpoint, soap, xmlZipName, baseCDRDir string) (string, error) {
    cdrInfo, err := SendToSunatStructured(endpoint, soap, xmlZipName, baseCDRDir)
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf(`{
  "estado": "%s",
  "codigo": "%s",
  "mensaje": "%s"
}`, cdrInfo.Estado, cdrInfo.ResponseCode, cdrInfo.Description), nil
}

/*
SendToSunatStructured realiza el envío del comprobante a SUNAT y procesa la respuesta CDR.

Esta es la función principal que implementa la comunicación completa con SUNAT:

1. ENVÍO HTTP:
   - Realiza POST al endpoint de SUNAT
   - Envía mensaje SOAP con autenticación WS-Security
   - Configura headers apropiados para SOAP

2. PROCESAMIENTO DE RESPUESTA:
   - Parsea respuesta SOAP de SUNAT
   - Extrae CDR (Comprobante de Recepción) en Base64
   - Decodifica y guarda CDR como archivo ZIP
   - Extrae XML del CDR para analizar el resultado

3. INTERPRETACIÓN DE CÓDIGOS:
   - Código "0": Aprobada
   - Códigos 4000-4999: Observada (aprobada con observaciones)
   - Otros códigos: Rechazada
   - Errores SOAP: Fallas de comunicación

Parámetros:
- endpoint: URL del webservice SUNAT
- soap: Mensaje SOAP completo para envío
- xmlZipName: Nombre del ZIP enviado (para nombrar CDR)
- baseCDRDir: Directorio base para guardar CDR

Retorna:
- *models.CDRInfo: Información estructurada de la respuesta
- error: Error si falla el proceso
*/
func SendToSunatStructured(endpoint, soap, xmlZipName, baseCDRDir string) (*models.CDRInfo, error) {
    // ==================== CONFIGURACIÓN Y ENVÍO HTTP ====================
    
    // Crear cliente HTTP estándar (se podría configurar timeout)
    client := &http.Client{}
    
    // Crear request POST con el mensaje SOAP como body
    req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(soap))
    if err != nil {
        return nil, err
    }

    // Configurar headers requeridos para SOAP
    req.Header.Set("Content-Type", `text/xml; charset="utf-8"`) // Tipo de contenido SOAP
    req.Header.Set("SOAPAction", "")                            // SOAPAction vacío según SUNAT

    // Enviar request a SUNAT
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // ==================== LECTURA Y PARSEO DE RESPUESTA SOAP ====================
    
    // Leer todo el contenido de la respuesta HTTP
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Estructura para parsear la respuesta SOAP de SUNAT
    // SUNAT puede retornar:
    // - sendBillResponse con applicationResponse (éxito)
    // - Fault con faultcode y faultstring (error)
    type Envelope struct {
        XMLName             xml.Name `xml:"Envelope"`
        ApplicationResponse string   `xml:"Body>sendBillResponse>applicationResponse"` // CDR en Base64
        FaultCode           string   `xml:"Body>Fault>faultcode"`                      // Código de error SOAP
        FaultString         string   `xml:"Body>Fault>faultstring"`                    // Descripción del error
    }

    // Parsear respuesta XML de SUNAT
    var envelope Envelope
    if err := xml.Unmarshal(bodyBytes, &envelope); err != nil {
        return nil, fmt.Errorf("error al parsear respuesta XML: %v", err)
    }

    // Verificar si hay un error SOAP (faultcode presente)
    if envelope.FaultCode != "" {
        // Retornar información de error sin procesar CDR
        return &models.CDRInfo{
            ResponseCode: envelope.FaultCode,
            Description:  envelope.FaultString,
            Estado:       "error",
        }, nil
    }

    // ==================== PROCESAMIENTO DEL CDR (COMPROBANTE DE RECEPCIÓN) ====================
    
    // Decodificar el applicationResponse que contiene el CDR en Base64
    // El CDR es un archivo ZIP que contiene el XML de respuesta de SUNAT
    decodedZip, err := base64.StdEncoding.DecodeString(envelope.ApplicationResponse)
    if err != nil {
        return nil, fmt.Errorf("error al decodificar base64: %v", err)
    }

    // Crear estructura de directorios para almacenar CDR
    // Formato: baseCDRDir/nombre_documento/
    zipBaseName := removeExtension(filepath.Base(xmlZipName)) 
    cdrDir := filepath.Join(baseCDRDir, zipBaseName)

    // Crear directorio si no existe
    if err := os.MkdirAll(cdrDir, 0755); err != nil {
        return nil, fmt.Errorf("error al crear carpeta CDR: %v", err)
    }

    // Guardar CDR ZIP con prefijo identificador
    // Formato: CDR-nombre_original.ZIP
    zipFileName := "CDR-" + filepath.Base(xmlZipName)
    zipFilePath := filepath.Join(cdrDir, zipFileName)
    if err := os.WriteFile(zipFilePath, decodedZip, 0644); err != nil {
        return nil, fmt.Errorf("error al guardar ZIP de respuesta: %v", err)
    }

    // Preparar CDR en Base64 para incluir en respuesta API
    cdrZipBase64 := base64.StdEncoding.EncodeToString(decodedZip)

    // ==================== EXTRACCIÓN Y ANÁLISIS DEL XML CDR ====================
    
    // Abrir CDR ZIP para extraer el XML de respuesta
    zipReader, err := zip.NewReader(bytes.NewReader(decodedZip), int64(len(decodedZip)))
    if err != nil {
        return nil, fmt.Errorf("error al leer ZIP: %v", err)
    }

    // Buscar archivo XML dentro del ZIP del CDR
    // SUNAT incluye un XML con la respuesta oficial del comprobante
    for _, file := range zipReader.File {
        // Verificar que sea un archivo XML (puede ser .XML o .xml)
        if ext := filepath.Ext(file.Name); ext == ".XML" || ext == ".xml" {
            // Abrir archivo XML del CDR
            rc, err := file.Open()
            if err != nil {
                return nil, err
            }
            defer rc.Close()

            // Leer contenido completo del XML
            content, err := io.ReadAll(rc)
            if err != nil {
                return nil, err
            }

            // Guardar XML del CDR como archivo separado para auditoría
            cdrXmlPath := filepath.Join(cdrDir, file.Name)
            if err := os.WriteFile(cdrXmlPath, content, 0644); err != nil {
                return nil, fmt.Errorf("error al guardar XML del CDR: %v", err)
            }

            // Estructura para parsear respuesta CDR de SUNAT
            // El CDR contiene ResponseCode y Description en DocumentResponse
            type CDR struct {
                ResponseCode string `xml:"DocumentResponse>Response>ResponseCode"` // Código de respuesta SUNAT
                Description  string `xml:"DocumentResponse>Response>Description"`  // Descripción del resultado
            }

            // Parsear XML del CDR para extraer resultado
            var cdr CDR
            if err := xml.Unmarshal(content, &cdr); err != nil {
                return nil, fmt.Errorf("error al parsear CDR: %v", err)
            }

            // ==================== INTERPRETACIÓN DE CÓDIGOS SUNAT ====================
            
            // Determinar estado final según código de respuesta SUNAT:
            // - "0": Aceptado (aprobada)
            // - "4000"-"4999": Aceptado con observaciones (observada)
            // - Otros códigos: Rechazado (rechazada)
            estado := "rechazada"
            if cdr.ResponseCode == "0" {
                estado = "aprobada"
            } else if cdr.ResponseCode >= "4000" && cdr.ResponseCode < "5000" {
                estado = "observada"
            }

            // Retornar información completa del CDR
            return &models.CDRInfo{
                ResponseCode: cdr.ResponseCode, // Código de respuesta SUNAT
                Description:  cdr.Description,  // Descripción oficial
                Estado:       estado,           // Estado interpretado
                CDRZipBase64: cdrZipBase64,     // CDR completo en Base64
                CDRZipPath:   zipFilePath,      // Ruta del archivo CDR guardado
            }, nil
        }
    }

    // Error si no se encuentra XML en el CDR (situación anómala)
    return nil, fmt.Errorf("no se encontró XML dentro del ZIP del CDR")
}


/*
removeExtension elimina la extensión de un nombre de archivo.

Utilidad helper para generar nombres de archivos relacionados:
- XML firmado: "documento.xml" → "documento"
- ZIP: "documento" → "documento.ZIP"
- CDR: "CDR-documento.ZIP"

Parámetros:
- file: Nombre del archivo con extensión

Retorna:
- string: Nombre del archivo sin extensión
*/
func removeExtension(file string) string {
    return file[:len(file)-len(filepath.Ext(file))]
}

