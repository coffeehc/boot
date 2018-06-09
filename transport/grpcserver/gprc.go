package grpcserver

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	_ "git.xiagaogao.com/coffee/boot/transport"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/tap"
)

func NewServer(ctx context.Context, grpcConfig *GRPCConfig, serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) (*grpc.Server, errors.Error) {
	//logger := logs.GetLogger(ctx)
	//logService := logs.GetLoggerService(ctx)
	grpcConfig.initGRPCConfig()
	server := grpc.NewServer(BuildGRPCOptions(ctx, grpcConfig, serviceInfo, errorService, logger)...)
	return server, nil
}

func BuildGRPCOptions(ctx context.Context, config *GRPCConfig, serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) []grpc.ServerOption {
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout))
	unaryServerInterceptor := newUnaryServerInterceptor(ctx, errorService, logger)
	grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"))
	//初始化Server
	grpc.EnableTracing = false
	if boot.IsDevModule() {
		grpc.EnableTracing = true
		unaryServerInterceptor.AppendInterceptor("logger", loggingInterceptor)
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
		grpc.InTapHandle(func(ctx1 context.Context, info *tap.Info) (context.Context, error) {
			ctx1 = logs.SetLogger(ctx1, logger)
			ctx1 = boot.SetServiceName(ctx1, serviceInfo.ServiceName)
			return ctx, nil
		}),
	}
}
