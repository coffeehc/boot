package consultool

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/hashicorp/consul/api"
)

func NewConsulBalancer(consulClient *api.Client, serviceInfo base.ServiceInfo) (loadbalancer.Balancer, base.Error) {
	consulRecolver, err := NewConsulResolver(consulClient, serviceInfo.GetServiceName(), serviceInfo.GetServiceTag())
	if err != nil {
		return nil, err
	}
	return loadbalancer.RoundRobin(consulRecolver), nil
}
