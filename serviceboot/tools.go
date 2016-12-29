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
func serviceDiscoverRegister(cxt context.Context, service base.Service, serviceInfo base.ServiceInfo, serviceConfig *ServiceConfig) {
	serviceDiscoveryRegister, err := service.GetServiceDiscoveryRegister()
	if err != nil {
		launchError(fmt.Errorf("获取没有指定serviceDiscoveryRegister失败,注册服务[%s]失败", serviceInfo.GetServiceName()))
	}
	if !serviceConfig.DisableServiceRegister {
		if serviceDiscoveryRegister == nil {
			launchError(fmt.Errorf("没有指定serviceDiscoveryRegister,注册服务[%s]失败", serviceInfo.GetServiceName()))
		}
		registerError := serviceDiscoveryRegister.RegService(cxt, serviceInfo, serviceConfig.GetHTTPServerConfig().ServerAddr)
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
