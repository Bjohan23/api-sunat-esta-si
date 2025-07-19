module ubl-go-conversor

go 1.23.10

require (
	github.com/beevik/etree v1.5.1                    // Manipulación y parseo de documentos XML (generación UBL, inserción firmas)
	github.com/google/uuid v1.6.0                     // Generación de UUIDs únicos para identificadores de documentos
	github.com/joho/godotenv v1.5.1                   // Carga de configuración desde archivos .env (BD, SUNAT, certificados)
	github.com/jung-kurt/gofpdf v1.16.2               // Generación de PDFs para representación impresa de facturas/boletas
	github.com/russellhaering/goxmldsig v1.5.0        // Firma digital XMLDSig según estándares W3C y SUNAT
	gorm.io/driver/mysql v1.5.7                      // Driver MySQL para conexión de base de datos
	gorm.io/gorm v1.25.12                            // ORM para persistencia de documentos y auditoría
	software.sslmate.com/src/go-pkcs12 v0.5.0        // Decodificación de certificados digitales PKCS#12 (.pfx)
)

require (
	github.com/go-sql-driver/mysql v1.7.0         // indirect - Driver nativo MySQL usado por GORM
	github.com/jinzhu/inflection v1.0.0           // indirect - Singularización/pluralización de nombres de tablas GORM
	github.com/jinzhu/now v1.1.5                  // indirect - Utilidades de tiempo y fecha para GORM
	github.com/jonboulle/clockwork v0.5.0         // indirect - Abstracción de tiempo para testing en goxmldsig
	golang.org/x/crypto v0.11.0                   // indirect - Primitivas criptográficas (RSA, SHA1, certificados X.509)
	golang.org/x/text v0.14.0                     // indirect - Procesamiento de texto y encoding para XML/SOAP
)
