package grpcserver

import (
	"context"
	"time"

	"google.golang.org/grpc/keepalive"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcrecovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

var scope = zap.String("scope", "grpc.server")

func NewServer(ctx context.Context, grpcConfig *GRPCServerConfig) (*grpc.Server, errors.Error) {
	grpcConfig = &GRPCServerConfig{}
	if !viper.IsSet("grpc") {
		log.Warn("没有配置GRPCConfig,使用默认配置", scope)
	}
	err := viper.UnmarshalKey("grpc", grpcConfig)
	if err != nil {
		log.Fatal("解析grpc配置失败", zap.Error(err), scope)
	}
	// server := grpc.NewServer()
	server := grpc.NewServer(BuildGRPCServerOptions(ctx, grpcConfig)...)
	return server, nil
}

func BuildGRPCServerOptions(ctx context.Context, config *GRPCServerConfig) []grpc.ServerOption {
	chainUnaryServers := []grpc.UnaryServerInterceptor{
		DebugLoggingInterceptor(),
		grpc_prometheus.UnaryServerInterceptor,
		grpcrecovery.UnaryServerInterceptor(),
	}
	chainStreamServers := []grpc.StreamServerInterceptor{
		grpc_prometheus.StreamServerInterceptor,
		grpcrecovery.StreamServerInterceptor(),
	}
	grpcAuth := ctx.Value(serverGrpcAuthKey)
	if grpcAuth != nil {
		authService, ok := grpcAuth.(GRPCServerAuth)
		if ok {
			chainUnaryServers = append(chainUnaryServers, buildAuthUnaryServerInterceptor(authService))
			chainStreamServers = append(chainStreamServers, buildAuthStreamServerInterceptor(authService))
		}
	}
	opts := []grpc.ServerOption{
		grpc.Creds(getCerts(ctx)),
		grpc.InitialWindowSize(4096),
		grpc.InitialConnWindowSize(1000),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.ChainStreamInterceptor(chainStreamServers...),
		grpc.ChainUnaryInterceptor(chainUnaryServers...),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Minute,
			Timeout:           10 * time.Second, // 類似 ClientParameters.Time 不過默認爲 2小時
			Time:              3 * time.Second,  // 類似 ClientParameters.Timeout 默認 20秒
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{ // 當服務器不允許ping 或 ping 太頻繁超過 MinTime 限制 服務器 會 返回ping失敗 此時 客戶端 不會認爲這個ping是 active RPCs
			MinTime:             time.Second * 3,
			PermitWithoutStream: true,
		}),
	}
	if config.MaxMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(config.MaxMsgSize),
			grpc.MaxSendMsgSize(config.MaxMsgSize))
	}
	return opts
}

// //metadata.FromIncomingContext(ctx)
