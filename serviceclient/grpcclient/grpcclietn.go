package grpcclient

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const errScopeGRPCClient = "grpcClient"

type GRPCClient interface {
	NewClientConn(cxt context.Context, serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer, timeout time.Duration) (*grpc.ClientConn, base.Error)
}

type _GRPCClient struct {
}

func NewGRPCClient() GRPCClient {
	return &_GRPCClient{}
}

func (client *_GRPCClient) NewClientConn(cxt context.Context, serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer, timeout time.Duration) (*grpc.ClientConn, base.Error) {
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 20, PermitWithoutStream: true}),
		grpc.WithBalancer(adopterToGRPCBalancer(balancer)),
		grpc.WithUserAgent("coffee's grpc client"),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithUnaryInterceptor(_unaryClientInterceptor.Interceptor),
	}
	if timeout > 0 {
		opts = append(opts, grpc.WithTimeout(timeout))
	}
	clientConn, err := grpc.DialContext(cxt, serviceInfo.GetServiceName(), opts...)
	if err != nil {
		return nil, base.NewErrorWrapper(errScopeGRPCClient, 0, err)
	}
	return clientConn, nil
}
