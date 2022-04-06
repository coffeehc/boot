package discovery

import (
	"context"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpc/grpcclient"
	"github.com/coffeehc/boot/configuration"
	"go.uber.org/zap"
)

func RPCServiceInitializationByAddress(ctx context.Context, rpcService configuration.RPCService, serverAddr string) error {
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

func RPCServiceInitialization(ctx context.Context, rpcService configuration.RPCService) error {
	conn, err := grpcclient.NewClientConnByServiceInfo(ctx, rpcService.GetRPCServiceInfo(), false)
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
