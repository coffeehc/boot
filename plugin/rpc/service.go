package rpc

import (
	"context"
	"sync"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "rpc"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetGRPCServer() *grpc.Server
	GetRPCServerAddr() string
	GetRegisterServiceId() string
}

func EnablePlugin(ctx context.Context) {
	if name == "" {
		log.Panic("插件名称没有初始化")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	service = newService(ctx)
	// reflection.Register(service.GetGRPCServer()) //是否开启远程控制
	log.Debug("初始化RPC服务", zap.String("rpcServerAddr", service.GetRPCServerAddr()))
	plugin.RegisterPlugin(name, service)
}
