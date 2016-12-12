package grpcclient

import (
	"context"
	"crypto/tls"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"time"
)

const err_scope_grpcClient = "grpcClient"

func init() {
	grpclog.SetLogger(&grpcbase.GrpcLogger{})
}

type GrpcClient interface {
	NewClientConn(cxt context.Context, serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer, timeout time.Duration) (*grpc.ClientConn, base.Error)
}

type _GrpcClient struct {
}

func NewGrpcClient() GrpcClient {
	return &_GrpcClient{}
}

func (this *_GrpcClient) NewClientConn(cxt context.Context, serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer, timeout time.Duration) (*grpc.ClientConn, base.Error) {
	opts := []grpc.DialOption{
		//grpc.WithBackoffConfig(grpc.BackoffConfig{
		//	MaxDelay: time.Second,
		//}),
		grpc.WithBalancer(BalancerWapper(balancer)),
		//grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
	}
	if timeout > 0 {
		opts = append(opts, grpc.WithTimeout(timeout))
	}
	clientConn, err := grpc.DialContext(cxt, serviceInfo.GetServiceName(), opts...)
	if err != nil {
		return nil, base.NewErrorWrapper(err_scope_grpcClient, err)
	}
	return clientConn, nil
}
