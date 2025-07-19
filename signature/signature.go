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

type X509KeyStore struct {
	PrivateKey  *rsa.PrivateKey
	Certificate *x509.Certificate
}

func (ks *X509KeyStore) GetKeyPair() (*rsa.PrivateKey, []byte, error) {
	return ks.PrivateKey, ks.Certificate.Raw, nil
}



func FirmaXML(xmlPath, pfxPath, pfxPassword string) (string, string, error) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	if err := doc.ReadFromFile(xmlPath); err != nil {
		return "", "", fmt.Errorf("error leyendo XML: %v", err)
	}

	root := doc.Root()

	// Leer el certificado PFX
	pfxData, err := os.ReadFile(pfxPath)
	if err != nil {
		return "", "", fmt.Errorf("error leyendo PFX: %v", err)
	}
	privKeyIface, cert, err := pkcs12.Decode(pfxData, pfxPassword)
	if err != nil {
		return "", "", fmt.Errorf("error decodificando PFX: %v", err)
	}
	privKey, ok := privKeyIface.(*rsa.PrivateKey)
	if !ok {
		return "", "", fmt.Errorf("la clave privada no es RSA")
	}

	keyStore := &X509KeyStore{PrivateKey: privKey, Certificate: cert}
	ctx := dsig.NewDefaultSigningContext(keyStore)
	ctx.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")

	// Buscar el nodo <ext:ExtensionContent>
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
