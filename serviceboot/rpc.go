package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/base/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RPCService interface {
	GetRPCServiceInfo() *boot.ServiceInfo
	InitRPCService(ctx context.Context, grpcConn *grpc.ClientConn, errorService xerror.Service, logger *zap.Logger) xerror.Error
}
