package grpcserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"math/big"
	"time"

	"github.com/coffeehc/base/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	viper.SetDefault("grpc.MaxConnectionIdle", time.Minute*30)
}

func GetMaxConnectionIdle() time.Duration {
	return viper.Get("grpc.MaxConnectionIdle").(time.Duration)
}

func SetMaxConnectionIdle(idle time.Duration) {
	viper.Set("grpc.MaxConnectionIdle", idle)
}

var EnableAccessLog bool = false

func DebugLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			log.Debug(fmt.Sprintf("FullMethod %s, took=%s, err=%v", info.FullMethod, time.Since(start), err), scope)
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}

const (
	contextkeyServerCerds = "_grpc.serverCredentials"
	serverGrpcAuthKey     = "__ServerGrpcAuthKey"
)

func SetGrpcAuth(ctx context.Context, auth GRPCServerAuth) context.Context {
	return context.WithValue(ctx, serverGrpcAuthKey, auth)
}

func SetSelfSignedCerds(ctx context.Context) context.Context {
	cret, pk, err := generateSelfSignedCertKey(1024)
	if err != nil {
		log.Error("创建自签名证书失败", zap.Error(err))
		return ctx
	}
	tlsCrt := &tls.Certificate{
		Certificate: [][]byte{cret.Raw},
		Leaf:        cret,
		PrivateKey:  pk,
	}
	return SetCerds(ctx, credentials.NewServerTLSFromCert(tlsCrt))
}

func SetCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	return context.WithValue(ctx, contextkeyServerCerds, creds)
}

func getCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextkeyServerCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}

func generateSelfSignedCertKey(keySize int) (*x509.Certificate, *rsa.PrivateKey, error) {
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

	// 3.创建证书,这里第二个参数和第三个参数相同则表示该证书为自签证书，返回值为DER编码的证书
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	rootCa, err := x509.ParseCertificate(certificate)
	return rootCa, priv, nil
}
