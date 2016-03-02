package microserviceboot_test

import (
	"testing"
)

func TestMain(m *testing.M) {
	//config := new(microserviceboot.BootConfig)
	//config.ServiceInfo = &microserviceboot.ServiceInfo{
	//	ServiceName:"test",
	//	Descriptor:"test micor Service project",
	//}
	//config.Service = &TestService{}
	//microserviceboot.ServiceLauncher(config)
}

type TestService struct {
}

func (this *TestService) Run() error {
	return nil
}

func (this *TestService) Stop() error {
	return nil
}
