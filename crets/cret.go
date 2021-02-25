package crets

import (
	"crypto/tls"
	"crypto/x509"

	"git.xiagaogao.com/coffee/base/log"
	"google.golang.org/grpc/credentials"
)

func BuildServerTLSConfig(caPem, serverPem, serverKey []byte) *tls.Config {
	// 加载服务端私钥和证书
	cert, err := tls.X509KeyPair(serverPem, serverKey)
	if err != nil {
		panic(err)
	}

	// 生成证书池，将根证书加入证书池
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Panic("添加CA错误")
	}

	// 初始化TLSConfig
	// ClientAuth有5种类型，如果要进行双向认证必须是RequireAndVerifyClientCert
	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
	}
}

func BuildClientTLSConfig(caPem, clientPem, clientKey []byte, serverName string) *tls.Config {
	if serverName == "" {
		serverName = "rpcclient.51apis.com"
	}
	// 加载客户端端私钥和证书
	cert, err := tls.X509KeyPair(clientPem, clientKey)
	if err != nil {
		panic(err)
	}

	// 生成证书池，将根证书加入证书池
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Panic("添加CA错误")
	}
	// 新建凭证
	// 注意ServerName需要与服务器证书内的Common Name一致
	// 客户端是根据根证书和ServerName对服务端进行验证的
	return &tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}
}

func NewServerCreds() credentials.TransportCredentials {
	tlsConfig := BuildServerTLSConfig(ca_pem, server_pem, server_key)
	return credentials.NewTLS(tlsConfig)
}

func NewClientCreds(serverName string) credentials.TransportCredentials {
	tlsConfig := BuildClientTLSConfig(ca_pem, client_pem, client_key, serverName)
	return credentials.NewTLS(tlsConfig)
}
