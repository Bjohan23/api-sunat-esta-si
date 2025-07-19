/*
Conversor de Facturas y Boletas a XML UBL 2.1 para SUNAT
=======================================================

Este paquete es el núcleo de la generación de XML siguiendo el estándar UBL 2.1
con las extensiones específicas de SUNAT para facturación electrónica en Perú.

Funciones principales:
1. ConvertirFacturaAUBL() - Transforma ComprobanteBase a estructura UBL
2. GenerarXMLBF() - Serializa la estructura UBL a archivo XML válido
3. Funciones de mapeo para cada sección del documento UBL

Cumple con:
- UBL 2.1 (Universal Business Language)
- Extensiones SUNAT (UBLExtensions)
- Catálogos oficiales SUNAT
- Validaciones de estructura XML

El XML generado es válido para envío directo a SUNAT después de la firma digital.
*/
package converters

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"

	"ubl-go-conversor/models"
)

/*
Invoice representa la estructura raíz del documento UBL 2.1 para facturas y boletas.
Esta estructura mapea directamente al XML que se envía a SUNAT.

Componentes principales:
- Namespaces: Definen los esquemas XML utilizados (UBL, SUNAT, XMLDSig)
- UBLExtensions: Extensiones específicas de SUNAT (firma digital, percepciones)
- Datos básicos: ID, fechas, tipo de documento, moneda
- Partes: Emisor (AccountingSupplierParty) y Cliente (AccountingCustomerParty)
- Totales: Impuestos (TaxTotal) y montos finales (LegalMonetaryTotal)
- Líneas: Detalle de productos/servicios (InvoiceLines)

Todos los elementos siguen la nomenclatura UBL estándar con prefijos:
- cbc: CommonBasicComponents (elementos simples)
- cac: CommonAggregateComponents (elementos complejos)
- ext: ExtensionComponents (extensiones)
*/
type Invoice struct {
	XMLName                 xml.Name                `xml:"Invoice"`
	// ==================== NAMESPACES XML ====================
	XmlnsXsi                string                  `xml:"xmlns:xsi,attr"`    // XML Schema Instance
	XmlnsXsd                string                  `xml:"xmlns:xsd,attr"`    // XML Schema Definition
	XmlnsCac                string                  `xml:"xmlns:cac,attr"`    // Common Aggregate Components
	XmlnsCbc                string                  `xml:"xmlns:cbc,attr"`    // Common Basic Components
	XmlnsCcts               string                  `xml:"xmlns:ccts,attr"`   // Core Component Technical Specification
	XmlnsDs                 string                  `xml:"xmlns:ds,attr"`     // XML Digital Signature
	XmlnsExt                string                  `xml:"xmlns:ext,attr"`    // Extension Components
	XmlnsQdt                string                  `xml:"xmlns:qdt,attr"`    // Qualified Data Types
	XmlnsUdt                string                  `xml:"xmlns:udt,attr"`    // Unqualified Data Types
	XmlnsSac                string                  `xml:"xmlns:sac,attr"`    // SUNAT Aggregate Components
	Xmlns                   string                  `xml:"xmlns,attr"`        // Default namespace
	
	// ==================== EXTENSIONES SUNAT ====================
	UBLExtensions           UBLExtensions           `xml:"ext:UBLExtensions"` // Contenedor para firma digital y percepciones
	
	// ==================== INFORMACIÓN BÁSICA DEL DOCUMENTO ====================
	UBLVersionID            string                  `xml:"cbc:UBLVersionID"`    // Versión UBL (2.1)
	CustomizationID         CustomizationID         `xml:"cbc:CustomizationID"` // Versión de implementación SUNAT
	ProfileID               ProfileID               `xml:"cbc:ProfileID"`       // Tipo de operación (catálogo 51)
	ID                      string                  `xml:"cbc:ID"`              // Serie-Número del comprobante
	IssueDate               string                  `xml:"cbc:IssueDate"`       // Fecha de emisión (YYYY-MM-DD)
	IssueTime               string                  `xml:"cbc:IssueTime"`       // Hora de emisión (HH:MM:SS)
	DueDate                 string                  `xml:"cbc:DueDate,omitempty"` // Fecha de vencimiento (opcional)
	InvoiceTypeCode         InvoiceTypeCode         `xml:"cbc:InvoiceTypeCode"` // Tipo de documento (01=Factura, 03=Boleta)
	Notes                   []Note                  `xml:"cbc:Note,omitempty"`  // Leyendas (importes en letras, etc.)
	DocumentCurrencyCode    DocumentCurrencyCode    `xml:"cbc:DocumentCurrencyCode"` // Moneda (PEN, USD, EUR)
	LineCountNumeric        int                     `xml:"cbc:LineCountNumeric"`     // Cantidad de líneas de detalle
	
	// ==================== FIRMA DIGITAL ====================
	Signature               Signature               `xml:"cac:Signature"`       // Información del certificado digital
	
	// ==================== PARTES DEL DOCUMENTO ====================
	AccountingSupplierParty AccountingSupplierParty `xml:"cac:AccountingSupplierParty"` // Datos del emisor
	AccountingCustomerParty AccountingCustomerParty `xml:"cac:AccountingCustomerParty"` // Datos del cliente
	
	// ==================== CONDICIONES DE PAGO ====================
	PaymentTerms            []PaymentTerms          `xml:"cac:PaymentTerms,omitempty"` // Forma de pago y cuotas
	
	// ==================== TOTALES E IMPUESTOS ====================
	TaxTotal                []TaxTotal              `xml:"cac:TaxTotal"`       // Resumen de impuestos (IGV)
	LegalMonetaryTotal      LegalMonetaryTotal      `xml:"cac:LegalMonetaryTotal"` // Totales monetarios finales
	
	// ==================== DETALLE DE LÍNEAS ====================
	InvoiceLines            []InvoiceLine           `xml:"cac:InvoiceLine"`    // Productos/servicios vendidos
}

