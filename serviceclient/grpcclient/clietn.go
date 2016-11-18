package grpcclient

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
	"time"
)

func init() {
	grpclog.SetLogger(&grpcbase.GrpcLogger{})
}

type GrpcClient interface {
	NewClientConn(serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer) (*grpc.ClientConn, base.Error)
}

type _GrpcClient struct {
}

func NewGrpcClient() GrpcClient {
	return &_GrpcClient{}
}

func (this *_GrpcClient) NewClientConn(serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer) (*grpc.ClientConn, base.Error) {
	opts := []grpc.DialOption{
		grpc.WithBalancer(BalancerWapper(balancer)),
		//grpc.WithInsecure(),
		//grpc.WithBlock(),
		grpc.WithDialer(func(addr string, t time.Duration) (net.Conn, error) {
			logger.Debug("addr is %s", addr)
			return nil, nil
		}),
	}

	clientConn, err := grpc.Dial(serviceInfo.GetServiceName(), opts...)
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	return clientConn, nil
}
