package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/transport/grpcclient"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RPCService interface {
	GetRPCServiceInfo() boot.ServiceInfo
	InitRPCService(ctx context.Context, grpcConn *grpc.ClientConn, errorService errors.Service, logger *zap.Logger) errors.Error
}

type RPCServiceInitialization func(RPCService) errors.Error

func newRPCServiceInitialization(ctx context.Context, grpcConnFactory grpcclient.GRPCConnFactory, errorService errors.Service, logger *zap.Logger) RPCServiceInitialization {
	errorService = errorService.NewService("rpc")
	return func(service RPCService) errors.Error {
		conn, err := grpcConnFactory.NewClientConn(ctx, service.GetRPCServiceInfo(), false)
		if err != nil {
			return err
		}
		return service.InitRPCService(ctx, conn, errorService, logger)
	}
}
