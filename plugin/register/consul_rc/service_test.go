package consul_rc

import (
	"context"
	"testing"
	"time"

	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin"
	"github.com/coffeehc/boot/testutils"
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
