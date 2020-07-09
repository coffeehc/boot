package consul_rc

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/consul"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/plugin/register/internal"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "consul_registercenter"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Fatal("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	internal.RegisterCenter
	//CheckRegister(ctx context.Context)
	//CheckDeregister(checkId string)
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
	consul.EnablePlugin(ctx)
	rpc.EnablePlugin(ctx)
	internal.EnablePlugin(ctx)
	service = newService()
	err := internal.GetService().SetRegisterCenter(service)
	if err != nil {
		log.Fatal("添加注册中心失败", err.GetFieldsWithCause()...)
	}
	plugin.RegisterPluginByFast(name, nil, nil)
}
