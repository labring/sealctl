package cert

import (
	"crypto"
	"crypto/x509"
	"fmt"
)

var caList = []Config{
	{
		Path:         "pki",
		BaseName:     "ca",
		CommonName:   "kubernetes",
		Organization: nil,
		Year:         100,
		AltNames:     AltNames{},
		Usages:       nil,
	},
	{
		Path:         "pki",
		BaseName:     "front-proxy-ca",
		CommonName:   "front-proxy-ca",
		Organization: nil,
		Year:         100,
		AltNames:     AltNames{},
		Usages:       nil,
	},
	{
		Path:         "pki/etcd",
		BaseName:     "ca",
		CommonName:   "etcd-ca",
		Organization: nil,
		Year:         100,
		AltNames:     AltNames{},
		Usages:       nil,
	},
}

var certList = []Config{
	{
		Path:         "pki",
		BaseName:     "apiserver",
		CAName:       "kubernetes",
		CommonName:   "kube-apiserver",
		Organization: nil,
		Year:         100,
		AltNames:     AltNames{},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	},
}

func GenerateAll() error {
	CACerts := map[string]*x509.Certificate{}
	CAKeys := map[string]crypto.Signer{}
	for _, ca := range caList {
		caCert, caKey, err := NewCaCertAndKey(ca)
		if err != nil {
			return err
		}
		CACerts[ca.CommonName] = caCert
		CAKeys[ca.CommonName] = caKey

		err = WriteCertAndKey(ca.Path, ca.BaseName, caCert, caKey)
		if err != nil {
			return err
		}
	}

	for _, cert := range certList {
		caCert,ok := CACerts[cert.CAName]
		if !ok {
			return fmt.Errorf("root ca cert not found %s",cert.CAName)
		}
		caKey,ok := CAKeys[cert.CAName]
		if !ok {
			return fmt.Errorf("root ca key not found %s",cert.CAName)
		}

		Cert,Key,err := NewCaCertAndKeyFromRoot(cert,caCert,caKey)
		if err != nil {
			return err
		}
		err = WriteCertAndKey(cert.Path,cert.BaseName,Cert,Key)
		if err != nil {
			return err
		}
	}
	return nil
}