type Note struct {
	Value            string `xml:",chardata"`
	LanguageLocaleID string `xml:"languageLocaleID,attr"`
}

type InvoiceLine struct {
	ID                  string             `xml:"cbc:ID"`
	InvoicedQuantity    InvoicedQuantity   `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount AmountWithCurrency `xml:"cbc:LineExtensionAmount"`
	PricingReference    PricingReference   `xml:"cac:PricingReference"`
	TaxTotal            TaxTotal           `xml:"cac:TaxTotal"`
	Item                Item               `xml:"cac:Item"`
	Price               Price              `xml:"cac:Price"`
}

type InvoicedQuantity struct {
	Value                  float64 `xml:",chardata"`
	UnitCode               string  `xml:"unitCode,attr"`
	UnitCodeListID         string  `xml:"unitCodeListID,attr"`
	UnitCodeListAgencyName string  `xml:"unitCodeListAgencyName,attr"`
}

type PricingReference struct {
	AlternativeConditionPrice AlternativeConditionPrice `xml:"cac:AlternativeConditionPrice"`
}

type AlternativeConditionPrice struct {
	PriceAmount   AmountWithCurrency `xml:"cbc:PriceAmount"`
	PriceTypeCode PriceTypeCode      `xml:"cbc:PriceTypeCode"`
}

type PriceTypeCode struct {
	Value          string `xml:",chardata"`
	ListName       string `xml:"listName,attr"`
	ListAgencyName string `xml:"listAgencyName,attr"`
	ListURI        string `xml:"listURI,attr"`
}

type Item struct {
	Description               CDATAString               `xml:"cbc:Description"`
	SellersItemIdentification SellersItemIdentification `xml:"cac:SellersItemIdentification"`
	CommodityClassification   CommodityClassification   `xml:"cac:CommodityClassification"`
}

type SellersItemIdentification struct {
	ID CDATAString `xml:"cbc:ID"`
}

type CommodityClassification struct {
	ItemClassificationCode ItemClassificationCode `xml:"cbc:ItemClassificationCode"`
}

type ItemClassificationCode struct {
	Value          string `xml:",chardata"`
	ListID         string `xml:"listID,attr"`
	ListAgencyName string `xml:"listAgencyName,attr"`
	ListName       string `xml:"listName,attr"`
}

type Price struct {
	PriceAmount AmountWithCurrency `xml:"cbc:PriceAmount"`
}

/*
============================ FUNCIÓN PRINCIPAL DE CONVERSIÓN ============================

ConvertirFacturaAUBL transforma un ComprobanteBase (estructura interna) a Invoice (estructura UBL).
Esta es la función núcleo que mapea todos los datos del comprobante peruano al estándar internacional UBL 2.1.

Proceso de conversión:
1. Configurar parámetros base (profileID, extensiones)
2. Mapear leyendas del comprobante a elementos Note
3. Crear extensiones UBL requeridas por SUNAT
4. Construir estructura Invoice completa con todos sus componentes
5. Aplicar namespaces y versiones UBL correctas

El resultado es una estructura Invoice lista para serializar a XML válido.
*/
func ConvertirFacturaAUBL(f models.ComprobanteBase) Invoice {
	// Tipo de operación según catálogo 51 de SUNAT
	// 0101 = Venta interna (operación más común)
	profileID := "0101"
	
	// Convertir leyendas del comprobante (ej: importe en letras) a elementos UBL Note
	notes := []Note{}
	for _, leyenda := range f.Leyendas {
		notes = append(notes, Note{
			Value:            leyenda.Descripcion, // Texto de la leyenda
			LanguageLocaleID: leyenda.Codigo,      // Código de tipo de leyenda (catálogo 52)
		})
	}

	// ==================== EXTENSIONES UBL PARA SUNAT ====================
	var extensiones []UBLExtension

	// 1. Extensión vacía obligatoria para firma digital
	//    SUNAT requiere esta extensión para insertar la firma XMLDSig
	extensiones = append(extensiones, UBLExtension{
		ExtensionContent: ExtensionContent{}, // Contenido vacío, se llena al firmar
	})

	// 2. Si hay percepción, se agrega como otra extensión
	if percepcion := crearPercepcion(f); percepcion != nil {
		extensiones = append(extensiones, *percepcion)
	}

	invoice := Invoice{
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsCac:  "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCbc:  "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		XmlnsCcts: "urn:un:unece:uncefact:documentation:2",
		XmlnsDs:   "http://www.w3.org/2000/09/xmldsig#",
		XmlnsExt:  "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",
		XmlnsQdt:  "urn:oasis:names:specification:ubl:schema:xsd:QualifiedDatatypes-2",
		XmlnsUdt:  "urn:un:unece:uncefact:data:specification:UnqualifiedDataTypesSchemaModule:2",
		XmlnsSac:  "urn:sunat:names:specification:ubl:peru:schema:xsd:SunatAggregateComponents-1",
		Xmlns:     "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",

		UBLExtensions: UBLExtensions{
			UBLExtension: extensiones,
		},
		UBLVersionID: "2.1",
		CustomizationID: CustomizationID{
			Value:            "2.0",
			SchemeAgencyName: "PE:SUNAT",
		},
		ProfileID: ProfileID{
			Value:            profileID,
			SchemeName:       "Tipo de Operacion",
			SchemeAgencyName: "PE:SUNAT",
			SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo51",
		},
		ID:                      f.Serie + "-" + f.Numero,
		IssueDate:               f.FechaEmision,
		IssueTime:               f.HoraEmision,
		DueDate:                 f.FechaVencimiento,
		InvoiceTypeCode:         crearInvoiceTypeCode(f),
		DocumentCurrencyCode:    crearCurrencyCode(f.Moneda),
		LineCountNumeric:        len(f.Items),
		Signature:               crearFirma(f),
		AccountingSupplierParty: crearEmisor(f.Emisor),
		AccountingCustomerParty: crearCliente(f.Cliente),
		PaymentTerms:            crearPaymentTerms(f),
		TaxTotal:                crearTaxTotals(f),
		LegalMonetaryTotal:      crearTotalesMonetarios(f),
		InvoiceLines:            crearLineas(f.Items, f.Moneda),
		Notes:                   notes,
	}

	return invoice
}

func crearInvoiceTypeCode(f models.ComprobanteBase) InvoiceTypeCode {
	return InvoiceTypeCode{
		Value:          f.TipoDocumento,
		ListAgencyName: "PE:SUNAT",
		ListName:       "Tipo de Documento",
		ListURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo01",
		ListID:         "0101",
	}
}

func crearCurrencyCode(moneda string) DocumentCurrencyCode {
	return DocumentCurrencyCode{
		Value:          moneda,
		ListID:         "ISO 4217 Alpha",
		ListName:       "Currency",
		ListAgencyName: "United Nations Economic Commission for Europe",
	}
}

func GenerarXMLBF(f models.ComprobanteBase, rutaArchivo string) error {
	invoice := ConvertirFacturaAUBL(f)
	xmlData, err := xml.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar XML: %v", err)
	}
	xmlString := xml.Header + string(xmlData)
	xmlString = limpiarXML(xmlString)
	return os.WriteFile(rutaArchivo, []byte(xmlString), 0644)
}

func limpiarXML(xmlStr string) string {
	reAttrs := regexp.MustCompile(`\s+\w+(?::\w+)?=""`)
	xmlStr = reAttrs.ReplaceAllString(xmlStr, "")
	reEmptySelfClosing := regexp.MustCompile(`<\w+(:\w+)?[^>]*/>`)
	xmlStr = reEmptySelfClosing.ReplaceAllString(xmlStr, "")
	return xmlStr
}
