package consultool

import (
	"fmt"

	"context"
	"net"
	"strconv"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/hashicorp/consul/api"
)

const errScopeConsulRegister = "consul register"

type consulServiceRegister struct {
	client    *api.Client
	serviceID string
	checkID   string
}

//NewConsulServiceRegister 构建一个 base.ServiceDiscoveryRegister的基于 consul 的实现实例
func NewConsulServiceRegister(consulClient *api.Client) (base.ServiceDiscoveryRegister, base.Error) {
	if consulClient == nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, errScopeConsulRegister, "没有指定 consulClient")
	}
	return &consulServiceRegister{
		client: consulClient,
	}, nil
}

func (csr *consulServiceRegister) RegService(cxt context.Context, serviceInfo base.ServiceInfo, serviceAddr string) base.Error {
	if serviceAddr == "" {
		return base.NewError(-1, errScopeConsulRegister, "serverAddr is nil")
	}
	csr.serviceID = fmt.Sprintf("%s-%s", serviceInfo.GetServiceName(), serviceAddr)
	csr.checkID = fmt.Sprintf("service:%s", csr.serviceID)
	_, port, err := net.SplitHostPort(serviceAddr)
	if err != nil {
		return base.NewError(-1, errScopeConsulRegister, "serviceAddr is not a tcp addr")
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
	err = csr.client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Error("注册服务失败:%s", err)
		return base.NewError(base.ErrCodeBaseSystemServiceRegister, errScopeConsulRegister, err.Error())
	}
	//context.WithValue(cxt, Context_ConsulClient, csr.client)
	return nil
}

//const Context_ConsulClient = &"__consulClient"
//
//func GetConsulClient(cxt context.Context) (*api.Client, base.Error) {
//	i := cxt.Value(Context_ConsulClient)
//	if i == nil {
//		return nil, base.NewError(base.ErrCodeBaseSystemInit, errScopeConsulRegister, "no create consul client")
//	}
//	if client, ok := i.(*api.Client); ok {
//		return client, nil
//	}
//	return nil, base.NewError(base.ErrCodeBaseSystemInit, errScopeConsulRegister, "no create consul client")
//}
