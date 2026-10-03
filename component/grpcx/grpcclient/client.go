package grpcclient

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"github.com/coffeehc/boot/component/grpcx/grpcquic"
	"github.com/coffeehc/boot/plugin/manage/metrics"
	"github.com/piotrkowalczuk/promgrpc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/resolver"
	"time"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/grpcrecovery"
	"github.com/coffeehc/boot/configuration"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/roundrobin"
	_ "google.golang.org/grpc/encoding/gzip"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
)

func EnableQuic(ctx context.Context, enable bool) context.Context {
	return context.WithValue(ctx, "_EnableQuic", enable)
}

func getEnableQuic(ctx context.Context) bool {
	v := ctx.Value("_EnableQuic")
	if v == nil {
		return false
	}
	return v.(bool)
}

var scope = zap.String("scope", "grpc.client")

// NewClientConnByServiceInfo 创建惰性连接；ctx 用于凭据装配，RPC 使用各自的调用 context。
// 调用方拥有返回连接，须在不再使用时 Close；返回错误表示目标或配置无效。
// dialOptions 在默认选项和 SetDialOptions 设置的选项之后应用，遵循 gRPC 的覆盖或叠加语义。
func NewClientConnByServiceInfo(ctx context.Context, serviceInfo configuration.ServiceInfo, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	if serviceInfo.TargetUrl == "" {
		return nil, errors.MessageError("没有设置TargetUrl")
	}
	opts := BuildDialOption(ctx, serviceInfo.ServiceName, dialOptions...)
	log.Debug("需要获取的客户端地址", zap.String("target", serviceInfo.TargetUrl))
	clientConn, err := grpc.NewClient(serviceInfo.TargetUrl, opts...)
	if err != nil {
		log.Error("创建服务端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}

// NewClientConnByResolverBuilder 使用连接级 resolver 创建惰性连接，调用方负责 Close。
// 额外连接选项通过 SetDialOptions 传入；resolverBuilders 保持原有可变参数形式。
func NewClientConnByResolverBuilder(ctx context.Context, serviceInfo configuration.ServiceInfo, resolverBuilders ...resolver.Builder) (*grpc.ClientConn, error) {
	if serviceInfo.TargetUrl == "" {
		return nil, errors.MessageError("没有设置TargetUrl")
	}
	opts := BuildDialOption(ctx, serviceInfo.ServiceName)
	opts = append(opts, grpc.WithResolvers(resolverBuilders...))
	clientConn, err := grpc.NewClient(serviceInfo.TargetUrl, opts...)
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	log.Debug("需要链接的服务端地址", zap.String("target", clientConn.Target()))
	return clientConn, nil
}

// NewClientConn 创建惰性连接，成功不表示服务已连接；RPC 使用各自的调用 context。
// serverAddr 必须非空，serverServiceName 用作指标标签；调用方负责 Close。
// dialOptions 在默认选项和 SetDialOptions 设置的选项之后应用，遵循 gRPC 的覆盖或叠加语义。
func NewClientConn(ctx context.Context, serverAddr string, serverServiceName string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	if serverAddr == "" {
		return nil, errors.MessageError("没有设置 gRPC 服务地址")
	}
	opts := BuildDialOption(ctx, serverServiceName, dialOptions...)
	clientConn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}

// BuildDialOption 组装 gRPC 客户端的连接选项，凭据须通过 context 或 dialOptions 显式设置。
// 默认不配置业务重试、等待就绪或压缩；调用方可追加原生 DialOption 或 CallOption。
// 应用顺序为默认选项、SetDialOptions 的选项、dialOptions；链式拦截器等仍按 gRPC 语义叠加。
func BuildDialOption(ctx context.Context, serverServiceName string, dialOptions ...grpc.DialOption) []grpc.DialOption {
	chainUnaryClient := []grpc.UnaryClientInterceptor{
		grpcrecovery.UnaryClientInterceptor(),
	}
	chainStreamClient := []grpc.StreamClientInterceptor{
		grpcrecovery.StreamClientInterceptor(),
	}
	// 全局重试无法确认写操作是否幂等，重试策略应由具体服务按方法配置。
	defaultServiceConfig := `{"loadBalancingConfig":[{"round_robin":{}}]}`
	csh := promgrpc.ClientStatsHandler(
		promgrpc.CollectorWithNamespace("grpc"),
		promgrpc.CollectorWithConstLabels(prometheus.Labels{"service": serverServiceName}),
	)
	if err := metrics.RegisterMetrics(csh); err != nil {
		var registered prometheus.AlreadyRegisteredError
		if stderrors.As(err, &registered) {
			if existing, ok := registered.ExistingCollector.(*promgrpc.StatsHandler); ok {
				csh = existing
			}
		} else {
			log.Error("注册 gRPC 客户端指标失败", zap.Error(err), scope)
		}
	}
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*8),
			grpc.MaxCallSendMsgSize(1024*1024*8),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Minute,
			Timeout:             20 * time.Second,
			PermitWithoutStream: false,
		}),
		grpc.WithDefaultServiceConfig(defaultServiceConfig),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithChainStreamInterceptor(chainStreamClient...),
		grpc.WithChainUnaryInterceptor(chainUnaryClient...),
	}
	if perRPCCredentials := GetPerRPCCredentials(ctx); perRPCCredentials != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(perRPCCredentials))
	}
	creds := GetClientCerts(ctx)
	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	enableQUiC := getEnableQuic(ctx)
	if enableQUiC {
		tlsConfig := &tls.Config{
			NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithContextDialer(grpcquic.NewQuicDialer(tlsConfig)))
	}
	opts = append(opts, grpc.WithStatsHandler(csh))
	contextOptions, _ := ctx.Value(dialOptionsContextKey{}).([]grpc.DialOption)
	opts = append(opts, contextOptions...)
	opts = append(opts, dialOptions...)
	return opts
}
