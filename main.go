/*
API de Facturación Electrónica para SUNAT - Perú
===============================================

Este es el punto de entrada principal de la API REST que maneja la generación,
firma digital y envío de comprobantes electrónicos (facturas y boletas) a SUNAT
siguiendo el estándar UBL 2.1 con extensiones SUNAT.

Flujo principal:
1. Recibe JSON con datos del comprobante
2. Valida datos según normativas SUNAT
3. Genera XML UBL 2.1 con extensiones SUNAT
4. Firma digitalmente el XML usando certificado PKCS#12
5. Comprime el XML firmado en ZIP
6. Construye mensaje SOAP para SUNAT
7. Envía a SUNAT y procesa respuesta CDR
8. Genera PDF de representación impresa
9. Almacena todo en base de datos con auditoría
10. Retorna respuesta estructurada al cliente

Arquitectura:
- config: Configuración externa (BD, SUNAT, certificados)
- models: Estructuras de datos (comprobantes, respuestas, BD)
- validators: Validaciones de negocio según SUNAT
- converters: Generación de XML UBL 2.1
- signature: Firma digital con certificados X.509
- utils: Comunicación SOAP con SUNAT
- database: Persistencia y auditoría
- pdf: Generación de representación impresa
*/
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
	"ubl-go-conversor/database"
	"ubl-go-conversor/models"
	"ubl-go-conversor/pdf"
	"ubl-go-conversor/repository"
	"ubl-go-conversor/signature"
	"ubl-go-conversor/utils"
	"ubl-go-conversor/validator"
)

// Variables globales para configuración y repositorios
// Estas se inicializan una vez al arrancar la aplicación
var appConfig *config.Config           // Configuración de la aplicación (.env)
var docRepo *repository.DocumentRepository // Repositorio para operaciones de documentos
var auditRepo *repository.AuditRepository   // Repositorio para logs de auditoría

// main es el punto de entrada de la aplicación
// Inicializa todos los componentes necesarios y arranca el servidor HTTP
func main() {
	// PASO 1: Cargar configuración desde .env y variables de entorno
	appConfig = config.Load()
	
	// PASO 2: Inicializar conexión a MySQL y crear tablas si no existen
	if err := database.Initialize(appConfig); err != nil {
		log.Fatal("Error inicializando base de datos:", err)
	}
	
	// PASO 3: Inicializar repositorios para operaciones de base de datos
	db := database.GetDB()
	docRepo = repository.NewDocumentRepository(db)
	auditRepo = repository.NewAuditRepository(db)
	
	// PASO 4: Configurar rutas HTTP
	// POST /api/v1/invoices - Endpoint principal para crear facturas/boletas
	http.HandleFunc("/api/v1/invoices", manerjarDocumento)
	// GET /api/v1/documents/{id}/{action} - Endpoints para consultar documentos
	http.HandleFunc("/api/v1/documents/", manerjarDocumentos)
	
	// PASO 5: Arrancar servidor HTTP
	serverAddr := ":" + appConfig.Server.Port
	fmt.Printf("Servidor iniciado en http://%s%s\n", appConfig.Server.Host, serverAddr)
	
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		log.Fatal("Error al iniciar servidor:", err)
	}
}

