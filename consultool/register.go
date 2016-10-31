package consultool

import (
	"fmt"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/hashicorp/consul/api"
)

type ConsulServiceRegister struct {
	client    *api.Client
	serviceId string
	checkId   string
}

func NewConsulServiceRegister(consulConfig *api.Config) (*ConsulServiceRegister, base.Error) {
	if consulConfig == nil {
		consulConfig = api.DefaultConfig()
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Error("创建 Consul Client 失败")
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error())
	}
	return &ConsulServiceRegister{
		client: consulClient,
	}, nil

}

func (this *ConsulServiceRegister) RegService(serviceInfo base.ServiceInfo, serviceAddr string, servicePort int) base.Error {
	if serviceAddr == "" {
		serviceAddr = base.GetLocalIp().String()
	}
	this.serviceId = fmt.Sprintf("%s-%s", serviceInfo.GetServiceName(), serviceAddr)
	this.checkId = fmt.Sprintf("service:%s", this.serviceId)
	registration := &api.AgentServiceRegistration{
		Name:              serviceInfo.GetServiceName(),
		Tags:              base.WarpTags(serviceInfo.GetServiceTags()),
		Port:              servicePort,
		Address:           serviceAddr,
		EnableTagOverride: true,
		Checks: api.AgentServiceChecks([]*api.AgentServiceCheck{
			{
				HTTP:     fmt.Sprintf("http://%s:%d/debug/pprof/threadcreate?debug=1", serviceAddr, servicePort),
				Interval: "10s",
			},
			{
				HTTP:     fmt.Sprintf("http://%s:%d/debug/pprof/block?debug=1", serviceAddr, servicePort),
				Interval: "10s",
			},
		}),
	}
	err := this.client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Error("注册服务失败:%s", err)
		return base.NewError(base.ERROR_CODE_BASE_SERVICE_REGISTER_ERROR, err.Error())
	}
	return nil
}
