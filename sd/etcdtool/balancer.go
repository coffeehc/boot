package etcdtool

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/loadbalancer"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"github.com/coreos/etcd/clientv3"
)

func NewEtcdBalancer(cxt context.Context, client *clientv3.Client, serviceInfo serviceboot.ServiceInfo) (loadbalancer.Balancer, errors.Error) {
	etcdRecolver, err := newEtcdResolver(client, serviceInfo.GetServiceName(), serviceInfo.GetServiceTag())
	if err != nil {
		return nil, err
	}
	return loadbalancer.RoundRobin(etcdRecolver), nil
}
