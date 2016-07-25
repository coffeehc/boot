package base

type ServiceDiscoveryRegister interface {
	//注册服务
	RegService(info ServiceInfo, endpoints []EndPoint, servicePort int) Error
}