/*
manerjarDocumento es el endpoint principal que procesa facturas y boletas electrónicas
Implementa el flujo completo desde la recepción del JSON hasta el envío a SUNAT

Proceso de 6 pasos según normativa SUNAT:
1. Validación de datos de entrada
2. Generación de XML UBL 2.1 
3. Firma digital del XML
4. Compresión en ZIP
5. Construcción de mensaje SOAP
6. Envío a SUNAT y procesamiento de CDR

Además incluye:
- Persistencia en base de datos con auditoría
- Generación de PDF de representación impresa
- Respuesta estructurada según requerimientos
*/
func manerjarDocumento(w http.ResponseWriter, r *http.Request) {
	// ==================== VALIDACIÓN DE ENTRADA ====================
	
	// Solo acepta método POST para crear documentos
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parsear JSON del request a estructura ComprobanteBase
	// Esta estructura contiene todos los datos necesarios para generar la factura/boleta
	var documento models.ComprobanteBase
	err := json.NewDecoder(r.Body).Decode(&documento)
	if err != nil {
		http.Error(w, "Error al leer JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validar datos según normativas SUNAT (RUC, series, totales, etc.)
	// El validator verifica reglas de negocio específicas de facturación electrónica
	err = validator.ValidarComprobanteBase(documento)
	if err != nil {
		http.Error(w, "Error de validación: "+err.Error(), http.StatusBadRequest)
		return
	}

	// ==================== PERSISTENCIA INICIAL ====================
	
	// Generar ID único del documento: RUC-TipoDoc-Serie-Numero
	// Ejemplo: "20123456789-01-F001-00000123"
	documentID := models.GenerateDocumentID(documento.Emisor.RUC, documento.TipoDocumento, documento.Serie, documento.Numero)
	
	// Crear registro inicial en base de datos con estado "processing"
	// Esto permite rastrear el documento desde el inicio del proceso
	dbDocument := &models.Document{
		ID:         documentID,           // ID único del documento
		RUC:        documento.Emisor.RUC, // RUC del emisor
		TipoDoc:    documento.TipoDocumento, // 01=Factura, 03=Boleta
		Serie:      documento.Serie,      // Serie del comprobante (F001, B001)
		Numero:     documento.Numero,     // Número correlativo
		Cliente:    documento.Cliente.RazonSocial, // Nombre/razón social del cliente
		ClienteDoc: documento.Cliente.NumeroDoc,   // DNI/RUC del cliente
		Total:      documento.TotalImportePagar,   // Importe total a pagar
		Moneda:     documento.Moneda,     // PEN, USD, EUR
		Estado:     models.StatusProcessing, // Estado inicial: "processing"
	}
	
	// Guardar en base de datos - si falla, abortar proceso
	if err := docRepo.Create(dbDocument); err != nil {
		http.Error(w, "Error al crear documento en BD: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Registrar acción de creación en logs de auditoría
	auditRepo.CreateLog(documentID, repository.ActionCreated, "Documento creado", r.RemoteAddr)

	// ==================== PASO 1: GENERACIÓN DE XML UBL 2.1 ====================
	
	// Crear directorio de salida si no existe
	if _, err := os.Stat("out"); os.IsNotExist(err) {
		err = os.Mkdir("out", 0755)
		if err != nil {
			http.Error(w, "Error al crear carpeta: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Generar nombre del archivo XML con formato estándar SUNAT
	// Formato: RUC-TipoDocumento-Serie-Numero.xml
	// Ejemplo: "20123456789-01-F001-00000123.xml"
	nombreXML := fmt.Sprintf("out/%s-%s-%s-%s.xml", documento.Emisor.RUC, documento.TipoDocumento, documento.Serie, documento.Numero)

	// Generar XML UBL 2.1 según el tipo de documento
	// Solo soporta facturas (01) y boletas (03) por ahora
	if documento.TipoDocumento == "01" || documento.TipoDocumento == "03" {
		// El conversor transforma la estructura ComprobanteBase a XML UBL 2.1
		// Incluye todas las extensiones SUNAT requeridas y validaciones de estructura
		err = conversor.GenerarXMLBF(documento, nombreXML)
		if err != nil {
			http.Error(w, "Error al generar XML: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("PASO 1: XML generado exitosamente: %s\n", nombreXML)
	} else {
		// Rechazar tipos de documento no implementados (notas de crédito/débito)
		http.Error(w, "Tipo de documento no soportado: "+documento.TipoDocumento, http.StatusBadRequest)
		return
	}

	// ==================== PASO 2: FIRMA DIGITAL ====================
	
	// Firmar XML usando certificado digital PKCS#12
	// La firma cumple con estándares XMLDSig y normativas SUNAT
	// Retorna: digest (SHA1) y signatureValue (RSA)
	digest, signatureValue, err := signature.FirmaXML(
		nombreXML,                    // Archivo XML a firmar
		appConfig.Certificate.Path,   // Ruta del certificado .pfx
		appConfig.Certificate.Password, // Contraseña del certificado
	)
	if err != nil {
		http.Error(w, "Error al firmar XML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("PASO 2: XML firmado correctamente.")
	fmt.Println("Hash SHA1 (DigestValue):", digest)        // Hash del contenido firmado
	fmt.Println("Firma RSA (SignatureValue):", signatureValue) // Firma digital RSA
	
	// Guardar hashes de la firma en base de datos para auditoría
	docRepo.UpdateHashes(documentID, digest, signatureValue)
	auditRepo.CreateLog(documentID, repository.ActionSigned, "XML firmado digitalmente", r.RemoteAddr)
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

	// Actualizar estado en BD según respuesta SUNAT
	var estadoDB string
	switch cdrInfo.Estado {
	case "aprobada":
		estadoDB = models.StatusApproved
		auditRepo.CreateLog(documentID, repository.ActionApproved, "Documento aprobado por SUNAT", r.RemoteAddr)
	case "rechazada":
		estadoDB = models.StatusRejected
		auditRepo.CreateLog(documentID, repository.ActionRejected, "Documento rechazado por SUNAT", r.RemoteAddr)
	case "observada":
		estadoDB = models.StatusObserved
		auditRepo.CreateLog(documentID, repository.ActionError, "Documento observado por SUNAT", r.RemoteAddr)
	default:
		estadoDB = models.StatusError
		auditRepo.CreateLog(documentID, repository.ActionError, "Error en respuesta SUNAT", r.RemoteAddr)
	}
	
	docRepo.UpdateStatus(documentID, estadoDB, cdrInfo.ResponseCode, cdrInfo.Description)

	// Leer archivos para incluir en respuesta
	xmlContent, _ := ioutil.ReadFile(nombreXML)
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlContent)
	
	// Generar PDF
	pdfPath := pdf.GeneratePDFPath(documento)
	err = pdf.GeneratePDF(documento, pdfPath)
	if err != nil {
		fmt.Printf("Warning: No se pudo generar PDF: %v\n", err)
	}
	
	// Actualizar rutas de archivos en BD
	docRepo.UpdateFilePaths(documentID, nombreXML, pdfPath, cdrInfo.CDRZipPath, zipPath)
	
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

// consultarEstado consulta el estado del documento desde la BD
func consultarEstado(w http.ResponseWriter, r *http.Request, documentID string) {
	// Buscar documento en la base de datos
	doc, err := docRepo.GetByID(documentID)
	if err != nil {
		http.Error(w, "Documento no encontrado", http.StatusNotFound)
		return
	}
	
	// Obtener logs de auditoría
	logs, _ := auditRepo.GetLogsByDocumentID(documentID)
	
	status := map[string]interface{}{
		"document_id":    doc.ID,
		"ruc":           doc.RUC,
		"tipo_documento": doc.TipoDoc,
		"serie":         doc.Serie,
		"numero":        doc.Numero,
		"cliente":       doc.Cliente,
		"total":         doc.Total,
		"moneda":        doc.Moneda,
		"estado":        doc.Estado,
		"codigo_sunat":  doc.CodigoSUNAT,
		"mensaje_sunat": doc.MensajeSUNAT,
		"created_at":    doc.CreatedAt,
		"updated_at":    doc.UpdatedAt,
		"processed_at":  doc.ProcessedAt,
		"files": map[string]string{
			"xml": doc.XMLPath,
			"pdf": doc.PDFPath,
			"cdr": doc.CDRPath,
			"zip": doc.ZIPPath,
		},
		"hashes": map[string]string{
			"sha1": doc.HashSHA1,
			"rsa":  doc.HashRSA,
		},
		"audit_logs": logs,
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
