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
			Timeout:           20 * time.Second,
			Time:              2 * time.Hour,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Minute * 5,
			PermitWithoutStream: false,
		}),
	}
}
