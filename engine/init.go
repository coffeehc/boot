package engine

import (
	"context"

	"git.xiagaogao.com/coffee/boot/configuration"
)

func initService(ctx context.Context, serviceInfo configuration.ServiceInfo) {
	configuration.InitConfiguration(ctx, serviceInfo)
}
