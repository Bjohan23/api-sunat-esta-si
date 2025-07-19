// Archivo: converters/documento.go
package converters

import (
	"encoding/xml"
	"math"
	"strconv"
	"strings"
	"ubl-go-conversor/models"
)

// Estructura base del documento UBL

type UBLExtensions struct {
	UBLExtension []UBLExtension `xml:"ext:UBLExtension"`
}

type UBLExtension struct {
	ExtensionContent ExtensionContent `xml:"ext:ExtensionContent"`
}
type ExtensionContent struct {
	SUNATPerception *SUNATPerception `xml:"sac:SUNATPerception,omitempty"`
}
type CustomizationID struct {
	Value            string `xml:",chardata"`
	SchemeAgencyName string `xml:"schemeAgencyName,attr"`
}

type ProfileID struct {
	Value            string `xml:",chardata"`
	SchemeName       string `xml:"schemeName,attr"`
	SchemeAgencyName string `xml:"schemeAgencyName,attr"`
	SchemeURI        string `xml:"schemeURI,attr"`
}

type InvoiceTypeCode struct {
	Value          string `xml:",chardata"`
	ListAgencyName string `xml:"listAgencyName,attr"`
	ListName       string `xml:"listName,attr"`
	ListURI        string `xml:"listURI,attr"`
	ListID         string `xml:"listID,attr"`
}

type DocumentCurrencyCode struct {
	Value          string `xml:",chardata"`
	ListID         string `xml:"listID,attr"`
	ListName       string `xml:"listName,attr"`
	ListAgencyName string `xml:"listAgencyName,attr"`
}

// ========================== ESTRUCTURA PARA FIRMA DIGITAL ==========================
type Signature struct {
	ID                         string                     `xml:"cbc:ID"`
	SignatoryParty             SignatoryParty             `xml:"cac:SignatoryParty"`
	DigitalSignatureAttachment DigitalSignatureAttachment `xml:"cac:DigitalSignatureAttachment"`
}

type SignatoryParty struct {
	PartyIdentification PartyIdentification `xml:"cac:PartyIdentification"`
	PartyName           PartyName           `xml:"cac:PartyName"`
}

type DigitalSignatureAttachment struct {
	ExternalReference ExternalReference `xml:"cac:ExternalReference"`
}

type ExternalReference struct {
	URI string `xml:"cbc:URI"`
}

// ========================== ESTRUCTURA PARA DATOS DEL EMISOR Y CLIENTE ==========================
type AccountingSupplierParty struct {
	Party Party `xml:"cac:Party"`
}

type AccountingCustomerParty struct {
	Party Party `xml:"cac:Party"`
}

type Party struct {
	PartyIdentification PartyIdentification `xml:"cac:PartyIdentification"`
	PartyName           PartyName           `xml:"cac:PartyName"`
	PartyTaxScheme      PartyTaxScheme      `xml:"cac:PartyTaxScheme"`
	PartyLegalEntity    PartyLegalEntity    `xml:"cac:PartyLegalEntity"`
	Contact             Contact             `xml:"cac:Contact,omitempty"`
}

type PartyIdentification struct {
	ID IDWithScheme `xml:"cbc:ID"`
}

type IDWithScheme struct {
	Value            string `xml:",chardata"`
	SchemeID         string `xml:"schemeID,attr,omitempty"`
	SchemeName       string `xml:"schemeName,attr,omitempty"`
	SchemeAgencyName string `xml:"schemeAgencyName,attr,omitempty"`
	SchemeURI        string `xml:"schemeURI,attr,omitempty"`
}

type PartyName struct {
	Name CDATAString `xml:"cbc:Name"`
}

type CDATAString struct {
	Value string `xml:",cdata"`
}

type PartyTaxScheme struct {
	RegistrationName CDATAString     `xml:"cbc:RegistrationName"`
	CompanyID        IDWithScheme    `xml:"cbc:CompanyID"`
	TaxScheme        TaxSchemeSimple `xml:"cac:TaxScheme"`
}

type TaxSchemeSimple struct {
	ID IDWithScheme `xml:"cbc:ID"`
}

type PartyLegalEntity struct {
	RegistrationName    CDATAString         `xml:"cbc:RegistrationName"`
	RegistrationAddress RegistrationAddress `xml:"cac:RegistrationAddress,omitempty"`
}

