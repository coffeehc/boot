package grpcclient

import (
	"crypto/tls"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
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
		grpc.WithTimeout(time.Second * 3),
		grpc.WithBackoffConfig(grpc.BackoffConfig{
			MaxDelay: 5 * time.Second,
		}),
		grpc.WithBalancer(BalancerWapper(balancer)),
		//grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
	}

	clientConn, err := grpc.Dial(serviceInfo.GetServiceName(), opts...)
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	return clientConn, nil
}
