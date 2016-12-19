package serviceboot

import (
	"context"
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"time"
	"os"
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

func Util_StopService(service base.Service){
	if service != nil && service.Stop != nil {
		stopErr := service.Stop()
		if stopErr != nil {
			logger.Error("关闭服务失败,%s\n", stopErr)
		}
	}
}

func Util_StartService(service base.Service) (err base.Error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			if e, ok := err1.(base.Error); ok {
				err = e
				return
			}
			err = base.NewError(base.ERRCODE_BASE_SYSTEM_UNKNOWN, err_scope_startService, fmt.Sprintf("service crash,cause is %s", err1))
		}
	}()
	if service == nil {
		panic(base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR, err_scope_startService, "没有 Service 的实例"))
	}
	if service.Run == nil {
		panic(base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR, err_scope_startService, "没有指定Run方法"))
	}
	err1 := service.Run()
	if err1 != nil {
		panic(err1)
	}
	logger.Info("服务已正常启动")
	return
}

func launchError(err error) {
	logger.Error("启动失败:%s", err.Error())
	time.Sleep(500 * time.Millisecond)
	os.Exit(-1)
}
