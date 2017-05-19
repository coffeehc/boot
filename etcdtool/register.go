package etcdtool

import (
	"context"

	"fmt"

	"net"

	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coreos/etcd/clientv3"
	"github.com/pquerna/ffjson/ffjson"
)

type ServiceRegisterInfo struct {
	ServiceInfo *base.SimpleServiceInfo `json:"info"`
}

var timeout = time.Second * 5

func NewEtcdServiceRegister(client *clientv3.Client) (base.ServiceDiscoveryRegister, base.Error) {
	return &etcdServiceRegister{
		client: client,
	}, nil
}

type etcdServiceRegister struct {
	client *clientv3.Client
}

func (reg *etcdServiceRegister) RegService(cxt context.Context, info base.ServiceInfo, serviceAddr string) (deregister func(), err base.Error) {
	// 注册格式  internal_ms.${servicename}.${tag}.${instance:port}
	if info.GetServiceName() == "" && info.GetServiceTag() == "" {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etcd", "没有指定ServiceInfo内容")
	}
	if _, err := net.ResolveTCPAddr("tcp", serviceAddr); err != nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etcd", "服务地址不是一个标准的tcp地址")
	}
	leaseGrantResponse, _err := reg.client.Lease.Grant(cxt, int64(timeout/time.Second))
	if _err != nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etcd", "创建租约失败")
	}
	keepAlive(cxt, reg.client, leaseGrantResponse.ID)
	serviceKey := fmt.Sprintf("%s%s", buildServiceKeyPrefix(info.GetServiceName()), serviceAddr)
	logger.Debug("serviceKey is %s", serviceKey)
	value, _ := ffjson.Marshal(&ServiceRegisterInfo{ServiceInfo: info.(*base.SimpleServiceInfo)})
	reg.client.KV.Put(cxt, serviceKey, string(value), clientv3.WithLease(leaseGrantResponse.ID))
	return func() {
		reg.client.KV.Delete(cxt, serviceKey)
	}, nil
}

func keepAlive(cxt context.Context, client *clientv3.Client, leaseId clientv3.LeaseID) base.Error {
	cancel, cancelFunc := context.WithCancel(cxt)
	leaseKeepAliveResponseChe, _err := client.Lease.KeepAlive(cancel, leaseId)
	if _err != nil {
		return base.NewError(base.ErrCodeBaseSystemInit, "etcd", "KeepAlive创建租约失败")
	}
	go func(leaseKeepAliveResponse <-chan *clientv3.LeaseKeepAliveResponse) {
		timer := time.NewTimer(timeout / 2)
		var retry = func() {
			cancelFunc()
			time.Sleep(time.Second)
			keepAlive(cxt, client, leaseId)
		}
		for {
			select {
			case response, ok := <-leaseKeepAliveResponse:
				//logger.Debug("Revision:%d,TTL:%ds", response.Revision, response.TTL)
				if !ok || response == nil {
					logger.Debug("管道关闭,重新建立链接")
					retry()
					return
				}
				//TODO 超时设置
			case <-timer.C:
				logger.Debug("超时了,重新建立连接")
				retry()
				return

			}
			timer.Reset(timeout / 2)
		}
	}(leaseKeepAliveResponseChe)
	return nil
}
