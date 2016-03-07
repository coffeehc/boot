package consultool

import (
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/common"
	"github.com/hashicorp/consul/api"
	"net"
	"time"
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

func (this *ConsulServiceRegister) RegService(serverAddr string, serviceInfo common.ServiceInfo, endpints []common.EndPoint) error {
	addr, _ := net.ResolveTCPAddr("tcp", serverAddr)
	ip := addr.IP
	if ip == nil {
		ip = common.GetLocalIp()
	}
	logger.Debug("addr is %t")
	this.serviceId = fmt.Sprintf("%s-%s", serviceInfo.GetServiceName(), ip.String())
	this.checkId = fmt.Sprintf("service:%s", this.serviceId)
	registration := &api.AgentServiceRegistration{
		ID:                this.serviceId,
		Name:              serviceInfo.GetServiceName(),
		Tags:              serviceInfo.GetServiceTags(),
		Port:              addr.Port,
		Address:           ip.String(),
		EnableTagOverride: false,
		Check: &api.AgentServiceCheck{
			TTL: "10s",
		},
		//Check:api.AgentServiceChecks{
		//	&api.AgentServiceCheck{
		//		HTTP:fmt.Sprintf("http://%s:%d/check/hralth", ip.String(), addr.Port),
		//		Interval:"10s",
		//	},
		//},
	}
	err := this.client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Error("注册服务失败:%s", err)
		return err
	}
	this.client
	go func() {
		this.client.Agent().CheckRegister()
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
