package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"ubl-go-conversor/models"
)

func ValidarComprobanteBase(f models.ComprobanteBase) error {
	if err := verificarCamposObligatorios(f); err != nil {
		return fmt.Errorf("faltan campos obligatorios: %v", err)
	}

	if err := validarEmisor(f.Emisor); err != nil {
		return fmt.Errorf("error en emisor: %v", err)
	}

	if err := validarCliente(f.Cliente, f.TipoDocumento); err != nil {
		return fmt.Errorf("error en cliente: %v", err)
	}

	if err := validarCamposBasicos(f); err != nil {
		return err
	}

	if len(f.Items) == 0 {
		return errors.New("la factura debe tener al menos un ítem")
	}
	for i, item := range f.Items {
		if err := validarItem(item, i); err != nil {
			return err
		}
	}

	if err := validarTotales(f); err != nil {
		return err
	}

	return nil
}

func verificarCamposObligatorios(f models.ComprobanteBase) error {
	esGratuito := false
	for _, item := range f.Items {
		if item.TipoAfectacionIGV == "21" {
			esGratuito = true
			break
		}
	}
	if f.Serie == "" {
		return errors.New("serie es obligatoria")
	}
	if f.Numero == "" {
		return errors.New("número es obligatorio")
	}
	if f.FechaEmision == "" {
		return errors.New("fechaEmision es obligatoria")
	}
	if f.HoraEmision == "" {
		return errors.New("horaEmision es obligatoria")
	}
	if f.TipoDocumento == "" {
		return errors.New("tipoDocumento es obligatorio")
	}
	if f.Moneda == "" {
		return errors.New("moneda es obligatoria")
	}
	if f.FormaPago == "" {
		return errors.New("formaPago es obligatoria")
	}
	if !esGratuito && f.TotalGravado == 0 && f.TotalIGV == 0 && f.TotalPrecioVenta == 0 {
		return errors.New("los totales no pueden estar todos en cero")

	}
	if f.TotalImportePagar == 0 && !esGratuito {
		return errors.New("totalImportePagar es obligatorio")
	}
	if f.Emisor.RUC == "" || f.Emisor.RazonSocial == "" || f.Emisor.Direccion == "" {
		return errors.New("datos obligatorios del emisor incompletos")
	}
	if f.Cliente.NumeroDoc == "" || f.Cliente.TipoDoc == "" || f.Cliente.RazonSocial == "" {
		return errors.New("datos obligatorios del cliente incompletos")
	}
	return nil
}

func validarEmisor(emisor models.Emisor) error {
	if len(emisor.RUC) != 11 {
		return errors.New("el RUC debe tener 11 dígitos")
	}
	if _, err := strconv.Atoi(emisor.RUC); err != nil {
		return errors.New("el RUC debe contener solo números")
	}
	if emisor.RazonSocial == "" {
		return errors.New("la razón social es obligatoria")
	}
	if emisor.Direccion == "" {
		return errors.New("la dirección es obligatoria")
	}
	return nil
}

func validarCliente(cliente models.Cliente, tipoComprobante string) error {
	tiposValidos := map[string]bool{
		"1": true, // DNI
		"6": true, // RUC
		"4": true, // Carnet extranjería
		"7": true, // Pasaporte
	}

	if !tiposValidos[cliente.TipoDoc] {
		return fmt.Errorf("tipo de documento '%s' no válido", cliente.TipoDoc)
	}

	switch cliente.TipoDoc {
	case "1":
		if len(cliente.NumeroDoc) != 8 {
			return errors.New("el DNI debe tener 8 dígitos")
		}
	case "6":
		if len(cliente.NumeroDoc) != 11 {
			return errors.New("el RUC debe tener 11 dígitos")
		}
	}

	if cliente.TipoDoc == "1" || cliente.TipoDoc == "6" {
		if _, err := strconv.Atoi(cliente.NumeroDoc); err != nil {
			return errors.New("el número de documento debe contener solo números")
		}
	}

	if tipoComprobante == "01" && cliente.TipoDoc != "6" {
		return errors.New("las facturas (01) solo pueden emitirse a clientes con RUC (tipo 6)")
	}
	if tipoComprobante == "03" && cliente.TipoDoc == "6" {
		return errors.New("las boletas (03) no deben emitirse a clientes con RUC (tipo 6), use DNI u otro")
	}

	return nil
}

