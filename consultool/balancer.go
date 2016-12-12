package consultool

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
)

func NewConsulBalancer(cxt context.Context,consulClient *api.Client, serviceInfo base.ServiceInfo) (loadbalancer.Balancer, base.Error) {
	consulRecolver, err := NewConsulResolver(consulClient, serviceInfo.GetServiceName(), serviceInfo.GetServiceTag())
	if err != nil {
		cxt.Deadline()
		return nil, err
	}
	return loadbalancer.RoundRobin(consulRecolver), nil
}
