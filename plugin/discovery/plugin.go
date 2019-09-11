package discovery

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/boot/component/etcdsd"
	"git.xiagaogao.com/coffee/boot/plugin"
)

var impl plugin.Plugin
var mutex = new(sync.Mutex)

func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if impl != nil {
		return
	}
	etcdsd.InitEtcdClient()
	plugin.RegisterPluginByFast("serviceDiscovery", nil, nil)
}
