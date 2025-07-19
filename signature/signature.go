/*
Paquete de Firma Digital para Documentos XML SUNAT
================================================

Este paquete maneja la firma digital de documentos XML según los estándares XMLDSig
y las especificaciones técnicas de SUNAT para facturación electrónica.

Funcionalidades:
1. Carga de certificados digitales PKCS#12 (.pfx)
2. Firma XMLDSig enveloped del documento XML completo
3. Inserción de la firma en la extensión UBL correcta
4. Generación de DigestValue (SHA1) y SignatureValue (RSA)

Cumple con:
- XML Digital Signature (XMLDSig) W3C
- Algoritmo de canonicalización C14N Exclusive
- Hash SHA1 para digest
- Firma RSA con certificados X.509
- Especificaciones técnicas SUNAT

La firma se inserta en <ext:ExtensionContent> del documento UBL.
*/
package signature

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"os"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
	"software.sslmate.com/src/go-pkcs12"
)

/*
X509KeyStore implementa la interfaz KeyStore requerida por goxmldsig.
Almacena la clave privada RSA y el certificado X.509 para la firma digital.

Esta estructura encapsula los elementos criptográficos necesarios:
- PrivateKey: Clave privada RSA para generar la firma
- Certificate: Certificado X.509 que contiene la clave pública y metadatos
*/
type X509KeyStore struct {
	PrivateKey  *rsa.PrivateKey    // Clave privada RSA extraída del PKCS#12
	Certificate *x509.Certificate  // Certificado X.509 con clave pública y metadatos
}

// GetKeyPair implementa la interfaz KeyStore de goxmldsig
// Retorna la clave privada y el certificado en formato raw (DER)
func (ks *X509KeyStore) GetKeyPair() (*rsa.PrivateKey, []byte, error) {
	return ks.PrivateKey, ks.Certificate.Raw, nil
}

/*
FirmaXML es la función principal que firma digitalmente un archivo XML.
Implementa el proceso completo de firma XMLDSig según especificaciones SUNAT.

Parámetros:
- xmlPath: Ruta del archivo XML a firmar
- pfxPath: Ruta del certificado PKCS#12 (.pfx)
- pfxPassword: Contraseña del certificado

Retorna:
- string: DigestValue (hash SHA1 del contenido firmado)
- string: SignatureValue (firma RSA en base64)
- error: Error si algo falla en el proceso

Proceso:
1. Cargar y parsear el XML
2. Extraer certificado del archivo PKCS#12
3. Configurar contexto de firma XMLDSig
4. Firmar el documento completo (enveloped signature)
5. Insertar firma en <ext:ExtensionContent>
6. Guardar XML firmado
7. Extraer valores de digest y signature
*/
func FirmaXML(xmlPath, pfxPath, pfxPassword string) (string, string, error) {
	// ==================== CARGA Y PARSEO DEL XML ====================
	
	// Crear documento etree para manipulación XML
	doc := etree.NewDocument()
	// Configurar lector de caracteres para manejar encoding
	doc.ReadSettings.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	// Cargar archivo XML desde disco
	if err := doc.ReadFromFile(xmlPath); err != nil {
		return "", "", fmt.Errorf("error leyendo XML: %v", err)
	}

	// Obtener elemento raíz del documento para la firma
	root := doc.Root()

	// ==================== CARGA DEL CERTIFICADO DIGITAL ====================
	
	// Leer archivo PKCS#12 (.pfx) desde disco
	pfxData, err := os.ReadFile(pfxPath)
	if err != nil {
		return "", "", fmt.Errorf("error leyendo PFX: %v", err)
	}
	
	// Decodificar PKCS#12 para extraer clave privada y certificado
	// PKCS#12 es el formato estándar para almacenar certificados digitales
	privKeyIface, cert, err := pkcs12.Decode(pfxData, pfxPassword)
	if err != nil {
		return "", "", fmt.Errorf("error decodificando PFX: %v", err)
	}
	
	// Verificar que la clave privada sea RSA (requerido por SUNAT)
	privKey, ok := privKeyIface.(*rsa.PrivateKey)
	if !ok {
		return "", "", fmt.Errorf("la clave privada no es RSA")
	}

	// ==================== CONFIGURACIÓN DE FIRMA XMLDSIG ====================
	
	// Crear almacén de claves con el certificado cargado
	keyStore := &X509KeyStore{PrivateKey: privKey, Certificate: cert}
	
	// Crear contexto de firma con configuraciones SUNAT
	ctx := dsig.NewDefaultSigningContext(keyStore)
	// Configurar canonicalización C14N Exclusive (requerido por SUNAT)
	ctx.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")

	// ==================== LOCALIZACIÓN DEL PUNTO DE INSERCIÓN ====================
	
	// Buscar el nodo <ext:ExtensionContent> donde se insertará la firma
	// SUNAT requiere que la firma vaya dentro de la primera extensión UBL
	extNodes := doc.FindElements("//ext:ExtensionContent")
	if len(extNodes) == 0 {
		return "", "", fmt.Errorf("no se encontró <ext:ExtensionContent>")
	}

	// Firmar el documento completo
	signedDoc, err := ctx.SignEnveloped(root)
	if err != nil {
		return "", "", fmt.Errorf("error firmando XML: %v", err)
	}

	signature := signedDoc.FindElement(".//ds:Signature")
	if signature == nil {
		return "", "", fmt.Errorf("no se encontró <ds:Signature>")
	}
	signature.CreateAttr("Id", "SignatureSP")

	// Insertar la firma en el nodo <ext:ExtensionContent>
	extNodes[0].AddChild(signature)

	if err := doc.WriteToFile(xmlPath); err != nil {
		return "", "", fmt.Errorf("error guardando XML firmado: %v", err)
	}

	var digestValue, signatureValue string
	if ref := signature.FindElement(".//ds:Reference"); ref != nil {
		if dv := ref.FindElement("ds:DigestValue"); dv != nil {
			digestValue = dv.Text()
		}
	}
	if sv := signature.FindElement("ds:SignatureValue"); sv != nil {
		signatureValue = sv.Text()
	}

	return digestValue, signatureValue, nil
}
