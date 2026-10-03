package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin/manage/metrics"
	"github.com/piotrkowalczuk/promgrpc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/grpcrecovery"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

var scope = zap.String("scope", "grpc.server")

// NewServer 创建尚未监听端口的 gRPC server；显式配置优先，nil 时读取 grpc 配置段。
// 配置解析和校验失败返回错误，调用方负责注册服务、Serve 和关闭 server。
func NewServer(ctx context.Context, grpcConfig *GRPCServerConfig) (*grpc.Server, error) {
	if grpcConfig == nil {
		grpcConfig = &GRPCServerConfig{}
		if err := viper.UnmarshalKey("grpc", grpcConfig); err != nil {
			return nil, fmt.Errorf("解析 grpc 配置失败: %w", err)
		}
	}
	if grpcConfig.MaxMsgSize < 0 {
		return nil, fmt.Errorf("grpc.max_msg_size 不能为负数")
	}
	if GetMaxConnectionIdle() < 0 {
		return nil, fmt.Errorf("grpc.max_connection_idle 不能为负数")
	}
	server := grpc.NewServer(BuildGRPCServerOptions(ctx, grpcConfig)...)
	return server, nil
}

// BuildGRPCServerOptions 组装拦截器、消息限制、心跳和指标，不修改传入配置。
// config 可为 nil；调用方可追加原生 ServerOption 覆盖默认值。
func BuildGRPCServerOptions(ctx context.Context, config *GRPCServerConfig) []grpc.ServerOption {
	chainUnaryServers := make([]grpc.UnaryServerInterceptor, 0)
	if EnableAccessLog {
		log.Debug("开启GRPC访问日志")
		chainUnaryServers = append(chainUnaryServers, DebugLoggingInterceptor())
	}
	chainUnaryServers = append(chainUnaryServers, grpcrecovery.UnaryServerInterceptor())
	chainStreamServers := []grpc.StreamServerInterceptor{
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
	maxMsgSize := 8 * 1024 * 1024
	maxConcurrentStreams := uint32(100)
	if config != nil {
		if config.MaxMsgSize > 0 {
			maxMsgSize = config.MaxMsgSize
		}
		if config.MaxConcurrentStreams > 0 {
			maxConcurrentStreams = config.MaxConcurrentStreams
		}
	}
	ssh := promgrpc.ServerStatsHandler(
		promgrpc.CollectorWithNamespace("grpc"),
		promgrpc.CollectorWithConstLabels(prometheus.Labels{"service": configuration.GetServiceInfo().ServiceName}),
	)
	if err := metrics.RegisterMetrics(ssh); err != nil {
		var registered prometheus.AlreadyRegisteredError
		if errors.As(err, &registered) {
			if existing, ok := registered.ExistingCollector.(*promgrpc.StatsHandler); ok {
				ssh = existing
			}
		} else {
			log.Error("注册 gRPC 服务端指标失败", zap.Error(err), scope)
		}
	}
	opts := []grpc.ServerOption{
		grpc.StatsHandler(ssh),
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.MaxConcurrentStreams(maxConcurrentStreams),
		grpc.ChainStreamInterceptor(chainStreamServers...),
		grpc.ChainUnaryInterceptor(chainUnaryServers...),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: GetMaxConnectionIdle(),
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Minute,
			PermitWithoutStream: false,
		}),
		grpc.SharedWriteBuffer(true),
	}
	if creds := GetServerCerts(ctx); creds != nil {
		opts = append(opts, grpc.Creds(creds))
	}
	return opts
}
