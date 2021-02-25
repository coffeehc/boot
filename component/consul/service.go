package consul

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/plugin"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "consul_client"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetConsulClient() *api.Client
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
	impl := newService()
	service = impl
	plugin.RegisterPluginByFast(name, nil, func(ctx context.Context) errors.Error {
		return impl.destroy()
	})

}
