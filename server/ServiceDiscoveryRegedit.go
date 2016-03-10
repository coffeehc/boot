package server

import "github.com/coffeehc/microserviceboot/common"

type ServiceDiscoveryRegister interface {
	//注册服务
	RegService(info common.ServiceInfo, endpoints []common.EndPoint) error
}
