package models

// APIResponse estructura de respuesta según requerimientos funcionales
type APIResponse struct {
	Estado      string `json:"estado"`                // aceptado, observado, rechazado
	Code        string `json:"code"`                  // Código de respuesta SUNAT
	Description string `json:"description"`           // Descripción detallada
	Hash        string `json:"hash,omitempty"`        // Hash del documento
	CDRZip      string `json:"cdr_zip,omitempty"`     // CDR en base64
	XMLFirmado  string `json:"xml_firmado,omitempty"` // XML firmado en base64
	PDFURL      string `json:"pdf_url,omitempty"`     // URL del PDF (futuro)
}

// ErrorResponse estructura para errores
type ErrorResponse struct {
	Estado      string `json:"estado"`      // "error"
	Code        string `json:"code"`        // Código de error
	Description string `json:"description"` // Descripción del error
	Details     string `json:"details,omitempty"` // Detalles adicionales
}

// CDRInfo información extraída del CDR
type CDRInfo struct {
	ResponseCode string `json:"response_code"`
	Description  string `json:"description"`
	Estado       string `json:"estado"` // calculado basado en response_code
	CDRZipBase64 string `json:"cdr_zip_base64,omitempty"` // CDR en base64
	CDRZipPath   string `json:"cdr_zip_path,omitempty"`   // Ruta del archivo CDR
}