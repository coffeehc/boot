package consultool

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
)

type consuleBalancerBuilder struct {
	consulClient *api.Client
}

func (this *consuleBalancerBuilder) NewBalancer(cxt context.Context, serviceInfo base.ServiceInfo) (loadbalancer.Balancer, base.Error) {
	return NewConsulBalancer(cxt, this.consulClient, serviceInfo)
}

func NewConsulBalancerBudiler(consulClient *api.Client) loadbalancer.BalancerBuilder {
	return &consuleBalancerBuilder{
		consulClient: consulClient,
	}
}
