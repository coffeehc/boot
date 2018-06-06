package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot"
)

func GetServiceName(ctx context.Context) string {
	return ctx.Value(boot.Ctx_Key_serviceName).(string)
}

func GetServiceInfo(ctx context.Context) ServiceInfo {
	return ctx.Value(boot.Ctx_Key_serviceInfo).(ServiceInfo)
}
