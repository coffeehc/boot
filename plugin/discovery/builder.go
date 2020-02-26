package discovery

import (
	"context"
	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcclient"
	"git.xiagaogao.com/coffee/boot/configuration"
	"go.uber.org/zap"
)

func RPCServiceInitializationByAddress(ctx context.Context, rpcService configuration.RPCService, serverAddr string) errors.Error {
	conn, err := grpcclient.NewClientConn(ctx, false, serverAddr)
	if err != nil {
		return errors.ConverError(err)
	}
	log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	_err := rpcService.InitRPCService(ctx, conn)
	if _err != nil {
		return _err
	}
	log.Info("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil

}

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
