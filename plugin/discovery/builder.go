package discovery

import (
	"context"
	"github.com/coffeehc/boot/plugin/discovery/ipsd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"time"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/grpcclient"
	"github.com/coffeehc/boot/configuration"
	"go.uber.org/zap"
)

func RPCServiceInitializationByResolverBuilder(ctx context.Context, rpcService configuration.RPCService, resolverBuilder ...resolver.Builder) error {
	opts := grpcclient.BuildDialOption(ctx, rpcService.GetRPCServiceInfo().ServiceName)
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	opts = append(opts, grpc.WithResolvers(resolverBuilder...))
	clientConn, err := grpc.NewClient(rpcService.GetRPCServiceInfo().TargetUrl, opts...)
	//clientConn, err := grpc.DialContext(ctx, rpcService.GetRPCServiceInfo().TargetUrl, opts...)
	// log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return errors.WrappedSystemError(err)
	}
	_err := rpcService.InitRPCService(ctx, clientConn)
	if _err != nil {
		return _err
	}
	log.Debug("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil
}

func RPCServiceInitializationByAddresses(ctx context.Context, rpcService configuration.RPCService, serverAddr ...string) (resolver.Builder, error) {
	resolverBuilder, err := ipsd.GetResolverBuilder(ctx, serverAddr...)
	if err != nil {
		log.Error("错误", zap.Error(err))
		return nil, err
	}
	conn, err := grpcclient.NewClientConnByResolverBuilder(ctx, rpcService.GetRPCServiceInfo(), resolverBuilder)
	if err != nil {
		log.Error("构建ResolverBuilder失败", zap.Error(err))
		return resolverBuilder, errors.ConverError(err)
	}
	log.Debug("需要链接的服务端地址", zap.Strings("target", serverAddr))
	_err := rpcService.InitRPCService(ctx, conn)
	if _err != nil {
		return resolverBuilder, _err
	}
	log.Debug("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return resolverBuilder, nil
}

func RPCServiceInitializationByAddress(ctx context.Context, rpcService configuration.RPCService, serverAddr string) error {
	conn, err := grpcclient.NewClientConn(ctx, serverAddr, rpcService.GetRPCServiceInfo().ServiceName)
	if err != nil {
		return errors.ConverError(err)
	}
	log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	_err := rpcService.InitRPCService(ctx, conn)
	if _err != nil {
		return _err
	}
	log.Debug("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil
}

func RPCServiceInitialization(ctx context.Context, rpcService configuration.RPCService) error {
	conn, err := grpcclient.NewClientConnByServiceInfo(ctx, rpcService.GetRPCServiceInfo())
	if err != nil {
		return err
	}
	err = rpcService.InitRPCService(ctx, conn)
	if err != nil {
		return err
	}
	log.Debug("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil
}
