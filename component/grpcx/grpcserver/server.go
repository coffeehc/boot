package grpcserver

import (
	"context"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin/manage/metrics"
	"github.com/piotrkowalczuk/promgrpc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"math"
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

func NewServer(ctx context.Context, grpcConfig *GRPCServerConfig) (*grpc.Server, error) {
	if grpcConfig == nil {
		grpcConfig = &GRPCServerConfig{}
	}
	if !viper.IsSet("grpc") {
		log.Warn("没有配置GRPCConfig,使用默认配置", scope)
	}
	err := viper.UnmarshalKey("grpc", grpcConfig)
	if err != nil {
		log.Panic("解析grpc配置失败", zap.Error(err), scope)
	}
	server := grpc.NewServer(BuildGRPCServerOptions(ctx, grpcConfig)...)
	return server, nil
}

func BuildGRPCServerOptions(ctx context.Context, config *GRPCServerConfig) []grpc.ServerOption {
	chainUnaryServers := make([]grpc.UnaryServerInterceptor, 0)
	if EnableAccessLog {
		log.Debug("开启GRPC访问日志")
		chainUnaryServers = append(chainUnaryServers, DebugLoggingInterceptor())
	}
	chainUnaryServers = append(chainUnaryServers, grpcrecovery.UnaryServerInterceptor())
	chainStreamServers := []grpc.StreamServerInterceptor{
		//grpc_prometheus.StreamServerInterceptor,
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
	if config.MaxConcurrentStreams == 0 {
		config.MaxConcurrentStreams = math.MaxUint32
	}
	ssh := promgrpc.ServerStatsHandler(
		promgrpc.CollectorWithNamespace("grpc"),
		promgrpc.CollectorWithConstLabels(prometheus.Labels{"service": configuration.GetServiceInfo().ServiceName}),
	)
	metrics.RegisterMetrics(ssh)
	opts := []grpc.ServerOption{
		grpc.StatsHandler(ssh),
		grpc.Creds(GetServerCerts(ctx)),
		grpc.InitialWindowSize(1024 * 1024 * 32),
		grpc.InitialConnWindowSize(1024 * 1024 * 4),
		grpc.ReadBufferSize(1024 * 128),
		grpc.WriteBufferSize(1024 * 128),
		grpc.MaxRecvMsgSize(1024 * 1024 * 8),
		grpc.MaxSendMsgSize(1024 * 1024 * 8),
		grpc.NumStreamWorkers(64),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.ChainStreamInterceptor(chainStreamServers...),
		grpc.ChainUnaryInterceptor(chainUnaryServers...),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: GetMaxConnectionIdle(), //time.Minute * 30,
			Timeout:           60 * time.Second,       // 類似 ClientParameters.Time 不過默認爲 2小時
			Time:              30 * time.Second,       // 類似 ClientParameters.Timeout 默認 20秒
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{ // 當服務器不允許ping 或 ping 太頻繁超過 MinTime 限制 服務器 會 返回ping失敗 此時 客戶端 不會認爲這個ping是 active RPCs
			MinTime:             time.Second * 5,
			PermitWithoutStream: true,
		}),
		grpc.SharedWriteBuffer(true),
	}
	if config.MaxMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(config.MaxMsgSize),
			grpc.MaxSendMsgSize(config.MaxMsgSize))
	}
	return opts
}

// //metadata.FromIncomingContext(ctx)
