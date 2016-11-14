package main

import (
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/microserviceboot/serviceboot/grpcboot"
)

func main() {
	serviceboot.ServiceLauncher(&Service{}, grpcboot.GRpcMicroServiceBuilder)
}
