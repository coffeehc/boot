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

const err_scope_startService = "startService"

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
	microService,err := Launch(service,serviceBuilder,cxt)
	if err != nil{
		launchError(err)
		return
	}
	defer microService.Stop()
	commons.WaitStop()
}

func Launch(service base.Service, serviceBuilder MicroServiceBuilder, cxt context.Context) (MicroService,base.Error){
	logger.Info("launch microService")
	startTime := time.Now()
	if service == nil {
		logger.Error("service is nil")
		return nil,base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR,"Launch","service is nil")
	}
	microService, err := serviceBuilder(service)
	if err != nil {
		return nil,err
	}
	logger.Info("Service initing")
	config, initErr := microService.Init(cxt)
	if initErr != nil {
		return nil,initErr
	}
	logger.Info("Service inited")
	serviceInfo := microService.GetServiceInfo()
	if serviceInfo == nil {
		return nil,base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR,"Launch","没有指定 ServiceInfo")
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	logger.Info("Service starting")
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
	return microService,nil
}


