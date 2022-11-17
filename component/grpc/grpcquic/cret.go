package grpcquic

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"time"

	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
)

func GenerateTlsSelfSignedCert() (tls.Certificate, error) {
	rootCa, rootKey, err := GenerateSelfSignedCertKey(2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	acPemBlock := pem.EncodeToMemory(&pem.Block{
		Type:    "CERTIFICATE",
		Headers: nil,
		Bytes:   rootCa.Raw,
	})
	keyPemBlock := pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(rootKey),
	})
	return tls.X509KeyPair(acPemBlock, keyPemBlock)
}

func GenerateSelfSignedCertKey(keySize int) (*x509.Certificate, *rsa.PrivateKey, error) {
	// 1.生成密钥对
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	// 2.创建证书模板
	serialNumberByte := make([]byte, 16)
	rand.Read(serialNumberByte)
	template := x509.Certificate{
		SerialNumber: big.NewInt(0).SetBytes(serialNumberByte), // 该号码表示CA颁发的唯一序列号，在此使用一个数来代表
		Issuer:       pkix.Name{},
		Subject:      pkix.Name{CommonName: fmt.Sprintf("%d", time.Now().Unix())},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, // 表示该证书是用来做服务端认证的
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// if ip := net.ParseIP(host); ip != nil {
	//   template.IPAddresses = append(template.IPAddresses, ip)
	// } else {
	//   template.DNSNames = append(template.DNSNames, host)
	// }

	// template.IPAddresses = append(template.IPAddresses, alternateIPs...)
	// template.DNSNames = append(template.DNSNames, alternateDNS...)

	// 3.创建证书,这里第二个参数和第三个参数相同则表示该证书为自签证书，返回值为DER编码的证书
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	rootCa, err := x509.ParseCertificate(certificate)
	return rootCa, priv, nil
}

func GenerateSelfSignedCertKey2(keySize int, host string, alternateIPs []net.IP, alternateDNS []string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// 1.生成密钥对
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	// 2.创建证书模板
	serialNumberByte := make([]byte, 16)
	rand.Read(serialNumberByte)
	template := x509.Certificate{
		SerialNumber: big.NewInt(0).SetBytes(serialNumberByte), // 该号码表示CA颁发的唯一序列号，在此使用一个数来代表
		Issuer:       pkix.Name{},
		Subject:      pkix.Name{CommonName: fmt.Sprintf("%s@%d", host, time.Now().Unix())},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, // 表示该证书是用来做服务端认证的
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	template.IPAddresses = append(template.IPAddresses, alternateIPs...)
	template.DNSNames = append(template.DNSNames, alternateDNS...)

	// 3.创建证书,这里第二个参数和第三个参数相同则表示该证书为自签证书，返回值为DER编码的证书
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	rootCa, err := x509.ParseCertificate(certificate)
	return rootCa, priv, nil
	// // 4.将得到的证书放入pem.Block结构体中
	// block := pem.Block{
	//   Type:    "CERTIFICATE",
	//   Headers: nil,
	//   Bytes:   certificate,
	// }
	// // 5.通过pem编码并写入磁盘文件
	// file, err := os.Create("ca.crt")
	// if err != nil {
	//   panic(err)
	// }
	// defer file.Close()
	// pem.Encode(file, &block)
	//
	// // 6.将私钥中的密钥对放入pem.Block结构体中
	// block = pem.Block{
	//   Type:    "RSA PRIVATE KEY",
	//   Headers: nil,
	//   Bytes:   x509.MarshalPKCS1PrivateKey(priv),
	// }
	// // 7.通过pem编码并写入磁盘文件
	// file, err = os.Create("ca.key")
	// if err != nil {
	//   panic(err)
	// }
	// pem.Encode(file, &block)
}

func CreateCertificate(templateCert *x509.Certificate, rootCa *x509.Certificate, rootKey interface{}) (*x509.Certificate, interface{}, error) {
	// 生成公钥私钥对
	priKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	rawCert, err := x509.CreateCertificate(rand.Reader, templateCert, rootCa, &priKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(rawCert)
	if err != nil {
		log.Error("错误", zap.Error(err))
		return nil, nil, err
	}
	return cert, priKey, nil
}

func LoadCertificate(rootCaPath, rootKeyPath string) (*x509.Certificate, interface{}, error) {
	// 解析根证书
	caFile, err := ioutil.ReadFile(rootCaPath)
	if err != nil {
		return nil, nil, err
	}
	caBlock, _ := pem.Decode(caFile)

	cert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	// 解析私钥
	keyFile, err := ioutil.ReadFile(rootKeyPath)
	if err != nil {
		return nil, nil, err
	}
	keyBlock, _ := pem.Decode(keyFile)
	praKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return cert, praKey, nil
}
