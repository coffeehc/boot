package loadbalancer

import (
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
)

type BalancerBuilder interface {
	NewBalancer(cxt context.Context,serviceInfo base.ServiceInfo) (Balancer, base.Error)
}
