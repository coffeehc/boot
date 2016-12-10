package grpcboot

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/coffeehc/microserviceboot/base"
	"math/big"
	"time"
)

func newDefaultTlsConfig() (*tls.Config, base.Error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(8888),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"xiagaogao"},
			OrganizationalUnit: []string{"com"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: false,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return nil, base.NewErrorWrapper("tslconfig",err)
	}
	cca, _ := x509.ParseCertificate(ca_b)
	pool := x509.NewCertPool()
	pool.AddCert(cca)
	cert := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}
	return &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		NextProtos:   []string{"h2"},
	}, nil
}
