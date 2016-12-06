package grpcclient

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type ClientConnFactory interface {
	GetClientConn(serviceInfo base.ServiceInfo, timeout time.Duration) (*grpc.ClientConn, <-chan struct{}, base.Error)
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

func (this *_ServiceClient) GetClientConn(serviceInfo base.ServiceInfo, timeout time.Duration) (*grpc.ClientConn, <-chan struct{}, base.Error) {
	balancer, err := this.builder.NewBalancer(serviceInfo)
	if err != nil {
		return nil, nil, err
	}
	cxt := context.Background()
	clientConn, err := this.grpcClient.NewClientConn(cxt, serviceInfo, balancer, timeout)
	return clientConn, cxt.Done(), err
}
