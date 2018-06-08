package serviceboot

import (
	"context"
	"flag"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/bootutils"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/transport/grpcserver"
)

var configPath = flag.String("config", "./config.yml", "配置文件路径")

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServiceInfo            *boot.SimpleServiceInfo `yaml:"service_info"`
	EnableAccessInfo       bool                    `yaml:"enableAccessInfo"`
	DisableServiceRegister bool                    `yaml:"disable_service_register"`
	ServerAddr             string                  `yaml:"server_addr"`
	GrpcConfig             *grpcserver.GRPCConfig  `yaml:"grpc_config"`
	EtcdConfig             *etcdsd.Config          `yaml:"etcd_config"`
}

//LoadConfig 加载ServiceConfiguration的配置
func loadServiceConfig(ctx context.Context) (*ServiceConfig, string, errors.Error) {
	errorService := errors.NewService("boot.config")
	config := &ServiceConfig{}
	loaded := false
	if boot.IsDevModule() && *configPath == "./config.yml" {
		err := bootutils.LoadConfig(ctx, "./config-dev.yml", config)
		if err == nil {
			*configPath = "./config-dev.yml"
			loaded = true
		}
	}
	if !loaded {
		err := bootutils.LoadConfig(ctx, *configPath, config)
		if err != nil {
			return nil, "", err
		}
	}
	if config.ServiceInfo == nil {
		return nil, "", errorService.BuildMessageError("没有配置ServiceInfo")
	}
	err := checkServiceInfoConfig(ctx, config.ServiceInfo, errorService)
	if err != nil {
		return nil, "", err
	}
	if config.ServerAddr == "" {
		return nil, "", errorService.BuildMessageError("没有配置ServiceAddr")
	}
	config.ServerAddr, err = bootutils.WarpServerAddr(config.ServerAddr)
	if err != nil {
		return nil, "", err
	}
	return config, *configPath, nil
}

func checkServiceInfoConfig(ctx context.Context, serviceInfo boot.ServiceInfo, errorService errors.Service) errors.Error {
	if serviceInfo == nil {
		return errorService.BuildMessageError("没有配置 ServiceInfo")
	}
	if serviceInfo.GetServiceName() == "" {
		return errorService.BuildMessageError("没有配置 ServiceName")
	}
	if serviceInfo.GetServiceTag() == "" {
		return errorService.BuildMessageError("没有配置 ServiceTag")
	}
	if serviceInfo.GetVersion() == "" {
		return errorService.BuildMessageError("没有配置 ServiceVersion")
	}
	return nil
}
