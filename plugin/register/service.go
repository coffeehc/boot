package register

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/plugin/manage"
	"git.xiagaogao.com/coffee/boot/plugin/register/internal"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "serviceRegister"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Fatal("Service没有初始化", scope)
	}
	return service
}

type Service interface {
}

func EnablePlugin(ctx context.Context) {
	if name == "" {
		log.Fatal("插件名称没有初始化")
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
