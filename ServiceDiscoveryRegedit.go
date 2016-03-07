package microserviceboot

import "github.com/coffeehc/microserviceboot/common"

type ServiceDiscoveryRegister interface {
	//注册服务
	RegService(serverAddr string, info common.ServiceInfo, endpoints []common.EndPoint) error
}
