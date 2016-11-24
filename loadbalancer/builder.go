package loadbalancer

import "github.com/coffeehc/microserviceboot/base"

type BalancerBuilder interface {
	NewBalancer(serviceInfo base.ServiceInfo) (Balancer, base.Error)
}
