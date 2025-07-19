package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"ubl-go-conversor/config"
	conversor "ubl-go-conversor/converters"
	"ubl-go-conversor/models"
	"ubl-go-conversor/pdf"
	"ubl-go-conversor/signature"
	"ubl-go-conversor/utils"
	"ubl-go-conversor/validator"
)

var appConfig *config.Config

func main() {
	// Cargar configuración
	appConfig = config.Load()
	
	http.HandleFunc("/api/v1/invoices", manerjarDocumento)
	http.HandleFunc("/api/v1/documents/", manerjarDocumentos)
	
	serverAddr := ":" + appConfig.Server.Port
	fmt.Printf("Servidor iniciado en http://%s%s\n", appConfig.Server.Host, serverAddr)
	
	err := http.ListenAndServe(serverAddr, nil)
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
	digest, signatureValue, err := signature.FirmaXML(nombreXML, appConfig.Certificate.Path, appConfig.Certificate.Password)
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
	Usuario := appConfig.SUNAT.Username
	Clave := appConfig.SUNAT.Password

	soapMessage, err := utils.BuildSOAP(documento.Emisor.RUC, Usuario, Clave, zipPath)
	if err != nil {
		http.Error(w, "Error al construir SOAP: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("PASO 4: SOAP generado.")

	// Paso 5: Enviar a SUNAT
	cdrInfo, err := utils.SendToSunatStructured(appConfig.SUNAT.URL, soapMessage, zipPath, "cdr")
	if err != nil {
		errorResponse := models.ErrorResponse{
			Estado:      "error",
			Code:        "500",
			Description: "Error al enviar a SUNAT",
			Details:     err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	fmt.Println("PASO 5 y 6: CDR recibido.")

	// Leer archivos para incluir en respuesta
	xmlContent, _ := ioutil.ReadFile(nombreXML)
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlContent)
	
	// Generar PDF
	pdfPath := pdf.GeneratePDFPath(documento)
	err = pdf.GeneratePDF(documento, pdfPath)
	if err != nil {
		fmt.Printf("Warning: No se pudo generar PDF: %v\n", err)
	}
	
	// Crear URL del PDF y ID del documento
	documentID := fmt.Sprintf("%s-%s-%s-%s", documento.Emisor.RUC, documento.TipoDocumento, documento.Serie, documento.Numero)
	pdfURL := fmt.Sprintf("http://%s:%s/api/v1/documents/%s/pdf", appConfig.Server.Host, appConfig.Server.Port, documentID)
	
	// Preparar respuesta según requerimientos
	response := models.APIResponse{
		Estado:      cdrInfo.Estado,
		Code:        cdrInfo.ResponseCode,
		Description: fmt.Sprintf("La Factura numero %s-%s, ha sido %s", documento.Serie, documento.Numero, cdrInfo.Estado),
		Hash:        fmt.Sprintf("SHA1:%s|RSA:%s", digest, signatureValue),
		CDRZip:      cdrInfo.CDRZipBase64,
		XMLFirmado:  xmlBase64,
		PDFURL:      pdfURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// manerjarDocumentos maneja las rutas de documentos (PDF, XML, etc.)
func manerjarDocumentos(w http.ResponseWriter, r *http.Request) {
	// Extraer el path después de /api/v1/documents/
	path := r.URL.Path[len("/api/v1/documents/"):]
	
	// Dividir el path para obtener el ID del documento y el tipo
	parts := splitPath(path)
	if len(parts) < 2 {
		http.Error(w, "Ruta inválida. Use /api/v1/documents/{id}/pdf", http.StatusBadRequest)
		return
	}
	
	documentID := parts[0]
	action := parts[1]
	
	switch action {
	case "pdf":
		servirPDF(w, r, documentID)
	case "xml":
		servirXML(w, r, documentID)
	case "status":
		consultarEstado(w, r, documentID)
	default:
		http.Error(w, "Acción no soportada. Use: pdf, xml, status", http.StatusBadRequest)
	}
}

// servirPDF sirve el archivo PDF del documento
func servirPDF(w http.ResponseWriter, r *http.Request, documentID string) {
	// Por ahora buscar en la carpeta out/ usando el documentID
	pdfPath := fmt.Sprintf("out/%s.pdf", documentID)
	
	// Verificar si el archivo existe
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		http.Error(w, "PDF no encontrado", http.StatusNotFound)
		return
	}
	
	// Servir el archivo PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", documentID))
	http.ServeFile(w, r, pdfPath)
}

// servirXML sirve el archivo XML del documento
func servirXML(w http.ResponseWriter, r *http.Request, documentID string) {
	xmlPath := fmt.Sprintf("out/%s.xml", documentID)
	
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		http.Error(w, "XML no encontrado", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xml", documentID))
	http.ServeFile(w, r, xmlPath)
}

// consultarEstado consulta el estado del documento (placeholder)
func consultarEstado(w http.ResponseWriter, r *http.Request, documentID string) {
	// Por ahora retorna un estado simple
	status := map[string]interface{}{
		"document_id": documentID,
		"status":      "processed",
		"message":     "Documento procesado correctamente",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// splitPath divide un path en partes separadas por /
func splitPath(path string) []string {
	var parts []string
	for _, part := range splitString(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

// splitString divide un string por un separador
func splitString(s, sep string) []string {
	var result []string
	current := ""
	
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}
