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

//ServiceLaunch Service 启动
func ServiceLaunch(cxt context.Context, service base.Service, serviceBuilder MicroServiceBuilder) {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	logger.InitLogger()
	defer logger.WaitToClose()
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if base.IsDevModule() {
		logger.SetDefaultLevel("/", logger.LevelDebug)
		logger.Debug("当前为:开发模式")
	} else {
		logger.Debug("当前为:生产模式")
	}
	microService, err := Launch(cxt, service, serviceBuilder)
	if err != nil {
		launchError(err)
		return
	}
	defer func() {
		microService.Stop()
		fmt.Printf("服务[%s]关闭\n", microService.GetServiceInfo().GetServiceName())
	}()
	commons.WaitStop()
}

//Launch 纯粹的启动微服务,不做系统信令监听
func Launch(cxt context.Context, service base.Service, serviceBuilder MicroServiceBuilder) (MicroService, base.Error) {
	logger.Info("launch microService")
	startTime := time.Now()
	if service == nil {
		logger.Error("service is nil")
		return nil, base.NewError(base.ErrCode_System, "Launch", "service is nil")
	}
	microService, err := serviceBuilder(service)
	if err != nil {
		return nil, err
	}
	logger.Info("Service initing")
	config, initErr := microService.Init(cxt)
	if initErr != nil {
		return nil, initErr
	}
	logger.Info("Service inited")
	serviceInfo := microService.GetServiceInfo()
	if serviceInfo == nil {
		return nil, base.NewError(base.ErrCode_System, "Launch", "没有指定 ServiceInfo")
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	logger.Info("Service starting")
	err = microService.Start(cxt)
	if err != nil {
		launchError(err)
	}
	httpServerConfig, err := config.GetHTTPServerConfig()
	if err != nil {
		launchError(err)
	}
	logger.Info("核心服务启动成功,服务地址:%s,启动耗时:%s", httpServerConfig.ServerAddr, time.Since(startTime))
	//注册是在服务完全启动之后
	deregisterFunc := serviceDiscoverRegister(cxt, microService.GetService(), microService.GetServiceInfo(), config)
	microService.AddCleanFunc(deregisterFunc)
	return microService, nil
}
