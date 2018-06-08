package grpcclient

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
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
	options    []grpc.DialOption
	etcdClient *clientv3.Client
}

func NewGRPCConnFactory(cxt context.Context, etcdClient *clientv3.Client, serviceInfo boot.ServiceInfo) GRPCConnFactory {
	logger = logs.GetLogger(cxt)
	//grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"))
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 10, PermitWithoutStream: true}),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithUserAgent("coffee's grpcserver client"),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(wapperUnartClientInterceptor(serviceInfo)),
		//grpc.NewGZIPDecompressor()
	}
	return &grpcClientImpl{options: opts, etcdClient: etcdClient}
}

func (client *grpcClientImpl) NewClientConn(ctx context.Context, serviceInfo boot.ServiceInfo, block bool) (*grpc.ClientConn, errors.Error) {
	target := fmt.Sprintf("%s://%s/%s", etcdsd.MicorScheme, serviceInfo.GetServiceTag(), serviceInfo.GetServiceName())
	if resolver.Get(target) == nil {
		err := etcdsd.RegisterResolver(ctx, client.etcdClient, serviceInfo)
		if err != nil {
			return nil, err
		}
	}
	opts := client.options
	if block {
		opts = append(client.options, grpc.WithBlock())
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*20)
	clientConn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, errors.NewErrorWrapper(errors.Error_System, errScopeGRPCClient+"."+serviceInfo.GetServiceName(), err)
	}
	return clientConn, nil
}
