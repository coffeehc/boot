package internal

import (
	"flag"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/serviceboot"
)

var configPath = flag.String("config", "./config.yml", "配置文件路径")

//ServiceConfiguration  服务配置接口服务接口定义
type ServiceConfiguration interface {
	GetServiceConfig() *serviceboot.ServiceConfig
}

//LoadConfig 加载ServiceConfiguration的配置
func LoadConfig(serviceConfig ServiceConfiguration) (string, base.Error) {
	err := base.LoadConfig(*configPath, serviceConfig)
	if err != nil {
		return "", err
	}
	logger.Debug("serviceboot Config is %#v", serviceConfig.GetServiceConfig())
	if serviceConfig.GetServiceConfig().ServiceInfo == nil {
		return "", base.NewError(base.ErrCodeBaseSystemConfig, "load config", "没有配置ServiceInfo")
	}
	return *configPath, nil
}

//CheckServiceInfoConfig 检测 ServiceInfo 是否配置正确
func CheckServiceInfoConfig(serviceInfo base.ServiceInfo) base.Error {
	const errorScope = "checkServiceInfo"
	if serviceInfo == nil {
		return base.NewError(-1, errorScope, "没有配置 ServiceInfo")
	}
	if serviceInfo.GetServiceName() == "" {
		return base.NewError(-1, errorScope, "没有配置 ServiceName")
	}
	if serviceInfo.GetServiceTag() == "" {
		return base.NewError(-1, errorScope, "没有配置 ServiceTag")
	}
	if serviceInfo.GetVersion() == "" {
		return base.NewError(-1, errorScope, "没有配置 ServiceVersion")
	}
	return nil
}
