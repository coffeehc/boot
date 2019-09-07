package discovery

import (
	"context"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcclient"
	"git.xiagaogao.com/coffee/boot/configuration"
	"go.uber.org/zap"
)

func RPCServiceInitialization(ctx context.Context, rpcService configuration.RPCService) errors.Error {
	conn, err := grpcclient.NewClientConnByRegister(ctx, rpcService.GetRPCServiceInfo(), false)
	if err != nil {
		return err
	}
	err = rpcService.InitRPCService(ctx, conn)
	if err != nil {
		return err
	}
	log.Info("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil
}
