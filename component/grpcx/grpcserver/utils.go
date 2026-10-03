package grpcserver

import (
	"context"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	viper.SetDefault("grpc.MaxConnectionIdle", time.Minute*30)
}

// GetMaxConnectionIdle 返回无活动 RPC 的连接保留时长；零值表示不限制。
// 优先读取 snake_case 配置，同时兼容已有 MaxConnectionIdle 配置键。
func GetMaxConnectionIdle() time.Duration {
	if viper.IsSet("grpc.max_connection_idle") {
		return viper.GetDuration("grpc.max_connection_idle")
	}
	return viper.GetDuration("grpc.MaxConnectionIdle")
}

// SetMaxConnectionIdle 设置无活动 RPC 的连接保留时长，须在创建 server 前调用。
func SetMaxConnectionIdle(idle time.Duration) {
	viper.Set("grpc.max_connection_idle", idle)
}

// EnableAccessLog 控制 unary 访问日志，须在创建 server 前设置。
var EnableAccessLog bool = false

// DebugLoggingInterceptor 在请求结束时记录方法、耗时和错误，不记录请求内容。
func DebugLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			log.Debug("gRPC 请求处理完成", zap.String("method", info.FullMethod), zap.Duration("duration", time.Since(start)), zap.Error(err), scope)
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}

const (
	serverGrpcAuthKey     = "__ServerGrpcAuthKey"
	contextKeyServerCerds = "_grpc.server.Credentials"
)

// SetServerCerds 设置非 nil 连接凭据，须在创建 server 前调用；同一 context 只设置一次。
func SetServerCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	if ctx.Value(contextKeyServerCerds) != nil {
		log.DPanic("****已经设置了TransportCredentials,不能多次设置****")
	}
	return context.WithValue(ctx, contextKeyServerCerds, creds)
}

// GetServerCerts 返回连接凭据，未设置时 server 使用明文 TCP。
func GetServerCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextKeyServerCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}

// SetGrpcAuth 设置并发安全的请求鉴权器，须在创建 server 前调用。
func SetGrpcAuth(ctx context.Context, auth GRPCServerAuth) context.Context {
	return context.WithValue(ctx, serverGrpcAuthKey, auth)
}

// SetSelfSignedCerds 为临时环境创建自签名证书，生成失败会 panic，避免降级为明文。
// 证书没有主机名 SAN，客户端需要显式固定证书校验；生产应使用受信任的服务证书。
func SetSelfSignedCerds(ctx context.Context) context.Context {
	cret, pk, err := generateSelfSignedCertKey(2048)
	if err != nil {
		log.Panic("创建自签名证书失败", zap.Error(err))
	}
	tlsCrt := &tls.Certificate{
		Certificate: [][]byte{cret.Raw},
		Leaf:        cret,
		PrivateKey:  pk,
	}
	return SetServerCerds(ctx, credentials.NewServerTLSFromCert(tlsCrt))
}

func generateSelfSignedCertKey(keySize int) (*x509.Certificate, *rsa.PrivateKey, error) {
	// 1.生成密钥对
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	// 2.创建证书模板
	serialNumberByte := make([]byte, 16)
	if _, err := rand.Read(serialNumberByte); err != nil {
		return nil, nil, err
	}
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
