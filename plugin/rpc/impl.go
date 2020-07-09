package rpc

import (
	"context"
	"net"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type serviceImpl struct {
	config        *RpcConfig
	server        *grpc.Server
	rpcServerAddr string
}

func (impl *serviceImpl) GetRPCServerAddr() string {
	if impl.server == nil {
		log.Fatal("rpc server没有初始化")
	}
	return impl.rpcServerAddr
}

func (impl *serviceImpl) GetGRPCServer() *grpc.Server {
	return impl.server
}

func (impl *serviceImpl) Start(ctx context.Context) errors.Error {
	addr, _ := net.ResolveTCPAddr("tcp4", impl.rpcServerAddr)
	addr.IP = net.IPv4zero
	lis, _err := net.Listen("tcp4", addr.String())
	if _err != nil {
		log.Fatal("启动RPC服务端口失败", zap.Error(_err))
	}
	go func() {
		err := impl.server.Serve(lis)
		if err != nil {
			log.Fatal("RPC服务异常关闭", zap.Error(err))
		}
	}()
	time.Sleep(time.Millisecond * 10)
	log.Debug("启动RPC服务", zap.String("rpcServerAddr", impl.rpcServerAddr))
	return nil
}
func (impl *serviceImpl) Stop(ctx context.Context) errors.Error {
	impl.server.Stop()
	log.Info("RPC服务关闭")
	return nil
}

//func RegisterRPCMotheds(register func(server *grpc.Server)) {
//	if server == nil {
//		log.Error("GRPC服务没有初始化,不能注册")
//	}
//	log.Debug("开始注册RPC接口方法")
//	register(server)
//}
