package consul_rc

import (
	"context"
	"testing"
	"time"

	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/testutils"
)

func TestService(t *testing.T) {
	testutils.InitTestConfig()
	ctx := context.TODO()
	serviceInfo := configuration.ServiceInfo{
		ServiceName: "test",
		Version:     "0.0.1",
		Scheme:      configuration.MicroServiceProtocolScheme,
	}
	configuration.InitConfiguration(ctx, serviceInfo)
	EnablePlugin(ctx)
	service := GetService()
	service.Register(ctx, serviceInfo)
	plugin.StartPlugins(ctx)
	time.Sleep(time.Hour)
}
