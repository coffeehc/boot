package register

import (
	"context"
	"sync"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"github.com/coffeehc/boot/plugin/manage"
	"github.com/coffeehc/boot/plugin/register/internal"
	"github.com/coffeehc/boot/plugin/rpc"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "serviceRegister"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
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
	manage.EnablePlugin(ctx)
	internal.EnablePlugin(ctx)
	impl := &serviceImpl{
		registerManage: internal.GetService(),
	}
	service = impl
	plugin.RegisterPlugin(name, impl)
}
