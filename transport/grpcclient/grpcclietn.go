package grpcclient

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	_ "git.xiagaogao.com/coffee/boot/transport"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

const (
	errScopeGRPCClient = "grpcClient"
)

var logger *zap.Logger

type GRPCConnFactory interface {
	NewClientConn(cxt context.Context, serviceInfo boot.ServiceInfo, block bool) (*grpc.ClientConn, errors.Error)
}

type grpcClientImpl struct {
	options      []grpc.DialOption
	etcdClient   *clientv3.Client
	errorService errors.Service
}

func NewGRPCConnFactory(cxt context.Context, etcdClient *clientv3.Client, serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) GRPCConnFactory {
	grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"))
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 10, PermitWithoutStream: true}),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(wapperUnartClientInterceptor(serviceInfo, errorService, logger)),
	}
	return &grpcClientImpl{options: opts, etcdClient: etcdClient, errorService: errorService}
}

func (impl *grpcClientImpl) NewClientConn(ctx context.Context, serviceInfo boot.ServiceInfo, block bool) (*grpc.ClientConn, errors.Error) {
	target := fmt.Sprintf("%s://%s/%s", etcdsd.MicorScheme, boot.RunModule(), serviceInfo.ServiceName)
	if resolver.Get(target) == nil {
		err := etcdsd.RegisterResolver(ctx, impl.etcdClient, serviceInfo)
		if err != nil {
			return nil, err
		}
	}
	opts := impl.options
	if block {
		opts = append(impl.options, grpc.WithBlock())
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*20)
	clientConn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, impl.errorService.WappedSystemError(err)
	}
	return clientConn, nil
}
