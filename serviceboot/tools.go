package serviceboot

import (
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

func LoadConfig(serviceConfig ServiceConfigration) string {
	*configPath = base.GetDefaultConfigPath(*configPath)
	err := base.LoadConfig(*configPath, serviceConfig)
	if err != nil {
		logger.Warn("加载服务器配置[%s]失败,%s", *configPath, err)
	}
	logger.Debug("serviceboot Config is %#v", serviceConfig)
	return *configPath
}

func CheckServiceInfoConfig(serviceInfo base.ServiceInfo) base.Error {
	if serviceInfo == nil {
		return base.NewError(-1, "没有配置 ServiceInfo")
	}
	if serviceInfo.GetServiceName() == "" {
		return base.NewError(-1, "没有配置 ServiceName")
	}
	if serviceInfo.GetServiceTag() == "" {
		return base.NewError(-1, "没有配置 ServiceTag")
	}
	if serviceInfo.GetVersion() == "" {
		return base.NewError(-1, "没有配置 ServiceVersion")
	}
	return nil
}

func ServiceRegister(service base.Service, serviceInfo base.ServiceInfo, serviceConfig *ServiceConfig) {
	serviceDiscoveryRegister, err := service.GetServiceDiscoveryRegister()
	if err != nil {
		launchError(fmt.Errorf("获取没有指定serviceDiscoveryRegister失败,注册服务[%s]失败", serviceInfo.GetServiceName()))
	}
	if !serviceConfig.DisableServiceRegister {
		if serviceDiscoveryRegister == nil {
			launchError(fmt.Errorf("没有指定serviceDiscoveryRegister,注册服务[%s]失败", serviceInfo.GetServiceName()))
		}
		registerError := serviceDiscoveryRegister.RegService(serviceInfo, serviceConfig.GetWebServerConfig().ServerAddr)
		if registerError != nil {
			launchError(fmt.Errorf("注册服务[%s]失败,%s", serviceInfo.GetServiceName(), registerError.Error()))
		}
		logger.Info("注册服务[%s]成功", serviceInfo.GetServiceName())
	}
}
