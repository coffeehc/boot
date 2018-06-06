package etcdtool

import (
	"context"

	"fmt"

	"net"

	"time"

	"git.xiagaogao.com/coffee/boot/bootutils"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"github.com/coreos/etcd/clientv3"
	"github.com/pquerna/ffjson/ffjson"
)

type ServiceRegisterInfo struct {
	ServiceInfo *serviceboot.SimpleServiceInfo `json:"info"`
}

var timeout = time.Second * 5

func NewEtcdServiceRegister(client *clientv3.Client) (serviceboot.ServiceDiscoveryRegister, errors.Error) {
	return &etcdServiceRegister{
		client: client,
	}, nil
}

type etcdServiceRegister struct {
	client     *clientv3.Client
	serviceKey string
}

func (reg *etcdServiceRegister) RegService(cxt context.Context, info serviceboot.ServiceInfo, serviceAddr string) (deregister func(), err errors.Error) {
	// 注册格式  internal_ms.${servicename}.${tag}.${instance:port}
	if info.GetServiceName() == "" && info.GetServiceTag() == "" {
		return nil, errors.NewError(errors.Error_System, "etcd", "没有指定ServiceName或者Tag内容")
	}
	err = reg.register(cxt, info, serviceAddr, false)
	if err != nil {
		return nil, err
	}
	return func() {
		_, _err := reg.client.KV.Delete(cxt, reg.serviceKey)
		if _err != nil {
			logger.Error("反注册服务失败," + _err.Error())
		}
	}, nil
}

func (reg *etcdServiceRegister) register(cxt context.Context, info serviceboot.ServiceInfo, serviceAddr string, reTry bool) errors.Error {
	addr, err := net.ResolveTCPAddr("tcp", serviceAddr)
	if err != nil {
		return errors.NewError(errors.Error_System, "etcd", fmt.Sprintf("服务地址不是一个标准的tcp地址:%s", err))
	}
	serverAddr := serviceAddr
	if addr.IP.Equal(net.IPv4zero) {
		localIp, err := bootutils.GetLocalIP()
		if err != nil {
			return errors.NewErrorWrapper(errors.Error_System, "etcd", err)
		}
		serverAddr = fmt.Sprintf("%s:%d", localIp, addr.Port)
	}
	serviceKey := fmt.Sprintf("%s%s", buildServiceKeyPrefix(info.GetServiceName(), info.GetServiceTag()), serverAddr)
	logger.Debug("serviceKey is %s", serviceKey)
	reg.serviceKey = serviceKey
	leaseGrantResponse, err := reg.client.Lease.Grant(cxt, int64(timeout/time.Second))
	if err != nil {
		if reTry {
			time.Sleep(time.Second * 3)
			go reg.register(cxt, info, serviceAddr, reTry)
			return nil
		}
		return errors.NewError(errors.Error_System, "etcd", "创建租约失败")
	}
	value, _ := ffjson.Marshal(&ServiceRegisterInfo{ServiceInfo: info.(*serviceboot.SimpleServiceInfo)})
	_, err = reg.client.Put(cxt, serviceKey, string(value), clientv3.WithLease(leaseGrantResponse.ID))
	if err != nil {
		if reTry {
			time.Sleep(time.Second * 3)
			go reg.register(cxt, info, serviceAddr, reTry)
			return nil
		}
		return errors.NewError(errors.Error_System, "etcd", "注册Service Key失败,"+err.Error())
	}
	baseErr := reg.keepAlive(cxt, leaseGrantResponse.ID, info, serviceKey, serviceAddr)
	if baseErr != nil {
		if reTry {
			time.Sleep(time.Second * 3)
			go reg.register(cxt, info, serviceAddr, reTry)
			return nil
		}
		return baseErr
	}
	return nil
}

func (reg *etcdServiceRegister) keepAlive(cxt context.Context, leaseId clientv3.LeaseID, info serviceboot.ServiceInfo, serviceKey string, serviceAddr string) errors.Error {
	cancel, cancelFunc := context.WithCancel(cxt)
	leaseKeepAliveResponseChe, _err := reg.client.Lease.KeepAlive(cancel, leaseId)
	if _err != nil {
		return errors.NewError(errors.Error_System, "etcd", "KeepAlive创建租约失败")
	}
	go func(leaseKeepAliveResponse <-chan *clientv3.LeaseKeepAliveResponse) {
		timer := time.NewTimer(timeout / 2)
		var reRegister = func() {
			cancelFunc()
			go reg.register(cxt, info, serviceAddr, true)
		}
		for {
			select {
			case response, ok := <-leaseKeepAliveResponse:
				if !ok {
					logger.Debug("管道关闭,重新建立链接")
					reRegister()
					return
				}
				if response == nil {
					logger.Debug("获取了一个空的response")
					reRegister()
					return
				}
				//TODO 超时设置
			case <-timer.C:
				logger.Debug("超时了,重新建立连接")
				reRegister()
				return

			}
			timer.Reset(timeout / 2)
		}
	}(leaseKeepAliveResponseChe)
	return nil
}
