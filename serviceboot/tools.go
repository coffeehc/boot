package serviceboot

import (
	"context"
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"os"
	"time"
)

func LoadConfig(serviceConfig ServiceConfigration) (string, base.Error) {
	*configPath = base.GetDefaultConfigPath(*configPath)
	err := base.LoadConfig(*configPath, serviceConfig)
	if err != nil {
		return "", err
	}
	logger.Debug("serviceboot Config is %#v", serviceConfig)
	if serviceConfig.GetServiceConfig().ServiceInfo == nil {
		return "", base.NewError(base.ERRCODE_BASE_SYSTEM_CONFIG_ERROR, base.ERR_SCOPE_LOADCONFIG, "没有配置ServiceInfo")
	}
	return *configPath, nil
}

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

func ServiceRegister(configPath string, service base.Service, serviceInfo base.ServiceInfo, serviceConfig *ServiceConfig, cxt context.Context) {
	serviceDiscoveryRegister, err := service.GetServiceDiscoveryRegister(configPath)
	if err != nil {
		launchError(fmt.Errorf("获取没有指定serviceDiscoveryRegister失败,注册服务[%s]失败", serviceInfo.GetServiceName()))
	}
	if !serviceConfig.DisableServiceRegister {
		if serviceDiscoveryRegister == nil {
			launchError(fmt.Errorf("没有指定serviceDiscoveryRegister,注册服务[%s]失败", serviceInfo.GetServiceName()))
		}
		registerError := serviceDiscoveryRegister.RegService(serviceInfo, serviceConfig.GetWebServerConfig().ServerAddr, cxt)
		if registerError != nil {
			launchError(fmt.Errorf("注册服务[%s]失败,%s", serviceInfo.GetServiceName(), registerError.Error()))
		}
		logger.Info("注册服务[%s]成功", serviceInfo.GetServiceName())
	}
}

func launchError(err error) {
	logger.Error("启动失败:%s", err.Error())
	time.Sleep(500 * time.Millisecond)
	os.Exit(-1)
}
