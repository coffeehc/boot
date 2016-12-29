package loadbalancer

import (
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
)

//BalancerBuilder balancer 构建接口
type BalancerBuilder interface {
	NewBalancer(cxt context.Context, serviceInfo base.ServiceInfo) (Balancer, base.Error)
}

type _BalancerBuilder struct {
	Addrs []string
}

func (bb *_BalancerBuilder) NewBalancerWithAddrArray(cxt context.Context, serviceInfo base.ServiceInfo) (Balancer, base.Error) {
	if len(bb.Addrs) == 0 {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "BalancerBuilder", "no addrs")
	}
	return newAddrArrayBalancer(bb.Addrs)
}
