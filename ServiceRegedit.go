package microserviceboot

import "github.com/coffeehc/microserviceboot/common"

type ServiceRegedit interface {
	//注册服务
	RegeditService(common.ServiceInfo, []common.EndPoint)
}

type ConsulServiceRegedit struct {
}

func (this *ConsulServiceRegedit) RegeditService(serviceInfo common.ServiceInfo, endpints []common.EndPoint) {

}