type RegistrationAddress struct {
	ID               AddressID       `xml:"cbc:ID"`
	AddressTypeCode  AddressTypeCode `xml:"cbc:AddressTypeCode"`
	CityName         CDATAString     `xml:"cbc:CityName"`
	CountrySubentity CDATAString     `xml:"cbc:CountrySubentity"`
	District         CDATAString     `xml:"cbc:District"`
	AddressLine      AddressLine     `xml:"cac:AddressLine"`
	Country          Country         `xml:"cac:Country"`
}

type AddressID struct {
	Value            string `xml:",chardata"`
	SchemeName       string `xml:"schemeName,attr"`
	SchemeAgencyName string `xml:"schemeAgencyName,attr"`
}

type AddressTypeCode struct {
	Value          string `xml:",chardata"`
	ListAgencyName string `xml:"listAgencyName,attr"`
	ListName       string `xml:"listName,attr"`
}

type AddressLine struct {
	Line CDATAString `xml:"cbc:Line"`
}

type Country struct {
	IdentificationCode CountryCode `xml:"cbc:IdentificationCode"`
}

type CountryCode struct {
	Value          string `xml:",chardata"`
	ListID         string `xml:"listID,attr"`
	ListAgencyName string `xml:"listAgencyName,attr"`
	ListName       string `xml:"listName,attr"`
}

type Contact struct {
	Name CDATAString `xml:"cbc:Name"`
}

// Impuestos y montos
type PaymentTerms struct {
	ID             string              `xml:"cbc:ID"`                       // 1
	PaymentMeansID string              `xml:"cbc:PaymentMeansID,omitempty"` // 2
	Amount         *AmountWithCurrency `xml:"cbc:Amount,omitempty"`         // 3
	PaymentDueDate string              `xml:"cbc:PaymentDueDate,omitempty"` // 4
}

type TaxTotal struct {
	TaxAmount   AmountWithCurrency `xml:"cbc:TaxAmount"`
	TaxSubtotal []TaxSubtotal        `xml:"cac:TaxSubtotal"`
}

type TaxSubtotal struct {
	TaxableAmount AmountWithCurrency `xml:"cbc:TaxableAmount"`
	TaxAmount     AmountWithCurrency `xml:"cbc:TaxAmount"`
	TaxCategory   TaxCategory        `xml:"cac:TaxCategory"`
}

type TaxCategory struct {
	ID                     TaxCategoryID          `xml:"cbc:ID"`
	Percent                *float64               `xml:"cbc:Percent"`
	TaxExemptionReasonCode TaxExemptionReasonCode `xml:"cbc:TaxExemptionReasonCode,omitempty"`
	TaxScheme              TaxScheme              `xml:"cac:TaxScheme"`
}

type TaxCategoryID struct {
	Value            string `xml:",chardata"`
	SchemeID         string `xml:"schemeID,attr"`
	SchemeName       string `xml:"schemeName,attr"`
	SchemeAgencyName string `xml:"schemeAgencyName,attr"`
}

type TaxExemptionReasonCode struct {
	Value          string `xml:",chardata"`
	ListAgencyName string `xml:"listAgencyName,attr"`
	ListName       string `xml:"listName,attr"`
	ListURI        string `xml:"listURI,attr"`
}

type TaxScheme struct {
	ID          TaxSchemeID `xml:"cbc:ID"`
	Name        string      `xml:"cbc:Name"`
	TaxTypeCode string      `xml:"cbc:TaxTypeCode"`
}

type TaxSchemeID struct {
	Value            string `xml:",chardata"`
	SchemeID         string `xml:"schemeID,attr"`
	SchemeAgencyID   string `xml:"schemeAgencyID,attr,omitempty"`
	SchemeName       string `xml:"schemeName,attr,omitempty"`
	SchemeAgencyName string `xml:"schemeAgencyName,attr,omitempty"`
}

