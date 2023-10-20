package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/coffeehc/boot/component/grpc/grpcquic"
	"github.com/coffeehc/boot/plugin/manage/metrics"
	"github.com/piotrkowalczuk/promgrpc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"time"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpc/grpcrecovery"
	"github.com/coffeehc/boot/configuration"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
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

func NewClientConnByServiceInfo(ctx context.Context, serviceInfo configuration.ServiceInfo, block bool) (*grpc.ClientConn, error) {
	opts := BuildDialOption(ctx, block, serviceInfo.ServiceName)
	if serviceInfo.TargetUrl == "" {
		log.Panic("没有指定需要链接的ServiceInfo的RPC协议，无法创建链接")
	}
	log.Debug("需要获取的客户端地址", zap.String("target", serviceInfo.TargetUrl))
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, serviceInfo.TargetUrl, opts...)
	if err != nil {
		log.Error("创建服务端链接失败", zap.Error(err))
		return nil, errors.SystemError("创建grpc客户端")
	}
	return clientConn, nil
}

func NewClientConnByResolverBuilder(ctx context.Context, serviceInfo configuration.ServiceInfo, resolverBuilders ...resolver.Builder) (*grpc.ClientConn, error) {
	if serviceInfo.TargetUrl == "" {
		return nil, errors.MessageError("没有设置TargetUrl")
	}
	opts := BuildDialOption(ctx, false, serviceInfo.ServiceName)
	opts = append(opts, grpc.WithResolvers(resolverBuilders...))
	clientConn, err := grpc.DialContext(ctx, serviceInfo.TargetUrl, opts...)
	// log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}

func NewClientConn(ctx context.Context, block bool, serverAddr string, serverServiceName string) (*grpc.ClientConn, error) {
	opts := BuildDialOption(ctx, block, serverServiceName)
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, serverAddr, opts...)
	// log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}

func BuildDialOption(ctx context.Context, block bool, serverServiceName string) []grpc.DialOption {
	chainUnaryClient := []grpc.UnaryClientInterceptor{
		grpcrecovery.UnaryClientInterceptor(),
	}
	chainStreamClient := []grpc.StreamClientInterceptor{
		grpcrecovery.StreamClientInterceptor(),
	}
	defaultServiceConfig := `{
	   "LoadBalancingPolicy": "round_robin",
		"HealthCheckConfig":{
			"ServiceName":"%s"
		},
		"methodConfig": [{
		  "name": [{}],
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE","UNKNOWN","ABORTED" ]
		  }
		}]}`
	defaultServiceConfig = fmt.Sprintf(defaultServiceConfig, serverServiceName)
	defaultServiceConfig = `{
	   "LoadBalancingPolicy": "round_robin",
		"methodConfig": [{
		  "name": [{}],
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE","UNKNOWN","ABORTED" ]
		  }
		}]}`
	//defaultServiceConfig = `{"LoadBalancingPolicy": "round_robin"}`
	csh := promgrpc.ClientStatsHandler(
		promgrpc.CollectorWithNamespace("grpc"),
		promgrpc.CollectorWithConstLabels(prometheus.Labels{"service": serverServiceName}),
	)
	metrics.RegisterMetrics(csh)
	opts := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Millisecond * 300, // 第一次失败重试前后需等待多久
				Multiplier: 1.2,                    // 在失败的重试后乘以的倍数
				Jitter:     0.2,                    // 随机抖动因子
				MaxDelay:   time.Second * 5,        // backoff上限
			},
			MinConnectTimeout: time.Second * 3,
		}),
		grpc.WithAuthority(configuration.GetRunModel()),
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor("gzip"),
			grpc.WaitForReady(true),
			grpc.MaxCallRecvMsgSize(1024*1024*8),
			grpc.MaxCallSendMsgSize(1024*1024*2),
		),
		grpc.WithReturnConnectionError(),
		grpc.WithIdleTimeout(0),
		//grpc.WithDisableRetry(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			Timeout:             time.Second * 5,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultServiceConfig(defaultServiceConfig),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithChainStreamInterceptor(chainStreamClient...),
		grpc.WithChainUnaryInterceptor(chainUnaryClient...),
		grpc.WithInitialConnWindowSize(1024 * 1024 * 8),
		grpc.WithInitialWindowSize(1024 * 1024 * 16),
		//grpc.WithChannelzParentID(&channelz.Identifier{}),
		grpc.FailOnNonTempDialError(true),
		grpc.WithNoProxy(),
		grpc.WithReadBufferSize(1024 * 128),
		grpc.WithWriteBufferSize(1024 * 128),
	}
	perRPCCredentials := ctx.Value(perRPCCredentialsKey)
	if perRPCCredentials != nil {
		if prc, ok := perRPCCredentials.(credentials.PerRPCCredentials); ok {
			opts = append(opts, grpc.WithPerRPCCredentials(prc))
		}
	}
	tlsConfig := &tls.Config{
		NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
		InsecureSkipVerify: true,
	}
	enableQUiC := getEnableQuic(ctx)
	if !enableQUiC {
		creds := getCerts(ctx)
		if creds == nil {
			creds = insecure.NewCredentials()
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		creds := grpcquic.NewCredentials(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
		opts = append(opts, grpc.WithContextDialer(grpcquic.NewQuicDialer(tlsConfig)))
	}
	if block {
		opts = append(opts, grpc.WithBlock())
	}
	opts = append(opts, grpc.WithStatsHandler(csh))
	return opts
}
