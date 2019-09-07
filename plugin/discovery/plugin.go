package discovery

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/component/etcdsd"
	"git.xiagaogao.com/coffee/boot/plugin"
)

var impl plugin.Plugin
var mutex = new(sync.Mutex)

type pluginImpl struct {
}

func (impl *pluginImpl) Start(ctx context.Context) errors.Error {
	return nil
}
func (impl *pluginImpl) Stop(ctx context.Context) errors.Error {
	return nil
}

func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if impl != nil {
		return
	}
	etcdsd.InitEtcdClient()
	impl = &pluginImpl{}
	plugin.RegisterPlugin("serviceDiscovery", impl)
}
