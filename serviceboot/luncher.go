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
	if !flag.Parsed() {
		flag.Parse()
	}
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	logger, _ := zap.NewDevelopment()
	serviceConfig, configPath, err := loadServiceConfig(ctx)
	if err != nil {
		logger.Error("加载基础配置失败", zap.String(logs.K_Cause, err.Error()))
		return
	}
	serviceInfo := serviceConfig.ServiceInfo
	ctx = boot.SetServiceInfo(ctx, serviceInfo)
	var errorService = errors.NewService(serviceInfo.GetServiceName())
	ctx = errors.SetRootErrorService(ctx, errorService)
	logService, err1 := logs.NewService()
	if err != nil {
		logger.Panic("创建logService失败", zap.String(logs.K_Cause, err1.Error()))
		return
	}
	ctx = logs.SetLoggerService(ctx, logService)
	logger = logService.GetLogger()
	ctx = logs.SetLogger(ctx, logger)
	if boot.IsDevModule() {
		logger.Debug("当前为:开发模式")
	} else {
		logger.Debug("当前为:生产模式")
	}
	microService, err := Launch(ctx, service, serviceConfig, configPath)
	if err != nil {
		logger.Error("启动服务失败", zap.String(logs.K_ServiceName, serviceInfo.GetServiceName()), zap.String(logs.K_Cause, err.Error()))
		return
	}
	defer func() {
		microService.Stop(ctx)
		logger.Info("服务关闭", zap.String(logs.K_ServiceName, microService.GetServiceInfo().GetServiceName()))
	}()
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}

//Launch 纯粹的启动微服务,不做系统信令监听
func Launch(ctx context.Context, service Service, serviceConfig *ServiceConfig, configPath string) (MicroService, errors.Error) {
	logger := logs.GetLogger(ctx)
	errorService := errors.GetRootErrorService(ctx)
	startTime := time.Now()
	if service == nil {
		return nil, errorService.BuildSystemError("serviceboot is nil")
	}

	ctx = boot.SetServiceInfo(ctx, serviceConfig.ServiceInfo)
	microService, err := newMicroService(ctx, service)
	if err != nil {
		return nil, err
	}
	err = microService.Start(ctx, serviceConfig, configPath)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("核心服务启动成功,服务地址:%s,启动耗时:%s", serviceConfig.ServerAddr, time.Since(startTime)))
	return microService, nil
}
