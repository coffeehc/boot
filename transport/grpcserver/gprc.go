package grpcserver

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	_ "git.xiagaogao.com/coffee/boot/transport"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func NewServer(ctx context.Context, grpcConfig *GRPCConfig, serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) (*grpc.Server, errors.Error) {
	if grpcConfig == nil {
		logger.Debug("没有配置GRPCConfig,使用默认配置")
		grpcConfig = &GRPCConfig{}
	}
	grpcConfig.initGRPCConfig()
	server := grpc.NewServer(BuildGRPCOptions(ctx, grpcConfig, serviceInfo, errorService, logger)...)
	return server, nil
}

func BuildGRPCOptions(ctx context.Context, config *GRPCConfig, serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) []grpc.ServerOption {
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout))
	unaryServerInterceptor := newUnaryServerInterceptor(ctx, errorService, logger)
	//初始化Server
	grpc.EnableTracing = false
	if boot.IsDevModule() {
		grpc.EnableTracing = true
		unaryServerInterceptor.AppendInterceptor("logger", BuildLoggingInterceptor(errorService, logger))
	}
	unaryServerInterceptor.AppendInterceptor("prometheus", grpc_prometheus.UnaryServerInterceptor)
	return []grpc.ServerOption{
		grpc.InitialWindowSize(4096),
		grpc.InitialConnWindowSize(100),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.MaxRecvMsgSize(config.MaxMsgSize),
		grpc.MaxSendMsgSize(config.MaxMsgSize),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(unaryServerInterceptor.Interceptor),
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
