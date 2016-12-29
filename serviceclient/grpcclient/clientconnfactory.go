package grpcclient

import (
	"time"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//ClientConnFactory client connection factory
type ClientConnFactory interface {
	GetClientConn(cxt context.Context, serviceInfo base.ServiceInfo, timeout time.Duration) (*grpc.ClientConn, base.Error)
}

//NewClientConnFactory get the ClientConnFactory's instance
func NewClientConnFactory(builder loadbalancer.BalancerBuilder) ClientConnFactory {
	return &_ClientConnFactory{
		grpcClient: newGRPCClient(),
		builder:    builder,
	}
}

type _ClientConnFactory struct {
	grpcClient *_GRPCClient
	builder    loadbalancer.BalancerBuilder
}

func (factory *_ClientConnFactory) GetClientConn(cxt context.Context, serviceInfo base.ServiceInfo, timeout time.Duration) (*grpc.ClientConn, base.Error) {
	balancer, err := factory.builder.NewBalancer(cxt, serviceInfo)
	if err != nil {
		return nil, err
	}
	clientConn, err := factory.grpcClient.newClientConn(cxt, serviceInfo, balancer, timeout)
	return clientConn, err
}
