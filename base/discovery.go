package base

import "context"

type ServiceDiscoveryRegister interface {
	//注册服务
	RegService(info ServiceInfo, serviceAddr string, cxt context.Context) Error
}
