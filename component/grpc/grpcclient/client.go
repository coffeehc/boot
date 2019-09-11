package grpcclient

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/component/etcdsd"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcrecovery"
	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

var scope = zap.String("scope", "grpc.client")

func NewClientConnByRegister(ctx context.Context, serviceInfo configuration.ServiceInfo, block bool, defaultAddr ...string) (*grpc.ClientConn, errors.Error) {
	// ctx = boot.SetServiceName(ctx, serviceInfo.ServiceName)
	// logger := impl.logger.WithOptions(zap.Fields(zap.String("rpc.service", serviceInfo.ServiceName)))
	chainUnaryClient := grpc_middleware.ChainUnaryClient(
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(),
	)
	chainStreamClient := grpc_middleware.ChainStreamClient(
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(),
	)
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithAuthority(configuration.GetModel()),
		// grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"), grpc.FailFast(true)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 10, PermitWithoutStream: false}),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(chainUnaryClient),
		grpc.WithStreamInterceptor(chainStreamClient),
		grpc.WithInitialConnWindowSize(10),
		grpc.WithInitialWindowSize(1024),
		grpc.WithChannelzParentID(0),
		grpc.FailOnNonTempDialError(true),
	}
	if serviceInfo.Scheme == "" {
		log.Fatal("没有指定需要链接的ServiceInfo的RPC协议，无法创建链接")
	}
	target := fmt.Sprintf("%s://%s/%s", serviceInfo.Scheme, configuration.GetModel(), serviceInfo.ServiceName)
	log.Debug("需要获取的客户端地址", zap.String("target", target))
	if resolver.Get(serviceInfo.Scheme) == nil {
		switch serviceInfo.Scheme {
		case configuration.MicroServiceProtocolScheme:
			err := etcdsd.Resolver(ctx)
			if err != nil {
				return nil, err
			}
		default:
			log.Fatal("不能识别的协议", zap.String("scheme", serviceInfo.Scheme))
		}
	}
	if block {
		opts = append(opts, grpc.WithBlock())
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
	chainUnaryClient := grpc_middleware.ChainUnaryClient(
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(),
	)
	chainStreamClient := grpc_middleware.ChainStreamClient(
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(),
	)
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithAuthority(configuration.GetModel()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name), grpc.FailFast(true)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 15, PermitWithoutStream: true}), //20秒发送一个keepalive
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(chainUnaryClient),
		grpc.WithStreamInterceptor(chainStreamClient),
		grpc.WithInitialConnWindowSize(10),
		grpc.WithInitialWindowSize(1024),
		grpc.WithChannelzParentID(0),
		grpc.FailOnNonTempDialError(true),
	}
	if block {
		opts = append(opts, grpc.WithBlock())
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, serverAddr, opts...)
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}