type LegalMonetaryTotal struct {
	LineExtensionAmount AmountWithCurrency `xml:"cbc:LineExtensionAmount"`
	TaxInclusiveAmount  AmountWithCurrency `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount       AmountWithCurrency `xml:"cbc:PayableAmount"`
}

type AmountWithCurrency struct {
	Value      float64 `xml:",chardata"`
	CurrencyID string  `xml:"currencyID,attr"`
}

type SUNATPerception struct {
	XMLName            xml.Name           `xml:"sac:SUNATPerception"`
	SystemCode         string             `xml:"sac:SUNATPerceptionSystemCode"`
	Percent            float64            `xml:"sac:SUNATPerceptionPercent"`
	TotalInvoiceAmount AmountWithCurrency `xml:"sac:TotalInvoiceAmount"`
	PerceptionAmount   AmountWithCurrency `xml:"sac:SUNATPerceptionAmount"`
	PerceptionDate     string             `xml:"sac:SUNATPerceptionDate"`
	NetTotalPaid       AmountWithCurrency `xml:"sac:SUNATNetTotalCashed"`
}



// Estructura para la firma digital
func crearFirma(f models.ComprobanteBase) Signature {
	return Signature{
		ID: f.Serie + "-" + f.Numero,
		SignatoryParty: SignatoryParty{
			PartyIdentification: PartyIdentification{
				ID: IDWithScheme{
					Value: f.Emisor.RUC,
				},
			},
			PartyName: PartyName{
				Name: CDATAString{Value: f.Emisor.RazonSocial},
			},
		},
		DigitalSignatureAttachment: DigitalSignatureAttachment{
			ExternalReference: ExternalReference{
				URI: "#SignatureSP",
			},
		},
	}
}

// crear Emisor (Quien emite el comprobante)
func crearEmisor(emisor models.Emisor) AccountingSupplierParty {
	return AccountingSupplierParty{
		Party: Party{
			PartyIdentification: PartyIdentification{
				ID: IDWithScheme{
					Value:            emisor.RUC,
					SchemeID:         "6",
					SchemeName:       "Documento de Identidad",
					SchemeAgencyName: "PE:SUNAT",
					SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
				},
			},
			PartyName: PartyName{
				Name: CDATAString{Value: emisor.RazonSocial},
			},
			PartyTaxScheme: PartyTaxScheme{
				RegistrationName: CDATAString{Value: emisor.RazonSocial},
				CompanyID: IDWithScheme{
					Value:            emisor.RUC,
					SchemeID:         "6",
					SchemeName:       "SUNAT:Identificador de Documento de Identidad",
					SchemeAgencyName: "PE:SUNAT",
					SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
				},
				TaxScheme: TaxSchemeSimple{
					ID: IDWithScheme{
						Value:            emisor.RUC,
						SchemeID:         "6",
						SchemeName:       "SUNAT:Identificador de Documento de Identidad",
						SchemeAgencyName: "PE:SUNAT",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
					},
				},
			},
			PartyLegalEntity: PartyLegalEntity{
				RegistrationName: CDATAString{Value: emisor.RazonSocial},
				RegistrationAddress: RegistrationAddress{
					ID: AddressID{
						Value:            emisor.Ubigeo,
						SchemeName:       "Ubigeos",
						SchemeAgencyName: "PE:INEI",
					},
					AddressTypeCode: AddressTypeCode{
						Value:          "0000",
						ListAgencyName: "PE:SUNAT",
						ListName:       "Establecimientos anexos",
					},
					CityName:         CDATAString{Value: emisor.Provincia},
					CountrySubentity: CDATAString{Value: emisor.Departamento},
					District:         CDATAString{Value: emisor.Distrito},
					AddressLine: AddressLine{
						Line: CDATAString{Value: emisor.Direccion},
					},
					Country: Country{
						IdentificationCode: CountryCode{
							Value:          emisor.CodigoPais,
							ListID:         "ISO 3166-1",
							ListAgencyName: "United Nations Economic Commission for Europe",
							ListName:       "Country",
						},
					},
				},
			},
			Contact: Contact{
				Name: CDATAString{Value: ""},
			},
		},
	}
}

// crear cliente (Quien recibe el comprobante)
func crearCliente(cliente models.Cliente) AccountingCustomerParty {
	return AccountingCustomerParty{
		Party: Party{
			PartyIdentification: PartyIdentification{
				ID: IDWithScheme{
					Value:            cliente.NumeroDoc,
					SchemeID:         cliente.TipoDoc,
					SchemeName:       "Documento de Identidad",
					SchemeAgencyName: "PE:SUNAT",
					SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
				},
			},
			PartyName: PartyName{
				Name: CDATAString{Value: cliente.RazonSocial},
			},
			PartyTaxScheme: PartyTaxScheme{
				RegistrationName: CDATAString{Value: cliente.RazonSocial},
				CompanyID: IDWithScheme{
					Value:            cliente.NumeroDoc,
					SchemeID:         cliente.TipoDoc,
					SchemeName:       "SUNAT:Identificador de Documento de Identidad",
					SchemeAgencyName: "PE:SUNAT",
					SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
				},
				TaxScheme: TaxSchemeSimple{
					ID: IDWithScheme{
						Value:            cliente.NumeroDoc,
						SchemeID:         cliente.TipoDoc,
						SchemeName:       "SUNAT:Identificador de Documento de Identidad",
						SchemeAgencyName: "PE:SUNAT",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
					},
				},
			},
			PartyLegalEntity: PartyLegalEntity{
				RegistrationName: CDATAString{Value: cliente.RazonSocial},
				RegistrationAddress: RegistrationAddress{
					ID: AddressID{
						Value:            cliente.Ubigeo,
						SchemeName:       "Ubigeos",
						SchemeAgencyName: "PE:INEI",
					},
					AddressTypeCode: AddressTypeCode{
						Value:          "0000",
						ListAgencyName: "PE:SUNAT",
						ListName:       "Establecimientos anexos",
					},
					CityName:         CDATAString{Value: cliente.Provincia},
					CountrySubentity: CDATAString{Value: cliente.Departamento},
					District:         CDATAString{Value: cliente.Distrito},
					AddressLine: AddressLine{
						Line: CDATAString{Value: cliente.Direccion},
					},
					Country: Country{
						IdentificationCode: CountryCode{
							Value:          cliente.CodigoPais,
							ListID:         "ISO 3166-1",
							ListAgencyName: "United Nations Economic Commission for Europe",
							ListName:       "Country",
						},
					},
				},
			},
		},
	}
}

// crea los totales de impuestos
func crearTaxTotals(f models.ComprobanteBase) []TaxTotal {
	// Acumulador de subtotales por tipo de afectación IGV
	var percent float64

	subtotales := map[string]struct {
		Base, IGV, Porcentaje float64
	}{}

	for _, item := range f.Items {

		switch item.TipoAfectacionIGV {
		case "10", "11", "12", "13", "14", "15", "16", "17":
			percent = 18.00
		default:
			percent = 0.00
		}

		s := subtotales[item.TipoAfectacionIGV]
		s.Base += item.ValorTotal
		s.IGV += item.IGV
		s.Porcentaje = percent
		subtotales[item.TipoAfectacionIGV] = s
	}

	var taxSubtotals []TaxSubtotal
	var totalIGV float64

	for tipo, s := range subtotales {
		item := models.ItemComprobante{
			TipoAfectacionIGV: tipo,
		}
		taxSubtotals = append(taxSubtotals, TaxSubtotal{
			TaxableAmount: newAmount(s.Base, f.Moneda),
			TaxAmount:     newAmount(s.IGV, f.Moneda),
			TaxCategory:   newTaxCategory(item),
		})
		totalIGV += s.IGV
	}

	return []TaxTotal{{
		TaxAmount:    newAmount(totalIGV, f.Moneda),
		TaxSubtotal: taxSubtotals,
	}}
}

// crea los totales monetarios
func crearTotalesMonetarios(f models.ComprobanteBase) LegalMonetaryTotal {
	// Calcular correctamente LineExtensionAmount según el tipo de afectación
	var lineExtensionAmount float64

	// Sumar valores según el tipo de afectación
	for _, item := range f.Items {
		switch item.TipoAfectacionIGV {
		case "10", "11", "12", "13", "14", "15", "16", "17": // Gravado
			lineExtensionAmount += item.ValorTotal
		case "20": // Exonerado
			lineExtensionAmount += item.ValorTotal
		case "21": // Gratuito
			// No se suma a LineExtensionAmount
		case "30", "31", "32", "33", "34", "35", "36", "37": // Inafecto
			lineExtensionAmount += item.ValorTotal
		case "40": // Exportación
			lineExtensionAmount += item.ValorTotal
		}
	}

	return LegalMonetaryTotal{
		LineExtensionAmount: AmountWithCurrency{
			Value:      lineExtensionAmount,
			CurrencyID: f.Moneda,
		},
		TaxInclusiveAmount: AmountWithCurrency{
			Value:      f.TotalPrecioVenta,
			CurrencyID: f.Moneda,
		},
		PayableAmount: AmountWithCurrency{
			Value:      f.TotalImportePagar,
			CurrencyID: f.Moneda,
		},
	}
}

// crearLineasFactura convierte los items a líneas UBL
func crearLineas(items []models.ItemComprobante, moneda string) []InvoiceLine {
	var lines []InvoiceLine
	for i, item := range items {
		priceAmount := item.ValorUnitario
		price := item.PrecioVentaUnitario
		if item.TipoAfectacionIGV == "21"  {
			priceAmount = 0.00
			price = item.ValorUnitario
		}

		lines = append(lines, InvoiceLine{
			ID: strconv.Itoa(i + 1),
			InvoicedQuantity: InvoicedQuantity{
				Value:                  item.Cantidad,
				UnitCode:               item.UnidadMedida,
				UnitCodeListID:         "UN/ECE rec 20",
				UnitCodeListAgencyName: "United Nations Economic Commission for Europe",
			},
			LineExtensionAmount: newAmount(item.ValorTotal, moneda),
			PricingReference: PricingReference{
				AlternativeConditionPrice: AlternativeConditionPrice{
					PriceAmount: newAmount(price, moneda),
					PriceTypeCode: PriceTypeCode{
						Value:          item.CodigoTipoPrecio,
						ListName:       "Tipo de Precio",
						ListAgencyName: "PE:SUNAT",
						ListURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo16",
					},
				},
			},
			TaxTotal: TaxTotal{
				TaxAmount: newAmount(item.IGV, moneda),
				TaxSubtotal: []TaxSubtotal{
					{
						TaxableAmount: newAmount(item.ValorTotal, moneda),
						TaxAmount:     newAmount(item.IGV, moneda),
						TaxCategory:   newTaxCategory(item),
					},
				},
			},
			Item: Item{
				Description: CDATAString{Value: item.Descripcion},
				SellersItemIdentification: SellersItemIdentification{
					ID: CDATAString{Value: item.CodigoProducto},
				},
				CommodityClassification: CommodityClassification{
					ItemClassificationCode: ItemClassificationCode{
						Value:          item.UNSPSC,
						ListID:         "UNSPSC",
						ListAgencyName: "GS1 US",
						ListName:       "Item Classification",
					},
				},
			},
			Price: Price{
				PriceAmount: newAmount(priceAmount, moneda),
			},
		})
	}
	return lines
}

// Función para determinar el código de categoría de impuesto según el tipo de afectación
func obtenerCodigoCategoriaTributo(tipoAfectacionIGV string) string {
	switch tipoAfectacionIGV {
	case "10": // Gravado - Operación Onerosa
		return "S"
	case "11": // Gravado - Retiro por premio
		return "S"
	case "12": // Gravado - Retiro por donación
		return "S"
	case "13": // Gravado - Retiro
		return "S"
	case "14": // Gravado - Retiro por publicidad
		return "S"
	case "15": // Gravado - Bonificaciones
		return "S"
	case "16": // Gravado - Retiro por entrega a trabajadores
		return "S"
	case "17": // Gravado - IVAP
		return "S"
	case "20": // Exonerado - Operación Onerosa
		return "E"
	case "21": // Exonerado - Transferencia gratuita
		return "Z"
	case "30": // Inafecto - Operación Onerosa
		return "O"
	case "31": // Inafecto - Retiro por bonificación
		return "O"
	case "32": // Inafecto - Retiro
		return "O"
	case "33": // Inafecto - Retiro por muestras médicas
		return "O"
	case "34": // Inafecto - Retiro por convenio colectivo
		return "O"
	case "35": // Inafecto - Retiro por premio
		return "O"
	case "36": // Inafecto - Retiro por publicidad
		return "O"
	case "37": // Inafecto - Transferencia gratuita
		return "O"
	case "40": // Exportación
		return "G"
	default:
		return "S" // Por defecto gravado
	}
}

// Función para determinar el código de tributo según el tipo de afectación
func obtenerCodigoTributo(tipoAfectacionIGV string) string {
	switch tipoAfectacionIGV {
	case "10", "11", "12", "13", "14", "15", "16", "17": // Gravado
		return "1000"
	case "20": // Exonerado
		return "9997"
	case "21": // Exonerado - Transferencia gratuita
		return "9996"
	case "30", "31", "32", "33", "34", "35", "36", "37": // Inafecto
		return "9998"
	case "40": // Exportación
		return "9995"
	default:
		return "1000" // Por defecto IGV
	}
}

// Función para determinar el nombre del tributo según el tipo de afectación
func obtenerNombreTributo(tipoAfectacionIGV string) string {
	switch tipoAfectacionIGV {
	case "10", "11", "12", "13", "14", "15", "16", "17": // Gravado
		return "IGV"
	case "20": // Exonerado
		return "EXO"
	case "21": // Exonerado - Transferencia gratuita
		return "GRA"
	case "30", "31", "32", "33", "34", "35", "36", "37": // Inafecto
		return "INA"
	case "40": // Exportación
		return "EXP"
	default:
		return "IGV" // Por defecto IGV
	}
}

// Función para determinar el tipo de tributo según el tipo de afectación
func obtenerTipoTributo(tipoAfectacionIGV string) string {
	switch tipoAfectacionIGV {
	case "10", "11", "12", "13", "14", "15", "16", "17": // Gravado
		return "VAT" // Impuesto General a las Ventas
	case "20": // Exonerado
		return "VAT" // Exonerado
	case "21": // Exonerado - Transferencia gratuita
		return "FRE" // Gratuito
	case "30", "31", "32", "33", "34", "35", "36", "37": // Inafecto
		return "INA" // Inafecto
	case "40": // Exportación
		return "FRE" // Gratuito o libre (free)
	default:
		return "" // Vacío si no se sabe el código correcto
	}
}

// Función auxiliar para crear puntero a float64
func floatPtr(f float64) *float64 {
	return &f
}


func crearPaymentTerms(f models.ComprobanteBase) []PaymentTerms {
	terms := []PaymentTerms{
		{
			ID:             "FormaPago",
			PaymentMeansID: f.FormaPago,
			Amount:         floatPtrAmount(f.TotalImportePagar, f.Moneda),
		},
	}

	if strings.EqualFold(f.FormaPago, "Credito") {
		for _, cuota := range f.Cuotas {
			terms = append(terms, PaymentTerms{
				ID:             "FormaPago",
				PaymentMeansID: cuota.NumeroCuota,
				PaymentDueDate: cuota.FechaVencimiento,
				Amount:         floatPtrAmount(cuota.Importe, f.Moneda),
			})
		}
	}

	return terms
}

func floatPtrAmount(val float64, currency string) *AmountWithCurrency {
	return &AmountWithCurrency{Value: val, CurrencyID: currency}
}

// Función para crear un nuevo AmountWithCurrency
func newAmount(value float64, currency string) AmountWithCurrency {
	return AmountWithCurrency{Value: value, CurrencyID: currency}
}

func newTaxCategory(item models.ItemComprobante) TaxCategory {
	var percent float64

	switch item.TipoAfectacionIGV {
	case "10", "11", "12", "13", "14", "15", "16":
		percent = 18.00
	default:
		percent = 0.00
	}
	return TaxCategory{
		ID: TaxCategoryID{
			Value:            obtenerCodigoCategoriaTributo(item.TipoAfectacionIGV),
			SchemeID:         "UN/ECE 5305",
			SchemeName:       "Tax Category Identifier",
			SchemeAgencyName: "United Nations Economic Commission for Europe",
		},
		Percent: floatPtr(percent),
		TaxExemptionReasonCode: TaxExemptionReasonCode{
			Value:          item.TipoAfectacionIGV,
			ListAgencyName: "PE:SUNAT",
			ListName:       "Afectacion del IGV",
			ListURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo07",
		},
		TaxScheme: TaxScheme{
			ID: TaxSchemeID{
				Value:            obtenerCodigoTributo(item.TipoAfectacionIGV),
				SchemeID:         "UN/ECE 5153",
				SchemeAgencyName: "PE:SUNAT",
			},
			Name:        obtenerNombreTributo(item.TipoAfectacionIGV),
			TaxTypeCode: obtenerTipoTributo(item.TipoAfectacionIGV),
		},
	}
}

func crearPercepcion(f models.ComprobanteBase) *UBLExtension {
	if f.TipoDocumento != "01" {
		return nil
	}
	var percent float64
	switch f.TipoPercepcion {
	case "01":
		percent = 2.00
	case "02":
		percent = 1.00
	case "03":
		percent = 0.50
	default:
		return nil
	}

	percepcionMonto := round(f.TotalImportePagar * (percent / 100))
	totalConPercepcion := round(f.TotalImportePagar + percepcionMonto)

	return &UBLExtension{
	ExtensionContent: ExtensionContent{
		SUNATPerception: &SUNATPerception{
			SystemCode:         f.TipoPercepcion,
			Percent:            percent,
			TotalInvoiceAmount: newAmount(f.TotalImportePagar, f.Moneda),
			PerceptionAmount:   newAmount(percepcionMonto, f.Moneda),
			PerceptionDate:     f.FechaEmision,
			NetTotalPaid:       newAmount(totalConPercepcion, f.Moneda),
		},
	},
}
}
func round(val float64) float64 {
	return math.Round(val*100) / 100
}
