package grpcclient

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type ClientConnFactory interface {
	GetClientConn(cxt context.Context, serviceInfo base.ServiceInfo, timeout time.Duration) (*grpc.ClientConn, base.Error)
}

func NewClientConnFactory(grpcClient GrpcClient, builder loadbalancer.BalancerBuilder) ClientConnFactory {
	if grpcClient == nil {
		grpcClient = NewGrpcClient()
	}
	return &_ServiceClient{
		grpcClient: grpcClient,
		builder:    builder,
	}
}

type _ServiceClient struct {
	grpcClient GrpcClient
	builder    loadbalancer.BalancerBuilder
}

func (this *_ServiceClient) GetClientConn(cxt context.Context, serviceInfo base.ServiceInfo, timeout time.Duration) (*grpc.ClientConn, base.Error) {
	balancer, err := this.builder.NewBalancer(cxt, serviceInfo)
	if err != nil {
		return nil, err
	}
	clientConn, err := this.grpcClient.NewClientConn(cxt, serviceInfo, balancer, timeout)
	return clientConn, err
}
