package etcd_dc

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/etcd"
	"git.xiagaogao.com/coffee/boot/plugin"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"

	"sync"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "etcd_discovery"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Fatal("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetResolverBuilder(ctx context.Context) (resolver.Builder, errors.Error)
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
	etcd.InitEtcdClient()
	impl := newService()
	service = impl
	plugin.RegisterPluginByFast(name, nil, nil)
}
