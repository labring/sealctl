package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"math"
	"math/big"
	"net"
	"time"
)

// TryLoadKeyFromDisk tries to load the key from the disk and validates that it is valid
func TryLoadKeyFromDisk(pkiPath string) (crypto.Signer, error) {
	// Parse the private key from a file
	privKey, err := keyutil.PrivateKeyFromFile(pkiPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't load the private key file %s", err)
	}

	// Allow RSA and ECDSA formats only
	var key crypto.Signer
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		key = k
	case *ecdsa.PrivateKey:
		key = k
	default:
		return nil, fmt.Errorf("couldn't convert the private key file %s", err)
	}

	return key, nil
}

// EncodeCertPEM returns PEM-endcoded certificate data
func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

func main() {
	certs, err := cert.CertsFromFile("./ca/ca.crt")
	caCert := certs[0]
	if err != nil {
		fmt.Printf("load ca and cert failed %s", err)
		return
	}
	cert := EncodeCertPEM(caCert)
	caKey,err := TryLoadKeyFromDisk("./ca/ca.key")
	clientCert,clientKey,err := NewCertAndKey(caCert,caKey)
	encodedClientKey,err := keyutil.MarshalPrivateKeyToPEM(clientKey)
	encodedClientCert := EncodeCertPEM(clientCert)
	config := &api.Config{
		Clusters: map[string]*api.Cluster{
			"kubernetes": {
				Server: "https://store.lameleg.com:6443",
				CertificateAuthorityData: cert,
			},
		},
		Contexts: map[string]*api.Context{
			"fanux@kubernetes": {
				Cluster:  "kubernetes",
				AuthInfo: "fanux",
			},
		},
		AuthInfos:      map[string]*api.AuthInfo{
			"fanux":&api.AuthInfo{
				ClientCertificateData: encodedClientCert,
				ClientKeyData:         encodedClientKey,
			},
		},
		CurrentContext: "fanux@kubernetes",
	}

	err = clientcmd.WriteToFile(*config, "./config/kubeconfig")
	if err != nil {
		fmt.Printf("write kubeconfig failed %s", err)
	}
}

func NewCertAndKey(caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, crypto.Signer, error) {
	key,err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil,nil, fmt.Errorf("generate client key error %s", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil,nil, fmt.Errorf("rand serial error %s", err)
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   "fanux",
			Organization: []string{"sealyun","sealos"},
		},
		DNSNames:     []string{"store.lameleg.com"},
		IPAddresses:  []net.IP{},
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(time.Hour * 24 * 365 * 99).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil,nil,fmt.Errorf("create cert failed %s", err)
	}
	cert,err := x509.ParseCertificate(certDERBytes)
	if err != nil {
		return nil,nil,fmt.Errorf("parse cert failed %s", err)
	}
	return cert,key,nil
}



