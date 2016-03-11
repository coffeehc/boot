package consultool

import (
	"fmt"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/common"
	"github.com/hashicorp/consul/api"
)

type ConsulServiceRegister struct {
	client    *api.Client
	serviceId string
	checkId   string
}

func NewConsulServiceRegister(consulConfig *api.Config) (*ConsulServiceRegister, error) {
	if consulConfig == nil {
		consulConfig = api.DefaultConfig()
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	return &ConsulServiceRegister{
		client: consulClient,
	}, nil

}

func (this *ConsulServiceRegister) RegService(serviceInfo common.ServiceInfo, endpints []common.EndPoint, servicePort int) error {
	ip := common.GetLocalIp()
	this.serviceId = fmt.Sprintf("%s-%s", serviceInfo.GetServiceName(), ip)
	this.checkId = fmt.Sprintf("service:%s", this.serviceId)
	registration := &api.AgentServiceRegistration{
		ID:                this.serviceId,
		Name:              serviceInfo.GetServiceName(),
		Tags:              common.WarpTags(serviceInfo.GetServiceTags()),
		Port:              servicePort,
		Address:           ip.String(),
		EnableTagOverride: false,
		Check: &api.AgentServiceCheck{
			TTL: "10s",
		},
	}
	err := this.client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Error("注册服务失败:%s", err)
		return err
	}
	go func() {
		timeout := 5 * time.Second
		timer := time.NewTimer(timeout)
		for {
			timer.Reset(timeout)
			select {
			case <-timer.C:
				this.client.Agent().PassTTL(this.checkId, fmt.Sprintf("ok %s", time.Now()))
			}
		}
	}()
	return nil
}
