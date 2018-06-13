package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RPCService interface {
	GetRPCServiceInfo() boot.ServiceInfo
	InitRPCService(ctx context.Context, grpcConn *grpc.ClientConn, errorService errors.Service, logger *zap.Logger) errors.Error
}
