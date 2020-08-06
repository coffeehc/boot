package grpcclient

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/component/grpc/grpcrecovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

var scope = zap.String("scope", "grpc.client")

func NewClientConnByRegister(ctx context.Context, serviceInfo configuration.ServiceInfo, resolverBuilder resolver.Builder, block bool) (*grpc.ClientConn, errors.Error) {
	opts := BuildDialOption(ctx, block)
	if serviceInfo.Scheme == "" {
		log.Fatal("没有指定需要链接的ServiceInfo的RPC协议，无法创建链接")
	}
	target := fmt.Sprintf("%s://%s/%s", serviceInfo.Scheme, configuration.GetRunModel(), serviceInfo.ServiceName)
	log.Debug("需要获取的客户端地址", zap.String("target", target))
	if resolver.Get(serviceInfo.Scheme) == nil {
		resolver.Register(resolverBuilder)
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		log.Error("创建服务端链接失败", zap.Error(err))
		return nil, errors.SystemError("创建grpc客户端")
	}
	return clientConn, nil
}

func NewClientConn(ctx context.Context, block bool, serverAddr string) (*grpc.ClientConn, errors.Error) {
	opts := BuildDialOption(ctx, block)
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, serverAddr, opts...)
	// log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}

func BuildDialOption(ctx context.Context, block bool) []grpc.DialOption {
	chainUnaryClient := []grpc.UnaryClientInterceptor{
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(),
	}
	chainStreamClient := []grpc.StreamClientInterceptor{
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(),
	}
	opts := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Second,      // 第一次失败重试前后需等待多久
				Multiplier: 1.5,              // 在失败的重试后乘以的倍数
				Jitter:     0.2,              // 随机抖动因子
				MaxDelay:   time.Second * 30, // backoff上限
			},
			MinConnectTimeout: time.Second,
		}),
		grpc.WithAuthority(configuration.GetRunModel()),
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor("gzip"),
			grpc.WaitForReady(false),
			// grpc.MaxCallRecvMsgSize(1024*32),
			// grpc.MaxCallSendMsgSize(1024*32),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 3,
			Timeout:             time.Second * 10,
			PermitWithoutStream: false,
		}),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithChainStreamInterceptor(chainStreamClient...),
		grpc.WithChainUnaryInterceptor(chainUnaryClient...),
		grpc.WithInitialConnWindowSize(1024 * 64),
		grpc.WithInitialWindowSize(1024 * 256),
		grpc.WithChannelzParentID(0),
		grpc.FailOnNonTempDialError(false),
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
	creds := getCerts(ctx)
	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	if block {
		opts = append(opts, grpc.WithBlock())
	}
	return opts
}
