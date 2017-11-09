package serviceboot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

//ServiceRegister 服务注册到服务发现中心,暂时支持的就是 consul
func serviceDiscoverRegister(cxt context.Context, service base.Service, serviceInfo base.ServiceInfo, serviceConfig *ServiceConfig) func() {
	serviceDiscoveryRegister, err := service.GetServiceDiscoveryRegister()
	if err != nil {
		launchError(fmt.Errorf("获取没有指定serviceDiscoveryRegister失败,注册服务[%s]失败", serviceInfo.GetServiceName()))
	}
	if !serviceConfig.DisableServiceRegister {
		if serviceDiscoveryRegister == nil {
			launchError(fmt.Errorf("没有指定serviceDiscoveryRegister,注册服务[%s]失败", serviceInfo.GetServiceName()))
		}
		httpServerConfig, err := serviceConfig.GetHTTPServerConfig()
		if err != nil {
			launchError(fmt.Errorf("没有可用的 Http server 的配置,注册服务[%s]失败", serviceInfo.GetServiceName()))
		}
		serverAddr := httpServerConfig.ServerAddr

		deregister, registerError := serviceDiscoveryRegister.RegService(cxt, serviceInfo, serverAddr)
		if registerError != nil {
			launchError(fmt.Errorf("注册服务[%s]失败,%s", serviceInfo.GetServiceName(), registerError.Error()))
		}
		logger.Info("注册服务[%s]成功", serviceInfo.GetServiceName())
		return deregister
	}
	return func() {}
}

func launchError(err error) {
	if base.IsBaseError(err) {
		berr := err.(base.Error)
		logger.Error("启动失败:[%s][%s]%s", berr.GetCode(), berr.GetScopes(), err.Error())
	} else {
		logger.Error("启动失败:%s", err.Error())
	}
	time.Sleep(500 * time.Millisecond)
	os.Exit(-1)
}
