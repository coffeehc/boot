package serviceboot

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"go.uber.org/zap"
)

//serviceLaunch Service 启动
func ServiceLaunch(ctx context.Context, service Service) {
	boot.InitModel()
	time.LoadLocation("Asia/Shanghai")
	if !flag.Parsed() {
		flag.Parse()
	}
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	logger, _ := zap.NewDevelopment()
	err1 := boot.CheckServiceInfoConfig(ctx, service.GetServiceInfo())
	if err1 != nil {
		logger.Error("校验服务信息失败", zap.String(logs.K_Cause, err1.Error()))
		return
	}
	serviceInfo := service.GetServiceInfo()
	errorService := errors.NewService(serviceInfo.ServiceName)
	ctx = boot.SetServiceName(ctx, serviceInfo.ServiceName)
	logService, err1 := logs.NewService(serviceInfo)
	if err1 != nil {
		logger.Error("创建logService失败", zap.String(logs.K_Cause, err1.Error()))
		return
	}
	logger = logService.GetLogger()
	boot.PrintServiceInfo(service.GetServiceInfo(), logger)
	ctx = boot.SetServiceName(ctx, serviceInfo.ServiceName)
	serviceConfig, configPath, err := loadServiceConfig(ctx, errorService, logger)
	if err != nil {
		logger.DPanic(err.Error(), err.GetFields()...)
		return
	}
	logger.Debug("运行模式", zap.String("model", boot.RunModel()))
	microService, err := Launch(ctx, service, serviceConfig, configPath, errorService, logger, logService)
	if err != nil {
		logger.DPanic(err.Error(), err.GetFields()...)
		return
	}
	defer func() {
		microService.Stop(ctx)
		logger.Info("服务关闭", zap.String(logs.K_ServiceName, serviceInfo.ServiceName))
	}()
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}

//Launch 纯粹的启动微服务,不做系统信令监听
func Launch(ctx context.Context, service Service, serviceConfig *ServiceConfig, configPath string, errorService errors.Service, logger *zap.Logger, loggerService logs.Service) (MicroService, errors.Error) {
	startTime := time.Now()
	if service == nil {
		return nil, errorService.SystemError("serviceboot is nil")
	}
	microService, err := newMicroService(ctx, service, configPath, errorService, logger, loggerService)
	if err != nil {
		return nil, err
	}
	err = microService.Start(ctx, serviceConfig)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("核心服务启动成功,服务地址:%s,启动耗时:%s", serviceConfig.ServerAddr, time.Since(startTime)))
	return microService, nil
}
