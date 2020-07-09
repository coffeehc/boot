package etcd_rc

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/etcd"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/plugin/register/internal"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "etcd_registercenter"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Fatal("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	internal.RegisterCenter
	GetClient() *clientv3.Client
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
	internal.EnablePlugin(ctx)
	service = newServiceImpl()
	internal.GetService().SetRegisterCenter(service)
	plugin.RegisterPluginByFast(name, nil, nil)
}
