package serviceboot

import (
	"os"
	"time"

	"fmt"
	"log"

	"context"
	"flag"
	"github.com/coffeehc/commons"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

/**
 *	Service 启动
 */
func ServiceLaunch(service base.Service, serviceBuilder MicroServiceBuilder, cxt context.Context) {
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

	microService, err := serviceBuilder(service)
	if err != nil {
		log.Printf("初始化微服务出错:%s\n", err.Error())
		return
	}
	logger.Info("Service initing")
	config, initErr := microService.Init(cxt)
	if initErr != nil {
		log.Printf("初始化微服务出错:%s\n", initErr.Error())
		return
	}
	logger.Info("Service inited")
	serviceInfo := microService.GetServiceInfo()
	if serviceInfo == nil {
		logger.Error("没有指定 ServiceInfo")
		return
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	logger.Info("Service starting")
	err = startService(service)
	defer microService.Stop()
	defer func(service base.Service) {
		if service != nil && service.Stop != nil {
			stopErr := service.Stop()
			if stopErr != nil {
				log.Printf("关闭服务失败,%s\n", stopErr)
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
	logger.Info("核心服务启动成功,服务地址:%s", config.GetWebServerConfig().ServerAddr)
	defer func() {
		fmt.Printf("服务[%s]关闭\n", serviceInfo.GetServiceName())
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

func launchError(err error) {
	logger.Error("启动失败:%s", err.Error())
	time.Sleep(500 * time.Millisecond)
	os.Exit(-1)
}
