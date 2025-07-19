# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based UBL (Universal Business Language) converter for SUNAT (Peru's tax authority) electronic invoicing. The application converts invoice/receipt JSON data into UBL XML format, digitally signs it, and sends it to SUNAT's billing service.

## Development Commands

### Running the application
```bash
go run main.go
```
The server starts on `http://localhost:8080` with endpoint `/EnviarSunat`

### Building the application
```bash
go build -o ubl-converter main.go
```

### Running tests
```bash
go test ./...
```

### Managing dependencies
```bash
go mod tidy
go mod download
```

## Architecture Overview

The application follows a layered architecture with clear separation of concerns:

### Core Packages

- **`main.go`**: HTTP server that exposes `/EnviarSunat` POST endpoint
- **`models/`**: Data structures for invoice/receipt information (`ComprobanteBase`)
- **`validator/`**: Business rule validation for SUNAT compliance
- **`converters/`**: UBL XML generation from JSON input
- **`signature/`**: Digital signature functionality using X.509 certificates
- **`utils/`**: SOAP message building and SUNAT communication
- **`cdr/`**: Directory for storing CDR (Comprobante de Recepción) responses

### Processing Pipeline

The application follows a 6-step process:
1. **Validation**: Validate input JSON against SUNAT business rules
2. **XML Generation**: Convert to UBL 2.1 XML format
3. **Digital Signing**: Sign XML with PKCS#12 certificate
4. **ZIP Compression**: Package signed XML into ZIP file
5. **SOAP Message**: Build SOAP envelope for SUNAT webservice
6. **SUNAT Submission**: Send to SUNAT and process CDR response

### Key Components

**Document Types Supported:**
- `01`: Factura (Invoice) - must be issued to RUC holders
- `03`: Boleta (Receipt) - issued to individuals (DNI)

**Tax Affectation Types:**
- `10-17`: Gravado (taxable with 18% IGV)
- `20`: Exonerado (tax exempt)
- `21`: Gratuito (free transfer)
- `30-37`: Inafecto (unaffected by tax)
- `40`: Exportación (export)

**Certificate Management:**
- Uses PKCS#12 certificates stored in `certificados/` directory
- Default test certificate: `certificado_prueba.pfx` with password "institutoisi"

## Important Files

- **`models/comprobante_base.go`**: Core data structures for invoices/receipts
- **`converters/boleta_factura.go`**: Main UBL conversion logic
- **`converters/documento.go`**: UBL XML structure definitions and helper functions
- **`signature/signature.go`**: Digital signature implementation using goxmldsig
- **`utils/sunat.go`**: SUNAT webservice communication and CDR processing
- **`validator/validaciones.go`**: Comprehensive business rule validation

## Dependencies

Key external libraries:
- `github.com/beevik/etree`: XML manipulation
- `github.com/russellhaering/goxmldsig`: XML digital signatures
- `software.sslmate.com/src/go-pkcs12`: PKCS#12 certificate handling

## Output Directories

- **`out/`**: Generated XML files and ZIP packages
- **`cdr/`**: CDR responses from SUNAT organized by document

## SUNAT Integration

- **Test Environment**: `https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService`
- **Credentials**: Default test credentials are "MODDATOS"/"MODDATOS"
- **Response Processing**: CDR files are automatically extracted and analyzed for approval/rejection status

## Validation Rules

The validator enforces SUNAT-specific business rules:
- Document series format (F001 for invoices, B001 for receipts)
- Tax calculation consistency
- Customer document type validation
- Date format and logical consistency
- Monetary total calculations

## Development Notes

- All XML output uses UBL 2.1 standard with SUNAT extensions
- Digital signatures follow XMLDSig specifications
- Error handling returns structured JSON responses
- The application creates necessary output directories automatically
- CDR processing extracts response codes to determine document status