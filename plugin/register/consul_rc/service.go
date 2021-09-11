package consul_rc

import (
	"context"
	"sync"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"github.com/coffeehc/boot/plugin/register/internal"
	"github.com/coffeehc/boot/plugin/rpc"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "consul_registercenter"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	internal.RegisterCenter
	CheckDeregister(checkId string)
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
	rpc.EnablePlugin(ctx)
	internal.EnablePlugin(ctx)
	service = newService()
	err := internal.GetService().SetRegisterCenter(service)
	if err != nil {
		log.Panic("添加注册中心失败", err.GetFieldsWithCause()...)
	}
	plugin.RegisterPluginByFast(name, nil, nil)
}
