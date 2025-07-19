package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SUNAT struct {
		URL      string
		Username string
		Password string
	}
	Server struct {
		Port string
		Host string
	}
	Certificate struct {
		Path     string
		Password string
	}
	Database struct {
		Host     string
		Port     string
		Name     string
		User     string
		Password string
	}
	Environment string
	LogLevel    string
}

func Load() *Config {
	// Cargar archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{}

	// Configuración SUNAT
	config.SUNAT.URL = getEnv("SUNAT_URL", "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService")
	config.SUNAT.Username = getEnv("SUNAT_USERNAME", "MODDATOS")
	config.SUNAT.Password = getEnv("SUNAT_PASSWORD", "MODDATOS")

	// Configuración del servidor
	config.Server.Port = getEnv("SERVER_PORT", "8080")
	config.Server.Host = getEnv("SERVER_HOST", "localhost")

	// Configuración de certificados
	config.Certificate.Path = getEnv("CERT_PATH", "certificados/certificado_prueba.pfx")
	config.Certificate.Password = getEnv("CERT_PASSWORD", "institutoisi")

	// Configuración de base de datos
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnv("DB_PORT", "5432")
	config.Database.Name = getEnv("DB_NAME", "facturacion_electronica")
	config.Database.User = getEnv("DB_USER", "postgres")
	config.Database.Password = getEnv("DB_PASSWORD", "password")

	// Configuración general
	config.Environment = getEnv("ENVIRONMENT", "development")
	config.LogLevel = getEnv("LOG_LEVEL", "info")

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}