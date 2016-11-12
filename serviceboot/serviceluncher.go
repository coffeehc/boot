package serviceboot

import (
	"os"
	"time"

	"fmt"
	"log"

	"flag"
	"github.com/coffeehc/commons"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

/**
 *	Service 启动
 */
func ServiceLauncher(service base.Service, serviceBuilder MicroServiceBuilder) {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	logger.InitLogger()
	defer logger.WaitToClose()
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	startTime := time.Now()
	if service == nil {
		logger.Error("service is nil")
		return
	}
	serviceInfo := service.GetServiceInfo()
	if serviceInfo == nil {
		logger.Error("没有指定 ServiceInfo")
		return
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	microService, err := serviceBuilder(service)
	if err != nil {
		log.Printf("初始化微服务出错:%s\n", err.Error())
		return
	}
	logger.Info("Service initing")
	config, initErr := microService.Init()
	if initErr != nil {
		log.Printf("初始化微服务出错:%s\n", err.Error())
		return
	}
	logger.Info("Service inited")
	logger.Info("Service starting")
	err = startService(service)
	defer microService.Stop()
	defer func(service base.Service) {
		if service != nil && service.Stop != nil {
			stopErr := service.Stop()
			if stopErr != nil {
				log.Printf("关闭服务失败,%s\n", stopErr)
				os.Exit(-1)
			}
		}
	}(service)
	if err != nil {
		log.Printf("start service error,%s\n", err)
		return
	}

	err = microService.Start()
	if err != nil {
		logger.Error("service start fail. %s", err)
		time.Sleep(time.Second)
		os.Exit(-1)
	}
	logger.Info("核心服务启动成功,服务地址:%s", config.GetServerAddr())
	serviceDiscoveryRegister := service.GetServiceDiscoveryRegister()
	if !config.DisableServiceRegister && serviceDiscoveryRegister != nil {
		registerError := serviceDiscoveryRegister.RegService(serviceInfo, config.GetServerAddr())
		if registerError != nil {
			logger.Info("注册服务[%s]失败,%s", service.GetServiceInfo().GetServiceName(), registerError.Error())
			time.Sleep(time.Second)
			os.Exit(-1)
		}
		logger.Info("注册服务[%s]成功", service.GetServiceInfo().GetServiceName())
	}
	defer func() {
		fmt.Printf("服务[%s]关闭\n", service.GetServiceInfo().GetServiceName())
	}()
	logger.Info("Service started [%s]", time.Since(startTime))
	commons.WaitStop()
}
func startService(service base.Service) (err base.Error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			if e, ok := err1.(base.Error); ok {
				err = e
				return
			}
			err = base.NewError(base.ERROR_CODE_BASE_SYSTEM_ERROR, fmt.Sprintf("service crash,cause is %s", err1))
		}
	}()
	if service == nil {
		panic(base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, "没有 Service 的实例"))
	}
	if service.Run == nil {
		panic(base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, "没有指定Run方法"))
	}
	err1 := service.Run()
	if err1 != nil {
		panic(err1)
	}
	logger.Info("服务已正常启动")
	return
}
