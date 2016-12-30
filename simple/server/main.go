package main

import (
	"context"

	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/microserviceboot/serviceboot/grpcboot"
)

func main() {
	serviceboot.ServiceLaunch(context.Background(), &_Service{}, grpcboot.GRPCMicroServiceBuilder)
}
