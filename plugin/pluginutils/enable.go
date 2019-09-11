package pluginutils

import (
	"context"

	"git.xiagaogao.com/coffee/boot/component/etcdsd"
	"git.xiagaogao.com/coffee/boot/plugin/discovery"
	"git.xiagaogao.com/coffee/boot/plugin/manage"
	"git.xiagaogao.com/coffee/boot/plugin/register"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
)

func EnableMicorPlugin(ctx context.Context) {
	rpc.EnablePlugin(ctx)
	etcdsd.InitEtcdClient()
	discovery.EnablePlugin(ctx)
	manage.EnablePlugin(ctx)
	register.EnablePlugin(ctx)
}
