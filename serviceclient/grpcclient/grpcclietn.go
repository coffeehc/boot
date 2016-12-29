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
)

const errScopeGRPCClient = "grpcClient"

type _GRPCClient struct {
}

func newGRPCClient() *_GRPCClient {
	return &_GRPCClient{}
}

func (client *_GRPCClient) newClientConn(cxt context.Context, serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer, timeout time.Duration) (*grpc.ClientConn, base.Error) {
	opts := []grpc.DialOption{
		//grpc.WithBackoffConfig(grpc.BackoffConfig{
		//	MaxDelay: time.Second,
		//}),
		grpc.WithBalancer(adopterToGRPCBalancer(balancer)),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return nil, &reconnectionError{err: err}
			}
			return conn, nil
		}), //这个非常重要,用于连接重试,否则很大概率在网络抖动或依赖服务重启的时候,试一次不同就再也不尝试,变成一个死链接
		grpc.WithUnaryInterceptor(_unaryClientInterceptor.Interceptor),
	}
	if timeout > 0 {
		opts = append(opts, grpc.WithTimeout(timeout))
	}
	clientConn, err := grpc.DialContext(cxt, serviceInfo.GetServiceName(), opts...)
	if err != nil {
		return nil, base.NewErrorWrapper(errScopeGRPCClient, err)
	}
	return clientConn, nil
}
