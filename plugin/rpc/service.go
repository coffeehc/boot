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

// Service 提供 RPC 服务注册入口和监听地址；生命周期由 plugin 管理。
// 业务服务须在插件启动前注册，地址支持与启动流程并发读取。
type Service interface {
	// GetGRPCServer 返回业务服务注册入口，不得在 Start 后继续注册。
	GetGRPCServer() *grpc.Server
	// GetRPCServerAddr 返回监听地址，端口 0 在 Start 成功后替换为实际端口。
	GetRPCServerAddr() string
	// GetRegisterServiceId 返回服务名与监听 IP 构成的注册标识。
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
