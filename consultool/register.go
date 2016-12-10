package consultool

import (
	"fmt"

	"context"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/hashicorp/consul/api"
	"net"
	"strconv"
)

const err_scope_consul_register  = "consul register"

type consulServiceRegister struct {
	client    *api.Client
	serviceId string
	checkId   string
}

func NewConsulServiceRegister(consulClient *api.Client) (base.ServiceDiscoveryRegister, base.Error) {
	if consulClient == nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR,err_scope_consul_register, "没有指定 consulClient")
	}
	return &consulServiceRegister{
		client: consulClient,
	}, nil

}

func (this *consulServiceRegister) RegService(serviceInfo base.ServiceInfo, serviceAddr string, cxt context.Context) base.Error {
	if serviceAddr == "" {
		return base.NewError(-1,err_scope_consul_register, "serverAddr is nil")
	}
	this.serviceId = fmt.Sprintf("%s-%s", serviceInfo.GetServiceName(), serviceAddr)
	this.checkId = fmt.Sprintf("service:%s", this.serviceId)
	_, port, err := net.SplitHostPort(serviceAddr)
	if err != nil {
		return base.NewError(-1,err_scope_consul_register, "serviceAddr is not a tcp addr")
	}
	p, _ := strconv.Atoi(port)
	registration := &api.AgentServiceRegistration{
		Name: serviceInfo.GetServiceName(),
		Tags: []string{serviceInfo.GetServiceTag()},
		Port: p,
		//Address:           addr, //http 获取节点的情况下,或出现问题
		EnableTagOverride: true,
		Checks: api.AgentServiceChecks([]*api.AgentServiceCheck{
			{
				HTTP:          fmt.Sprintf("%s://%s/health", serviceInfo.GetScheme(), serviceAddr),
				Interval:      "10s",
				Status:        "passing",
				TLSSkipVerify: true,
			},
		}),
	}
	err = this.client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Error("注册服务失败:%s", err)
		return base.NewError(base.ERROR_CODE_BASE_SERVICE_REGISTER_ERROR,err_scope_consul_register, err.Error())
	}
	context.WithValue(cxt, Context_ConsulClient, this.client)
	return nil
}

const Context_ConsulClient = "__consulClient"

func GetConsulClient(cxt context.Context) (*api.Client, base.Error) {
	i := cxt.Value(Context_ConsulClient)
	if i == nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR,err_scope_consul_register, "no create consul client")
	}
	if client, ok := i.(*api.Client); ok {
		return client, nil
	}
	return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR,err_scope_consul_register, "no create consul client")
}
