package grpcserver

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	_ "git.xiagaogao.com/coffee/boot/transport"
	"git.xiagaogao.com/coffee/boot/transport/grpcrecovery"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func NewServer(ctx context.Context, grpcConfig *GRPCConfig, serviceInfo *boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) (*grpc.Server, errors.Error) {
	if grpcConfig == nil {
		logger.Warn("没有配置GRPCConfig,使用默认配置")
		grpcConfig = &GRPCConfig{}
	}
	grpcConfig.initGRPCConfig()
	server := grpc.NewServer(BuildGRPCOptions(ctx, grpcConfig, serviceInfo, errorService, logger)...)
	return server, nil
}

func BuildGRPCOptions(ctx context.Context, config *GRPCConfig, serviceInfo *boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) []grpc.ServerOption {
	grpc.EnableTracing = false
	chainUnaryServer := grpc_middleware.ChainUnaryServer(
		grpc_prometheus.UnaryServerInterceptor,
		grpcrecovery.UnaryServerInterceptor(errorService, logger),
	)
	chainStreamServer := grpc_middleware.ChainStreamServer(
		grpc_prometheus.StreamServerInterceptor,
		grpcrecovery.StreamServerInterceptor(errorService, logger),
	)
	return []grpc.ServerOption{
		grpc.InitialWindowSize(4096),
		grpc.InitialConnWindowSize(100),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.MaxRecvMsgSize(config.MaxMsgSize),
		grpc.MaxSendMsgSize(config.MaxMsgSize),
		grpc.StreamInterceptor(chainStreamServer),
		grpc.UnaryInterceptor(chainUnaryServer),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Minute,
			Timeout:           30 * time.Second, //類似 ClientParameters.Time 不過默認爲 2小時
			Time:              10 * time.Second, //類似 ClientParameters.Timeout 默認 20秒
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{ //當服務器不允許ping 或 ping 太頻繁超過 MinTime 限制 服務器 會 返回ping失敗 此時 客戶端 不會認爲這個ping是 active RPCs
			MinTime:             time.Second * 10,
			PermitWithoutStream: true,
		}),
	}
}
