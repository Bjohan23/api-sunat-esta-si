package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	conversor "ubl-go-conversor/converters"
	"ubl-go-conversor/models"
	"ubl-go-conversor/signature"
	"ubl-go-conversor/utils"
	"ubl-go-conversor/validator"
)

func main() {
	http.HandleFunc("/EnviarSunat", manerjarDocumento)
	fmt.Println("Servidor iniciado en http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error al iniciar servidor:", err)
	}
}

func manerjarDocumento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var documento models.ComprobanteBase
	err := json.NewDecoder(r.Body).Decode(&documento)
	if err != nil {
		http.Error(w, "Error al leer JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = validator.ValidarComprobanteBase(documento)
	if err != nil {
		http.Error(w, "Error de validación: "+err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := os.Stat("out"); os.IsNotExist(err) {
		err = os.Mkdir("out", 0755)
		if err != nil {
			http.Error(w, "Error al crear carpeta: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	nombreXML := fmt.Sprintf("out/%s-%s-%s-%s.xml", documento.Emisor.RUC, documento.TipoDocumento, documento.Serie, documento.Numero)

	// Paso 1: Generar XML
	if documento.TipoDocumento == "01" || documento.TipoDocumento == "03" {
		err = conversor.GenerarXMLBF(documento, nombreXML)
		if err != nil {
			http.Error(w, "Error al generar XML: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("PASO 1: XML generado exitosamente: %s\n", nombreXML)
	} else {
		http.Error(w, "Tipo de documento no soportado: "+documento.TipoDocumento, http.StatusBadRequest)
		return
	}

	// Paso 2: Firmar XML
	digest, signatureValue, err := signature.FirmaXML(nombreXML, "certificados/certificado_prueba.pfx", "institutoisi")
	if err != nil {
		http.Error(w, "Error al firmar XML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("PASO 2: XML firmado correctamente.")
	fmt.Println("Hash SHA1 (DigestValue):", digest)
	fmt.Println("Firma RSA (SignatureValue):", signatureValue)
	// Paso 3: Comprimir ZIP
	var zipPath string
	zipParam := r.URL.Query().Get("zip")
	if zipParam != "" {
		zipPath = "out/" + zipParam
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			http.Error(w, "ZIP especificado no encontrado: "+zipPath, http.StatusBadRequest)
			return
		}
		fmt.Println("PASO 3: ZIP proporcionado manualmente:", zipPath)
	} else {
		zipPath, err = utils.ZipXML(nombreXML)
		if err != nil {
			http.Error(w, "Error al comprimir XML: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("PASO 3: ZIP creado automáticamente:", zipPath)
	}

	// Paso 4: Construir SOAP
	Usuario := "MODDATOS"
	Clave := "MODDATOS"

	soapMessage, err := utils.BuildSOAP(documento.Emisor.RUC, Usuario, Clave, zipPath)
	if err != nil {
		http.Error(w, "Error al construir SOAP: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("PASO 4: SOAP generado.")

	// Paso 5: Enviar a SUNAT
	sunatResponse, err := utils.SendToSunat("https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", soapMessage, zipPath, "cdr")
	if err != nil {
		http.Error(w, "Error al enviar a SUNAT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("PASO 5 y 6: CDR recibido.")

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprint(w, sunatResponse)
}
