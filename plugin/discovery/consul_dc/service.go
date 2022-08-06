package consul_dc

import (
	"context"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"

	"sync"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "consul_discovery"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetResolverBuilder(ctx context.Context) resolver.Builder
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
	plugin.RegisterPlugin(name, service)
}
