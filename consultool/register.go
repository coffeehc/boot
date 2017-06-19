package consultool

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/hashicorp/consul/api"
)

const errScopeConsulRegister = "consul register"

type consulServiceRegister struct {
	client *api.Client
}

//NewConsulServiceRegister 构建一个 base.ServiceDiscoveryRegister的基于 consul 的实现实例
func NewConsulServiceRegister(consulClient *api.Client) (base.ServiceDiscoveryRegister, base.Error) {
	if consulClient == nil {
		return nil, base.NewError(base.Error_System, errScopeConsulRegister, "没有指定 consulClient")
	}
	return &consulServiceRegister{
		client: consulClient,
	}, nil
}

func (csr *consulServiceRegister) RegService(cxt context.Context, serviceInfo base.ServiceInfo, serviceAddr string) (func(), base.Error) {
	if serviceAddr == "" {
		return nil, base.NewError(base.Error_System, errScopeConsulRegister, "serverAddr is nil")
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", serviceAddr)
	if err != nil {
		return nil, base.NewError(base.Error_System, errScopeConsulRegister, "serviceAddr is not a tcp addr")
	}
	addr := tcpAddr.IP.String()
	if tcpAddr.IP.Equal(net.IPv4zero) {
		addr, err = base.GetLocalIP()
		if err != nil {
			return nil, base.NewErrorWrapper(base.Error_System, "consul", err)
		}
		return nil, base.NewError(base.Error_System, errScopeConsulRegister, "没有指定具体的注册 IP")
	}
	serviceAddr = net.JoinHostPort(addr, strconv.Itoa(tcpAddr.Port))
	logger.Info("向Consul注册地址为:%s", serviceAddr)
	_, port, err := net.SplitHostPort(serviceAddr)
	if err != nil {
		return nil, base.NewError(base.Error_System, errScopeConsulRegister, "serviceAddr is not a tcp addr")
	}
	p, _ := strconv.Atoi(port)
	registration := &api.AgentServiceRegistration{
		ID:                serviceAddr,
		Name:              serviceInfo.GetServiceName(),
		Tags:              []string{serviceInfo.GetServiceTag()},
		Port:              p,
		Address:           addr, //http 获取节点的情况下,或出现问题
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
		return nil, base.NewError(base.Error_System, errScopeConsulRegister, err.Error())
	}
	return func() {
		csr.client.Agent().ServiceDeregister(serviceAddr)
		logger.Info("leave the consul,serviceID is %s", serviceAddr)
	}, nil
}
