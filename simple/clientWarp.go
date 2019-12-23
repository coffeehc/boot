package simple

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RPCService struct {
	client       simplemodel.GreeterClient
	errorService xerror.Service
	logger       *zap.Logger
}

func (impl *RPCService) GetRPCServiceInfo() boot.ServiceInfo {
	return boot.ServiceInfo{
		ServiceName: "simple_service",
		Version:     "0.0.1",
		Descriptor:  "测试Server",
		Scheme:      "http",
	}
}
func (impl *RPCService) InitRPCService(ctx context.Context, grpcConn *grpc.ClientConn, errorService xerror.Service, logger *zap.Logger) xerror.Error {
	client := simplemodel.NewGreeterClient(grpcConn)
	impl.logger = logger
	impl.errorService = errorService
	impl.client = client
	return nil
}

func (impl *RPCService) GetClient() simplemodel.GreeterClient {
	return impl.client
}
