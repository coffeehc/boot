package grpcclient

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"google.golang.org/grpc"
)

type GrpcClient interface {
	NewClientConn(serviceInfo base.ServiceInfo) (*grpc.ClientConn, base.Error)
}

type _GrpcClient struct {
}

func (this *_GrpcClient) NewClientConn(serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer) (*grpc.ClientConn, base.Error) {
	dialOption := grpc.WithBalancer(BalancerWapper(balancer))
	clientConn, err := grpc.Dial(serviceInfo.GetServiceName(), dialOption)
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	return clientConn, nil
}
