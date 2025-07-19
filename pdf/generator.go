package pdf

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"ubl-go-conversor/models"
)

// GeneratePDF genera un PDF de representación impresa de la factura/boleta
func GeneratePDF(documento models.ComprobanteBase, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Título del documento
	tipoDoc := "FACTURA ELECTRÓNICA"
	if documento.TipoDocumento == "03" {
		tipoDoc = "BOLETA DE VENTA ELECTRÓNICA"
	}

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tipoDoc)
	pdf.Ln(15)

	// Información del emisor
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "DATOS DEL EMISOR")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("RUC: %s", documento.Emisor.RUC))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Razón Social: %s", documento.Emisor.RazonSocial))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Dirección: %s", documento.Emisor.Direccion))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Distrito: %s - Provincia: %s - Departamento: %s", 
		documento.Emisor.Distrito, documento.Emisor.Provincia, documento.Emisor.Departamento))
	pdf.Ln(12)

	// Información del cliente
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "DATOS DEL CLIENTE")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	tipoDocCliente := "DNI"
	if documento.Cliente.TipoDoc == "6" {
		tipoDocCliente = "RUC"
	}
	pdf.Cell(0, 6, fmt.Sprintf("%s: %s", tipoDocCliente, documento.Cliente.NumeroDoc))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Razón Social: %s", documento.Cliente.RazonSocial))
	pdf.Ln(6)
	if documento.Cliente.Direccion != "" {
		pdf.Cell(0, 6, fmt.Sprintf("Dirección: %s", documento.Cliente.Direccion))
		pdf.Ln(6)
	}
	pdf.Ln(12)

	// Información del comprobante
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "INFORMACIÓN DEL COMPROBANTE")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Serie y Número: %s-%s", documento.Serie, documento.Numero))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Fecha de Emisión: %s", documento.FechaEmision))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Hora de Emisión: %s", documento.HoraEmision))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Moneda: %s", documento.Moneda))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Forma de Pago: %s", documento.FormaPago))
	pdf.Ln(12)

	// Detalle de items
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "DETALLE DE PRODUCTOS/SERVICIOS")
	pdf.Ln(10)

	// Headers de la tabla
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(15, 8, "Item")
	pdf.Cell(50, 8, "Descripción")
	pdf.Cell(20, 8, "Cantidad")
	pdf.Cell(25, 8, "V. Unitario")
	pdf.Cell(25, 8, "V. Total")
	pdf.Cell(20, 8, "IGV")
	pdf.Cell(25, 8, "P. Unitario")
	pdf.Ln(8)

	// Línea divisoria
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(2)

	// Items
	pdf.SetFont("Arial", "", 8)
	for i, item := range documento.Items {
		pdf.Cell(15, 6, fmt.Sprintf("%d", i+1))
		pdf.Cell(50, 6, truncateString(item.Descripcion, 30))
		pdf.Cell(20, 6, fmt.Sprintf("%.2f", item.Cantidad))
		pdf.Cell(25, 6, fmt.Sprintf("%.2f", item.ValorUnitario))
		pdf.Cell(25, 6, fmt.Sprintf("%.2f", item.ValorTotal))
		pdf.Cell(20, 6, fmt.Sprintf("%.2f", item.IGV))
		pdf.Cell(25, 6, fmt.Sprintf("%.2f", item.PrecioVentaUnitario))
		pdf.Ln(6)
	}

	pdf.Ln(8)

	// Totales
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(130, 6, "")
	pdf.Cell(30, 6, "Sub Total:")
	pdf.Cell(30, 6, fmt.Sprintf("%.2f", documento.TotalGravado))
	pdf.Ln(6)
	
	pdf.Cell(130, 6, "")
	pdf.Cell(30, 6, "IGV (18%):")
	pdf.Cell(30, 6, fmt.Sprintf("%.2f", documento.TotalIGV))
	pdf.Ln(6)
	
	pdf.Cell(130, 6, "")
	pdf.Cell(30, 6, "TOTAL:")
	pdf.Cell(30, 6, fmt.Sprintf("%.2f", documento.TotalImportePagar))
	pdf.Ln(12)

	// Leyendas
	if len(documento.Leyendas) > 0 {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, "OBSERVACIONES:")
		pdf.Ln(8)
		
		pdf.SetFont("Arial", "", 9)
		for _, leyenda := range documento.Leyendas {
			pdf.Cell(0, 6, leyenda.Descripcion)
			pdf.Ln(6)
		}
		pdf.Ln(8)
	}

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 6, fmt.Sprintf("Documento generado el %s", time.Now().Format("02/01/2006 15:04:05")))
	pdf.Ln(4)
	pdf.Cell(0, 6, "Representación impresa de comprobante electrónico")

	return pdf.OutputFileAndClose(outputPath)
}

// GeneratePDFPath genera la ruta donde se guardará el PDF
func GeneratePDFPath(documento models.ComprobanteBase) string {
	return fmt.Sprintf("out/%s-%s-%s-%s.pdf", 
		documento.Emisor.RUC, 
		documento.TipoDocumento, 
		documento.Serie, 
		documento.Numero)
}

// truncateString trunca un string si es muy largo
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}