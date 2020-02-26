package grpcserver

import (
	"context"
	"google.golang.org/grpc/grpclog"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcrecovery"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
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
	server := grpc.NewServer(BuildGRPCServerOptions(ctx, grpcConfig)...)
	return server, nil
}

func BuildGRPCServerOptions(ctx context.Context, config *GRPCServerConfig) []grpc.ServerOption {
	grpclog.SetLoggerV2(grpcrecovery.NewZapLogger())
	grpc.EnableTracing = false
	chainUnaryServers := []grpc.UnaryServerInterceptor{
		// DebufLoggingInterceptor(),
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
	chainUnaryServer := grpc_middleware.ChainUnaryServer(chainUnaryServers...)
	chainStreamServer := grpc_middleware.ChainStreamServer(chainStreamServers...)
	opts := []grpc.ServerOption{
		grpc.Creds(getCerts(ctx)),
		grpc.InitialWindowSize(4096),
		grpc.InitialConnWindowSize(100),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.StreamInterceptor(chainStreamServer),
		grpc.UnaryInterceptor(chainUnaryServer),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Minute,
			Timeout:           30 * time.Second, // 類似 ClientParameters.Time 不過默認爲 2小時
			Time:              10 * time.Second, // 類似 ClientParameters.Timeout 默認 20秒
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{ // 當服務器不允許ping 或 ping 太頻繁超過 MinTime 限制 服務器 會 返回ping失敗 此時 客戶端 不會認爲這個ping是 active RPCs
			MinTime:             time.Second * 10,
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
