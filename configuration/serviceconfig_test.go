package configuration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	// viper.SetEnvPrefix("ENV")
	// viper.AutomaticEnv()
	// os.Setenv("ENV_REMOTE_CONFIG.ENABLE","true")
	// os.Setenv("ENV_REMOTE_CONFIG.CONSUL_ADDR","127.0.0.1:8500")
	os.Setenv("ENV_CONSUL.TOKEN", "2e9c367d-b9d8-0e75-26d0-5fde5e7dfac7")
	EnableRemoteConfig()
	fmt.Print(viper.GetString("consul_config.consul_addr"))
	SetRunModel("dev")
	// viper.SetDefault("ServiceName","r")
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*3)
	InitConfiguration(ctx, ServiceInfo{
		ServiceName: "test",
	})
	t.Logf("%s:%s", _run_model, GetRunModel())
	t.Logf("serviceName:%s", GetServiceName())
}
