package configuration

import (
	"context"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	SetModel("dev")
	// viper.SetDefault("ServiceName","r")
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*3)
	InitConfiguration(ctx, ServiceInfo{
		ServiceName: "123",
	})
	t.Logf("model:%s", GetModel())
	t.Logf("serviceName:%s", GetServiceName())
}
