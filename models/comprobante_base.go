package models

type ComprobanteBase struct {
	Serie             string        `json:"serie"`
	Numero            string        `json:"numero"`
	FechaEmision      string        `json:"fechaEmision"`
	HoraEmision       string        `json:"horaEmision"`
	FechaVencimiento  string        `json:"fechaVencimiento,omitempty"`
	TipoDocumento     string        `json:"tipoDocumento"`
	Moneda            string        `json:"moneda"`
	Emisor            Emisor        `json:"emisor"`
	Cliente           Cliente       `json:"cliente"`
	TotalGravado      float64       `json:"totalGravado"`
	TotalIGV          float64       `json:"totalIGV"`
	TotalPrecioVenta  float64       `json:"totalPrecioVenta"`
	TotalImportePagar float64       `json:"totalImportePagar"`
	FormaPago		  string        `json:"formaPago"`
	Cuotas            []Cuota       `json:"cuotas,omitempty"`
	Items             []ItemComprobante `json:"items"`
	Leyendas          []Leyenda     `json:"leyendas"`
	TipoPercepcion    string        `json:"tipoPercepcion,omitempty"`
}
type Leyenda struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}
type Emisor struct {
	RUC             string `json:"ruc"`
	RazonSocial     string `json:"razonSocial"`
	NombreComercial string `json:"nombreComercial"`
	Ubigeo          string `json:"ubigeo"`
	Direccion       string `json:"direccion"`
	Departamento    string `json:"departamento"`
	Provincia       string `json:"provincia"`
	Distrito        string `json:"distrito"`
	CodigoPais      string `json:"codigoPais"`
	Correo          string `json:"correo"`
}

type Cliente struct {
	NumeroDoc    string `json:"numeroDoc"`
	RazonSocial  string `json:"razonSocial"`
	TipoDoc      string `json:"tipoDoc"` 
	Ubigeo       string `json:"ubigeo"`
	Direccion    string `json:"direccion"`
	Departamento string `json:"departamento"`
	Provincia    string `json:"provincia"`
	Distrito     string `json:"distrito"`
	CodigoPais   string `json:"codigoPais"`
	Correo       string `json:"correo"`
}

type ItemComprobante struct {
	ID                  string  `json:"id"`
	Cantidad            float64 `json:"cantidad"`
	UnidadMedida        string  `json:"unidadMedida"`
	Descripcion         string  `json:"descripcion"`
	ValorUnitario       float64 `json:"valorUnitario"`
	PrecioVentaUnitario float64 `json:"precioVentaUnitario"`
	ValorTotal          float64 `json:"valorTotal"` 
	IGV                 float64 `json:"igv"`
	CodigoProducto      string  `json:"codigoProducto"`
	CodigoProductoSUNAT string  `json:"codigoProductoSUNAT"` 
	CodigoTipoPrecio    string  `json:"codigoTipoPrecio"`   
	TipoAfectacionIGV   string  `json:"tipoAfectacionIGV"`   
	CodigoTributo       string  `json:"codigoTributo"`           
	UNSPSC              string  `json:"unspsc"`
}
type Cuota struct {
	NumeroCuota       string  `json:"numero"`       
	Importe      float64 `json:"importe"`     
	FechaVencimiento string `json:"fechaVencimiento"` 
}
