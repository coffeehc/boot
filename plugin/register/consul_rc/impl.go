package consul_rc

import (
	"context"
	"fmt"
	"net"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

type serviceImpl struct {
	client *api.Client
}

func newService() *serviceImpl {
	// service := consul.GetService()
	impl := &serviceImpl{
		// client: service.GetConsulClient(),
	}
	return impl
}

func (impl *serviceImpl) CheckDeregister(checkId string) {
	agent := impl.client.Agent()
	agent.CheckDeregister(checkId)
}

func (impl *serviceImpl) Register(ctx context.Context, serviceInfo configuration.ServiceInfo) errors.Error {
	if ctx.Err() != nil {
		return errors.MessageError("服务注册已经关闭")
	}
	agent := impl.client.Agent()
	rpcServerAddr := rpc.GetService().GetRPCServerAddr()
	addr, err := net.ResolveTCPAddr("tcp", rpcServerAddr)
	if err != nil {
		return errors.SystemError("RPC服务地址解析失败")
	}
	meta := serviceInfo.Metadata
	if meta == nil {
		meta = make(map[string]string)
	}
	meta["Version"] = serviceInfo.Version
	meta["Descriptor"] = serviceInfo.Descriptor
	meta["APIDefine"] = serviceInfo.APIDefine
	meta["Scheme"] = serviceInfo.Scheme
	meta["Address"] = rpcServerAddr
	for k, v := range serviceInfo.Metadata {
		meta[k] = v
	}
	serviceId := rpc.GetService().GetRegisterServiceId()
	deregisterCriticalServiceAfter := viper.GetString("register.deregisterCriticalServiceAfter")
	if configuration.GetRunModel() == configuration.Model_dev && deregisterCriticalServiceAfter == "" {
		deregisterCriticalServiceAfter = "30s"
	}
	register := &api.AgentServiceRegistration{
		ID:      serviceId,
		Name:    serviceInfo.ServiceName,
		Tags:    []string{configuration.GetRunModel()},
		Port:    addr.Port,
		Address: addr.IP.String(),
		Check: &api.AgentServiceCheck{
			CheckID:                        fmt.Sprintf("%s_grpcHealth", serviceId),
			Name:                           fmt.Sprintf("%s_grpcHealth", serviceId),
			GRPC:                           rpcServerAddr,
			Interval:                       "3s",
			Timeout:                        "2s",
			DeregisterCriticalServiceAfter: deregisterCriticalServiceAfter,
		},
		Meta: meta,
	}
	opts := api.ServiceRegisterOpts{
		ReplaceExistingChecks: true,
	}
	err = agent.ServiceRegisterOpts(register, opts)
	if err != nil {
		return errors.SystemError(err.Error())
	}

	return nil
}
