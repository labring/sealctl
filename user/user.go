package user

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

type Config struct {
	CACrtFile string // ca file, default is /etc/kubernetes/pki/ca.crt
	CAKeyFile string // ca key file, default is /etc/kubernetes/pki/ca.key
	OutPut string // kubconfig output file name default is ./kube/config
	User string
	Groups []string
	ClusterName string // default is kubernetes
	Apiserver string // default is https://apiserver.cluster.local:6443
	DNSNames []string
	IPAddresses []net.IP
}

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

func GenerateKubeconfig(conf Config) error{
	certs, err := cert.CertsFromFile(conf.CACrtFile)
	caCert := certs[0]
	if err != nil {
		return fmt.Errorf("load ca and cert failed %s", err)
	}
	cert := EncodeCertPEM(caCert)
	caKey,err := TryLoadKeyFromDisk(conf.CAKeyFile)
	if err != nil {
		return fmt.Errorf("load ca key file failed %s",err)
	}
	clientCert,clientKey,err := NewCertAndKey(caCert,caKey,conf.User,conf.Groups,conf.DNSNames,conf.IPAddresses)
	if err != nil {
		return fmt.Errorf("new client key failed %s",err)
	}
	encodedClientKey,err := keyutil.MarshalPrivateKeyToPEM(clientKey)
	if err != nil {
		return fmt.Errorf("encode client key failed %s", err)
	}
	encodedClientCert := EncodeCertPEM(clientCert)
	ctx := fmt.Sprintf("%s@%s",conf.User,conf.ClusterName)
	config := &api.Config{
		Clusters: map[string]*api.Cluster{
			conf.ClusterName: {
				Server: conf.Apiserver,
				CertificateAuthorityData: cert,
			},
		},
		Contexts: map[string]*api.Context{
			ctx: {
				Cluster:  conf.ClusterName,
				AuthInfo: conf.User,
			},
		},
		AuthInfos:      map[string]*api.AuthInfo{
			conf.User:&api.AuthInfo{
				ClientCertificateData: encodedClientCert,
				ClientKeyData:         encodedClientKey,
			},
		},
		CurrentContext: ctx,
	}

	err = clientcmd.WriteToFile(*config, conf.OutPut)
	if err != nil {
		return fmt.Errorf("write kubeconfig failed %s", err)
	}
	return nil
}

func NewCertAndKey(caCert *x509.Certificate, caKey crypto.Signer, user string, groups []string, DNSNames []string,IPAddresses []net.IP) (*x509.Certificate, crypto.Signer, error) {
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
			CommonName:   user,
			Organization: groups,
		},
		DNSNames:     DNSNames,
		IPAddresses:  IPAddresses,
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