// Archivo: converters/ubl_factura_boleta.go
package converters

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"

	"ubl-go-conversor/models"
)

type Invoice struct {
	XMLName                 xml.Name                `xml:"Invoice"`
	XmlnsXsi                string                  `xml:"xmlns:xsi,attr"`
	XmlnsXsd                string                  `xml:"xmlns:xsd,attr"`
	XmlnsCac                string                  `xml:"xmlns:cac,attr"`
	XmlnsCbc                string                  `xml:"xmlns:cbc,attr"`
	XmlnsCcts               string                  `xml:"xmlns:ccts,attr"`
	XmlnsDs                 string                  `xml:"xmlns:ds,attr"`
	XmlnsExt                string                  `xml:"xmlns:ext,attr"`
	XmlnsQdt                string                  `xml:"xmlns:qdt,attr"`
	XmlnsUdt                string                  `xml:"xmlns:udt,attr"`
	XmlnsSac                string                  `xml:"xmlns:sac,attr"`
	Xmlns                   string                  `xml:"xmlns,attr"`
	UBLExtensions           UBLExtensions           `xml:"ext:UBLExtensions"`
	UBLVersionID            string                  `xml:"cbc:UBLVersionID"`
	CustomizationID         CustomizationID         `xml:"cbc:CustomizationID"`
	ProfileID               ProfileID               `xml:"cbc:ProfileID"`
	ID                      string                  `xml:"cbc:ID"`
	IssueDate               string                  `xml:"cbc:IssueDate"`
	IssueTime               string                  `xml:"cbc:IssueTime"`
	DueDate                 string                  `xml:"cbc:DueDate,omitempty"`
	InvoiceTypeCode         InvoiceTypeCode         `xml:"cbc:InvoiceTypeCode"`
	Notes                   []Note                  `xml:"cbc:Note,omitempty"`
	DocumentCurrencyCode    DocumentCurrencyCode    `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric        int                     `xml:"cbc:LineCountNumeric"`
	Signature               Signature               `xml:"cac:Signature"`
	AccountingSupplierParty AccountingSupplierParty `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty AccountingCustomerParty `xml:"cac:AccountingCustomerParty"`
	PaymentTerms            []PaymentTerms          `xml:"cac:PaymentTerms,omitempty"`
	TaxTotal                []TaxTotal              `xml:"cac:TaxTotal"`
	LegalMonetaryTotal      LegalMonetaryTotal      `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines            []InvoiceLine           `xml:"cac:InvoiceLine"`
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

// ============================ FUNCIÓN PRINCIPAL ============================
func ConvertirFacturaAUBL(f models.ComprobanteBase) Invoice {
	profileID := "0101"
	notes := []Note{}
	for _, leyenda := range f.Leyendas {
		notes = append(notes, Note{
			Value:            leyenda.Descripcion,
			LanguageLocaleID: leyenda.Codigo,
		})
	}

	// Crear todas las extensiones necesarias
	var extensiones []UBLExtension

	// 1. Extension vacía para firma digital (requerida por SUNAT)
	extensiones = append(extensiones, UBLExtension{
		ExtensionContent: ExtensionContent{},
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
