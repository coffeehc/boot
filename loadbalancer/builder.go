package loadbalancer

import (
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
)

type BalancerBuilder interface {
	NewBalancer(cxt context.Context, serviceInfo base.ServiceInfo) (Balancer, base.Error)
}

type SimpleBalancerBuilder struct {
	Addrs []string
}

func (this *SimpleBalancerBuilder) NewBalancer(cxt context.Context, serviceInfo base.ServiceInfo) (Balancer, base.Error) {
	if len(this.Addrs) == 0 {
		return nil, base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR, "BalancerBuilder", "no addrs")
	}
	return NewSimpleBalancer(this.Addrs)
}
