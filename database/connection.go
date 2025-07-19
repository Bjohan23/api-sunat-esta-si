package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"ubl-go-conversor/config"
	"ubl-go-conversor/models"
)

var DB *gorm.DB

// Initialize conecta a la base de datos MySQL
func Initialize(cfg *config.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("error conectando a la base de datos: %v", err)
	}

	log.Println("Conexión a MySQL establecida correctamente")

	// Auto migración de tablas
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("error en migración: %v", err)
	}

	return nil
}

// AutoMigrate crea/actualiza las tablas en la base de datos
func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.Document{},
		&models.DocumentItem{},
		&models.AuditLog{},
	)
}

// GetDB retorna la instancia de la base de datos
func GetDB() *gorm.DB {
	return DB
}