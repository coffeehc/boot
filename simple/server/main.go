package main

import (
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/microserviceboot/serviceboot/grpcboot"
)

func main() {
	var service = &Service{}
	serviceboot.ServiceLauncher(service, grpcboot.GRpcMicroServiceBuilder)
}
