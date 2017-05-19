package etcdtool

import (
	"context"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/coreos/etcd/clientv3"
)

func NewEtcdBalancer(cxt context.Context, client *clientv3.Client, serviceInfo base.ServiceInfo) (loadbalancer.Balancer, base.Error) {
	consulRecolver, err := newEtcdResolver(client, serviceInfo.GetServiceName(), serviceInfo.GetServiceTag())
	if err != nil {
		return nil, err
	}
	return loadbalancer.RoundRobin(consulRecolver), nil
}
