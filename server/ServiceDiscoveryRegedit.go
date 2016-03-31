package server

import "github.com/coffeehc/microserviceboot/base"

type ServiceDiscoveryRegister interface {
	//注册服务
	RegService(info base.ServiceInfo, endpoints []base.EndPoint, servicePort int) error
}
