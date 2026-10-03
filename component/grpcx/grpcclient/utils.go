package grpcclient

import (
	"context"
	"crypto/tls"
	"github.com/coffeehc/base/log"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	perRPCCredentialsKey  = "_grpc._PerRPCCredentialsKey"
	contextKeyClientCerds = "_grpc.client.Credentials"
)

// dialOptionsContextKey 将连接装配选项与 context 中的其他值隔离。
type dialOptionsContextKey struct{}

// SetDialOptions 在派生 context 中追加原生连接选项，供所有 grpcclient 和 discovery 构造入口使用。
// 父 context 的选项先应用，本次选项后应用；不修改父 context 或调用方的选项切片。
// 选项仅在创建连接时读取，须在构造前设置；覆盖或叠加行为遵循对应 gRPC DialOption。
func SetDialOptions(ctx context.Context, dialOptions ...grpc.DialOption) context.Context {
	parentOptions, _ := ctx.Value(dialOptionsContextKey{}).([]grpc.DialOption)
	options := make([]grpc.DialOption, 0, len(parentOptions)+len(dialOptions))
	options = append(options, parentOptions...)
	options = append(options, dialOptions...)
	return context.WithValue(ctx, dialOptionsContextKey{}, options)
}

// SetPerRPCCredentials 设置每次 RPC 的凭据，prc 必须非 nil 且支持并发调用。
func SetPerRPCCredentials(ctx context.Context, prc credentials.PerRPCCredentials) context.Context {
	return context.WithValue(ctx, perRPCCredentialsKey, prc)
}

// GetPerRPCCredentials 返回当前装配 context 中的 RPC 凭据，未设置时为 nil。
func GetPerRPCCredentials(ctx context.Context) credentials.PerRPCCredentials {
	v := ctx.Value(perRPCCredentialsKey)
	if v == nil {
		return nil
	}
	return v.(credentials.PerRPCCredentials)
}

// SetInsecureSkipVerifyCerds 显式关闭服务端证书校验，仅用于调用方认可的临时自签名环境。
// 生产连接应使用 SetClientCerds 传入具备可信 CA 或证书校验逻辑的凭据。
func SetInsecureSkipVerifyCerds(ctx context.Context) context.Context {
	tlsConfig := &tls.Config{
		NextProtos:         []string{http2.NextProtoTLS},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}
	return SetClientCerds(ctx, credentials.NewTLS(tlsConfig))
}

// SetClientCerds 设置连接凭据，须在创建连接前调用；同一 context 只设置一次。
// 明文 TCP 连接必须显式传入 insecure.NewCredentials()。
func SetClientCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	if ctx.Value(contextKeyClientCerds) != nil {
		log.DPanic("****已经设置了TransportCredentials,不能多次设置****")
	}
	return context.WithValue(ctx, contextKeyClientCerds, creds)
}

// GetClientCerts 返回 context 中的连接凭据，未设置时为 nil。
// 也可由额外 DialOption 提供凭据；两处均未提供时由 gRPC 拒绝创建连接。
func GetClientCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextKeyClientCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}
