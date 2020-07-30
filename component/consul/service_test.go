package consul

import (
	"context"
	"testing"

	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/testutils"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

func testConfigInit() {
	config := &api.Config{
		Address:    "127.0.0.1:8500",
		Datacenter: "coffee",
	}
	viper.Set("consul", config)
	serviceInfo := configuration.ServiceInfo{
		ServiceName: "test",
		Version:     "0.0.1",
		Scheme:      configuration.MicroServiceProtocolScheme,
	}
	testutils.InitTestConfig()
	configuration.InitConfiguration(context.TODO(), serviceInfo)
}

func TestGetService(t *testing.T) {
	testConfigInit()
	ctx := context.TODO()
	EnablePlugin(ctx)
	service := GetService()
	client := service.GetConsulClient()
	//health := client.Health()
	dcs, err := client.Catalog().Datacenters()
	if err != nil {
		t.Fatal("获取dc失败")
	}
	for _, dc := range dcs {
		t.Logf("获取了dc：%s", dc)
	}
	agent := client.Agent()
	agent.ServiceDeregister("baseService_192.168.4.105")
	//serviceInfo := configuration.GetServiceInfo()
	//register := &api.AgentServiceRegistration{
	//	Kind:    api.ServiceKindTypical,
	//	ID:      "",
	//	Name:    serviceInfo.ServiceName,
	//	Tags:    []string{configuration.GetRunModel()},
	//	Address: "127.0.0.1",
	//	Port:    8080,
	//	Meta: map[string]string{
	//		"version": serviceInfo.Version,
	//	},
	//}
	//err = agent.ServiceRegisterOpts(register, api.ServiceRegisterOpts{
	//	ReplaceExistingChecks: true,
	//})
	//time.Sleep(time.Second*10)
	//time.Sleep(time.Hour)
}
