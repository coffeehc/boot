package grpcclient

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	_ "git.xiagaogao.com/coffee/boot/transport"
	"git.xiagaogao.com/coffee/boot/transport/grpcrecovery"
	"github.com/coreos/etcd/clientv3"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

type GRPCConnFactory interface {
	NewClientConn(cxt context.Context, serviceInfo *boot.ServiceInfo, block bool, defaultAddr ...string) (*grpc.ClientConn, errors.Error)
}

type grpcClientImpl struct {
	etcdClient   *clientv3.Client
	errorService errors.Service
	logger       *zap.Logger
}

func NewGRPCConnFactory(etcdClient *clientv3.Client, errorService errors.Service, logger *zap.Logger) GRPCConnFactory {
	errorService = errorService.NewService("grpc")
	return &grpcClientImpl{etcdClient: etcdClient, errorService: errorService, logger: logger}
}

func (impl *grpcClientImpl) NewClientConn(ctx context.Context, serviceInfo *boot.ServiceInfo, block bool, defaultAddr ...string) (*grpc.ClientConn, errors.Error) {
	ctx = boot.SetServiceName(ctx, serviceInfo.ServiceName)
	logger := impl.logger.WithOptions(zap.Fields(zap.String("rpc_t", serviceInfo.ServiceName)))
	chainUnaryClient := grpc_middleware.ChainUnaryClient(
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(impl.errorService, logger),
	)
	chainStreamClient := grpc_middleware.ChainStreamClient(
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(impl.errorService, logger),
	)
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithAuthority(boot.RunModel()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"), grpc.FailFast(true)),
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
	target := fmt.Sprintf("%s://%s/%s", etcdsd.MicorScheme, boot.RunModel(), serviceInfo.ServiceName)
	if resolver.Get(target) == nil {
		err := etcdsd.RegisterResolver(ctx, impl.etcdClient, serviceInfo, impl.errorService, impl.logger, defaultAddr...)
		if err != nil {
			return nil, err
		}
	}
	if block {
		opts = append(opts, grpc.WithBlock())
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, impl.errorService.WrappedSystemError(err)
	}
	return clientConn, nil
}

func NewClientConn(ctx context.Context, errorService errors.Service, logger *zap.Logger, serverAddr string) (*grpc.ClientConn, errors.Error) {
	errorService = errorService.NewService("grpc")
	chainUnaryClient := grpc_middleware.ChainUnaryClient(
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(errorService, logger),
	)
	chainStreamClient := grpc_middleware.ChainStreamClient(
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(errorService, logger),
	)
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithAuthority(boot.RunModel()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"), grpc.FailFast(true)),
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
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, serverAddr, opts...)
	if err != nil {
		return nil, errorService.WrappedSystemError(err)
	}
	return clientConn, nil
}