func validarCamposBasicos(f models.ComprobanteBase) error {
	tiposDocumento := map[string]bool{
		"01": true, "03": true, "07": true,
	}

	if !tiposDocumento[f.TipoDocumento] {
		return fmt.Errorf("tipo de documento '%s' no válido", f.TipoDocumento)
	}

	serieRegex := regexp.MustCompile(`^[A-Z][A-Z0-9]{3}$`)
	if !serieRegex.MatchString(f.Serie) {
		return fmt.Errorf("la serie '%s' debe tener formato válido (ej: F001, B001)", f.Serie)
	}

	switch f.TipoDocumento {
	case "01":
		if f.Serie[0] != 'F' {
			return fmt.Errorf("para facturas, la serie debe comenzar con 'F'")
		}
	case "03":
		if f.Serie[0] != 'B' {
			return fmt.Errorf("para boletas, la serie debe comenzar con 'B'")
		}
	case "07":
		if f.Serie[0] != 'F' && f.Serie[0] != 'B' {
			return fmt.Errorf("para notas de crédito, la serie debe comenzar con 'F' o 'B'")
		}
	}

	if len(f.Numero) == 0 || len(f.Numero) > 8 {
		return errors.New("el número debe tener entre 1 y 8 dígitos")
	}

	if _, err := time.Parse("2006-01-02", f.FechaEmision); err != nil {
		return errors.New("la fecha de emisión tiene formato inválido (YYYY-MM-DD)")
	}

	if f.HoraEmision != "" {
		horaRegex := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`)
		if !horaRegex.MatchString(f.HoraEmision) {
			return errors.New("la hora de emisión debe tener formato HH:MM:SS")
		}
	}

	if f.FechaVencimiento != "" {
		venc, err1 := time.Parse("2006-01-02", f.FechaVencimiento)
		emision, err2 := time.Parse("2006-01-02", f.FechaEmision)
		if err1 != nil || err2 != nil {
			return errors.New("formato de fecha inválido en vencimiento o emisión")
		}
		if venc.Before(emision) {
			return errors.New("la fecha de vencimiento no puede ser anterior a la fecha de emisión")
		}
	}

	monedasValidas := regexp.MustCompile(`^(PEN|USD|EUR)$`)
	if !monedasValidas.MatchString(f.Moneda) {
		return fmt.Errorf("la moneda '%s' no es válida (PEN, USD, EUR)", f.Moneda)
	}

	return nil
}

func validarItem(item models.ItemComprobante, indice int) error {
	if item.Descripcion == "" {
		return fmt.Errorf("el ítem %d debe tener descripción", indice+1)
	}
	if item.Cantidad <= 0 {
		return fmt.Errorf("el ítem %d debe tener cantidad mayor a 0", indice+1)
	}
	if item.ValorUnitario < 0 {
		return fmt.Errorf("el ítem %d no puede tener valor unitario negativo", indice+1)
	}

	tiposAfectacion := map[string]bool{
		"10": true, "11": true, "12": true, "13": true, "14": true, "15": true,
		"16": true, "17": true, "20": true, "21": true, "30": true, "31": true,
		"32": true, "33": true, "34": true, "35": true, "36": true, "37": true, "40": true,
	}

	if !tiposAfectacion[item.TipoAfectacionIGV] {
		return fmt.Errorf("el ítem %d tiene tipo de afectación IGV inválido: %s", indice+1, item.TipoAfectacionIGV)
	}

	if item.TipoAfectacionIGV != "21" {
		expected := item.ValorUnitario * item.Cantidad
		if abs(item.ValorTotal-expected) > 0.01 {
			return fmt.Errorf("el ítem %d: valor total inconsistente (esperado: %.2f, actual: %.2f)",
				indice+1, expected, item.ValorTotal)
		}
	}

	return nil
}

func validarTotales(f models.ComprobanteBase) error {
	var sumaGravado, sumaExonerado, sumaInafecto, sumaIGV float64

	for _, item := range f.Items {
		switch item.TipoAfectacionIGV {
		case "21":
			continue
		case "10", "11", "12", "13", "14", "15", "16", "17":
			sumaGravado += item.ValorTotal
		case "20", "40":
			sumaExonerado += item.ValorTotal
		case "30", "31", "32", "33", "34", "35", "36", "37":
			sumaInafecto += item.ValorTotal
		}
		sumaIGV += item.IGV
	}

	if abs(f.TotalGravado-sumaGravado) > 0.01 {
		return fmt.Errorf("total gravado inconsistente (esperado: %.2f, actual: %.2f)", sumaGravado, f.TotalGravado)
	}

	if abs(f.TotalIGV-sumaIGV) > 0.01 {
		return fmt.Errorf("total IGV inconsistente (esperado: %.2f, actual: %.2f)", sumaIGV, f.TotalIGV)
	}

	totalEsperado := sumaGravado + sumaExonerado + sumaInafecto + sumaIGV
	if abs(f.TotalPrecioVenta-totalEsperado) > 0.01 {
		return fmt.Errorf("total precio venta inconsistente (esperado: %.2f, actual: %.2f)", totalEsperado, f.TotalPrecioVenta)
	}

	if abs(f.TotalImportePagar-f.TotalPrecioVenta) > 0.01 {
		return errors.New("total importe a pagar debe ser igual al total precio venta")
	}

	return nil
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
