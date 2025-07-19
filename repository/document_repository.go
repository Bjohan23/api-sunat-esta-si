package repository

import (
	"time"

	"gorm.io/gorm"
	"ubl-go-conversor/models"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create crea un nuevo documento en la base de datos
func (r *DocumentRepository) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

// GetByID busca un documento por su ID
func (r *DocumentRepository) GetByID(id string) (*models.Document, error) {
	var doc models.Document
	err := r.db.Preload("Items").First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetByRUCSerieNumero busca un documento por RUC, serie y número
func (r *DocumentRepository) GetByRUCSerieNumero(ruc, serie, numero string) (*models.Document, error) {
	var doc models.Document
	err := r.db.Preload("Items").First(&doc, "ruc = ? AND serie = ? AND numero = ?", ruc, serie, numero).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// Update actualiza un documento existente
func (r *DocumentRepository) Update(doc *models.Document) error {
	return r.db.Save(doc).Error
}

// UpdateStatus actualiza solo el estado y información SUNAT
func (r *DocumentRepository) UpdateStatus(id, estado, codigoSUNAT, mensajeSUNAT string) error {
	updates := map[string]interface{}{
		"estado":        estado,
		"codigo_sunat":  codigoSUNAT,
		"mensaje_sunat": mensajeSUNAT,
		"updated_at":    time.Now(),
	}
	
	if estado == models.StatusApproved || estado == models.StatusRejected {
		updates["processed_at"] = time.Now()
	}
	
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateFilePaths actualiza las rutas de archivos generados
func (r *DocumentRepository) UpdateFilePaths(id string, xmlPath, pdfPath, cdrPath, zipPath string) error {
	updates := map[string]interface{}{
		"xml_path": xmlPath,
		"pdf_path": pdfPath,
		"cdr_path": cdrPath,
		"zip_path": zipPath,
		"updated_at": time.Now(),
	}
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateHashes actualiza los hashes de firma digital
func (r *DocumentRepository) UpdateHashes(id, hashSHA1, hashRSA string) error {
	updates := map[string]interface{}{
		"hash_sha1": hashSHA1,
		"hash_rsa":  hashRSA,
		"updated_at": time.Now(),
	}
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

// GetByRUC obtiene todos los documentos de un RUC
func (r *DocumentRepository) GetByRUC(ruc string, limit, offset int) ([]models.Document, error) {
	var docs []models.Document
	err := r.db.Where("ruc = ?", ruc).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&docs).Error
	return docs, err
}

// GetByStatus obtiene documentos por estado
func (r *DocumentRepository) GetByStatus(estado string, limit, offset int) ([]models.Document, error) {
	var docs []models.Document
	err := r.db.Where("estado = ?", estado).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&docs).Error
	return docs, err
}

// Delete elimina un documento (soft delete)
func (r *DocumentRepository) Delete(id string) error {
	return r.db.Delete(&models.Document{}, "id = ?", id).Error
}

// CreateItem crea un item de documento
func (r *DocumentRepository) CreateItem(item *models.DocumentItem) error {
	return r.db.Create(item).Error
}

// CreateItems crea múltiples items de documento
func (r *DocumentRepository) CreateItems(items []models.DocumentItem) error {
	return r.db.Create(&items).Error
}

// GetDocumentStats obtiene estadísticas de documentos
func (r *DocumentRepository) GetDocumentStats(ruc string) (map[string]interface{}, error) {
	var stats struct {
		Total     int64
		Aprobados int64
		Rechazados int64
		Pendientes int64
	}
	
	query := r.db.Model(&models.Document{})
	if ruc != "" {
		query = query.Where("ruc = ?", ruc)
	}
	
	query.Count(&stats.Total)
	query.Where("estado = ?", models.StatusApproved).Count(&stats.Aprobados)
	query.Where("estado = ?", models.StatusRejected).Count(&stats.Rechazados)
	query.Where("estado IN ?", []string{models.StatusPending, models.StatusProcessing}).Count(&stats.Pendientes)
	
	return map[string]interface{}{
		"total":      stats.Total,
		"aprobados":  stats.Aprobados,
		"rechazados": stats.Rechazados,
		"pendientes": stats.Pendientes,
	}, nil
}