package main

import (
	"os"

	"github.com/coffeehc/microserviceboot"
	"github.com/coffeehc/microserviceboot/common"
	"github.com/coffeehc/web"
)

func main() {
	config := new(microserviceboot.MicorServiceCofig)
	config.Service = &TestService{}
	config.WebConfig = &web.ServerConfig{
		OpenTLS:         true,
		CertFile:        "server.crt",
		KeyFile:         "server.key",
		HttpErrorLogout: os.Stderr,
	}
	microserviceboot.ServiceLauncher(config)
}

type TestService struct {
}

func (this *TestService) Run() error {
	return nil
}

func (this *TestService) Stop() error {
	return nil
}

func (this *TestService) GetServiceInfo() common.ServiceInfo {
	return common.ServiceInfo{
		ServiceName: "testService",
		Descriptor:  "测试用的 Service",
		Version:     "0.0.1",
	}
}
func (this *TestService) GetEndPoints() []common.EndPoint {
	return []common.EndPoint{}
}
