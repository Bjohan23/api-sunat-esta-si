package repository

import (
	"gorm.io/gorm"
	"ubl-go-conversor/models"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateLog crea un nuevo log de auditoría
func (r *AuditRepository) CreateLog(documentID, action, details, userIP string) error {
	auditLog := &models.AuditLog{
		DocumentID: documentID,
		Action:     action,
		Details:    details,
		UserIP:     userIP,
	}
	return r.db.Create(auditLog).Error
}

// GetLogsByDocumentID obtiene todos los logs de un documento
func (r *AuditRepository) GetLogsByDocumentID(documentID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("document_id = ?", documentID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetRecentLogs obtiene los logs más recientes
func (r *AuditRepository) GetRecentLogs(limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// Actions constantes para acciones de auditoría
const (
	ActionCreated   = "created"
	ActionValidated = "validated"
	ActionSigned    = "signed"
	ActionSent      = "sent"
	ActionApproved  = "approved"
	ActionRejected  = "rejected"
	ActionError     = "error"
)