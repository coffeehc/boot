package pluginutils

import (
	"context"

	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/plugin/discovery"
	"git.xiagaogao.com/coffee/boot/plugin/manage"
	"git.xiagaogao.com/coffee/boot/plugin/register"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
)

func EnableMicorPlugin(ctx context.Context) {
	plugin.EnablePlugins(ctx,
		rpc.EnablePlugin,
		discovery.EnablePlugin,
		manage.EnablePlugin,
		register.EnablePlugin,
	)
}
