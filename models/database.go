package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Document representa un comprobante electrónico en la base de datos
type Document struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(100)"`
	RUC         string    `json:"ruc" gorm:"type:varchar(11);index"`
	TipoDoc     string    `json:"tipo_doc" gorm:"type:varchar(2)"`
	Serie       string    `json:"serie" gorm:"type:varchar(4)"`
	Numero      string    `json:"numero" gorm:"type:varchar(8)"`
	Cliente     string    `json:"cliente" gorm:"type:varchar(500)"`
	ClienteDoc  string    `json:"cliente_doc" gorm:"type:varchar(20)"`
	Total       float64   `json:"total" gorm:"type:decimal(10,2)"`
	Moneda      string    `json:"moneda" gorm:"type:varchar(3)"`
	
	// Estados y procesamiento
	Estado      string    `json:"estado" gorm:"type:varchar(20);default:'pending'"` // pending, processing, approved, rejected, error
	CodigoSUNAT string    `json:"codigo_sunat" gorm:"type:varchar(10)"`
	MensajeSUNAT string   `json:"mensaje_sunat" gorm:"type:text"`
	
	// Archivos generados
	XMLPath     string    `json:"xml_path" gorm:"type:varchar(500)"`
	PDFPath     string    `json:"pdf_path" gorm:"type:varchar(500)"`
	CDRPath     string    `json:"cdr_path" gorm:"type:varchar(500)"`
	ZIPPath     string    `json:"zip_path" gorm:"type:varchar(500)"`
	
	// Hashes y firmas
	HashSHA1    string    `json:"hash_sha1" gorm:"type:varchar(100)"`
	HashRSA     string    `json:"hash_rsa" gorm:"type:varchar(500)"`
	
	// Metadata
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	
	// Relaciones
	Items       []DocumentItem `json:"items,omitempty" gorm:"foreignKey:DocumentID"`
}

// DocumentItem representa un item/línea de un comprobante
type DocumentItem struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	DocumentID   string  `json:"document_id" gorm:"type:varchar(100);index"`
	ItemNumber   int     `json:"item_number"`
	Codigo       string  `json:"codigo" gorm:"type:varchar(50)"`
	Descripcion  string  `json:"descripcion" gorm:"type:varchar(500)"`
	Cantidad     float64 `json:"cantidad" gorm:"type:decimal(10,4)"`
	ValorUnit    float64 `json:"valor_unitario" gorm:"type:decimal(10,4)"`
	ValorTotal   float64 `json:"valor_total" gorm:"type:decimal(10,2)"`
	IGV          float64 `json:"igv" gorm:"type:decimal(10,2)"`
	TipoAfecIGV  string  `json:"tipo_afectacion_igv" gorm:"type:varchar(2)"`
	
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLog para trazabilidad de operaciones
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DocumentID string    `json:"document_id" gorm:"type:varchar(100);index"`
	Action     string    `json:"action" gorm:"type:varchar(50)"` // created, validated, signed, sent, approved, rejected
	Details    string    `json:"details" gorm:"type:text"`
	UserIP     string    `json:"user_ip" gorm:"type:varchar(45)"`
	CreatedAt  time.Time `json:"created_at"`
}

// BeforeCreate genera un UUID para nuevos documentos
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// GetDocumentID genera un ID único basado en RUC-TipoDoc-Serie-Numero
func GenerateDocumentID(ruc, tipoDoc, serie, numero string) string {
	return ruc + "-" + tipoDoc + "-" + serie + "-" + numero
}

// DocumentStatus constantes para estados de documentos
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusError      = "error"
	StatusObserved   = "observed"
)

// DocumentType constantes para tipos de documentos
const (
	TypeFactura = "01"
	TypeBoleta  = "03"
	TypeCredito = "07"
	TypeDebito  = "08"
)