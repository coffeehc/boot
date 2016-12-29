package consultool

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
)

type consulBalancerBuilder struct {
	consulClient *api.Client
}

func (cbb *consulBalancerBuilder) NewBalancer(cxt context.Context, serviceInfo base.ServiceInfo) (loadbalancer.Balancer, base.Error) {
	return newConsulBalancer(cxt, cbb.consulClient, serviceInfo)
}

//NewConsulBalancerBuilder 返回loadbalancer.BalancerBuilder的 consul 实现
func NewConsulBalancerBuilder(consulClient *api.Client) loadbalancer.BalancerBuilder {
	return &consulBalancerBuilder{
		consulClient: consulClient,
	}
}

func newConsulBalancer(cxt context.Context, consulClient *api.Client, serviceInfo base.ServiceInfo) (loadbalancer.Balancer, base.Error) {
	consulRecolver, err := newConsulResolver(consulClient, serviceInfo.GetServiceName(), serviceInfo.GetServiceTag())
	if err != nil {
		return nil, err
	}
	return loadbalancer.RoundRobin(consulRecolver), nil
}
