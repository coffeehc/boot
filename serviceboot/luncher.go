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
func ServiceLaunch(serviceName string, ctx context.Context, service Service) {
	if !flag.Parsed() {
		flag.Parse()
	}
	ctx = context.WithValue(ctx, boot.Ctx_Key_serviceName, serviceName)
	var errorService = errors.NewService(serviceName)
	ctx = errors.SetRootErrorService(ctx, errorService)
	logger, _ := zap.NewDevelopment()
	logService, err := logs.NewService()
	if err != nil {
		logger.Panic("创建logService失败", zap.String(logs.K_Cause, err.Error()))
		return
	}
	logger = logService.GetLogger()
	ctx = logs.SetLoggerService(ctx, logService)
	ctx = logs.SetLogger(ctx, logger)
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if boot.IsDevModule() {
		logger.Debug("当前为:开发模式")
	} else {
		logger.Debug("当前为:生产模式")
	}
	microService, err := Launch(ctx, service)
	if err != nil {
		logger.Error("启动服务失败", zap.String(logs.K_ServiceName, serviceName), zap.String(logs.K_Cause, err.Error()))
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
func Launch(ctx context.Context, service Service) (MicroService, errors.Error) {
	logger := logs.GetLogger(ctx)
	errorService := errors.GetRootErrorService(ctx)
	startTime := time.Now()
	if service == nil {
		return nil, errorService.BuildSystemError("serviceboot is nil")
	}
	serviceConfig, configPath, err := loadServiceConfig(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, boot.Ctx_Key_serviceInfo, serviceConfig.ServiceInfo)
	microService, err := newMicroService(ctx, service)
	if err != nil {
		return nil, err
	}
	err = microService.Start(ctx, serviceConfig, configPath)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("核心服务启动成功,服务地址:%s,启动耗时:%s", serviceConfig.ServerAddr, time.Since(startTime)))
	//注册是在服务完全启动之后
	deregisterFunc, err := serviceDiscoverRegister(ctx, microService.GetService(), serviceConfig.ServiceInfo, serviceConfig)
	if err != nil {
		err.PrintLog(ctx)
		return nil, err
	}
	microService.AddCleanFunc(deregisterFunc)
	return microService, nil
}

func serviceDiscoverRegister(ctx context.Context, service Service, serviceInfo ServiceInfo, serviceConfig *ServiceConfig) (func(), errors.Error) {
	errorService := errors.GetRootErrorService(ctx)
	serviceDiscoveryRegister, err := service.GetServiceDiscoveryRegister()
	if err != nil {
		return nil, errorService.BuildSystemError("获取serviceDiscoveryRegister失败", zap.String(logs.K_ServiceName, serviceInfo.GetServiceName()))
	}

	if !serviceConfig.DisableServiceRegister {
		if serviceDiscoveryRegister == nil {
			return nil, errorService.BuildSystemError("没有指定serviceDiscoveryRegister", zap.String(logs.K_ServiceName, serviceInfo.GetServiceName()))
		}
		deregister, registerError := serviceDiscoveryRegister.RegService(ctx, serviceInfo, serviceConfig.ServerAddr)
		if registerError != nil {
			return nil, errorService.BuildSystemError("注册服务失败", zap.String(logs.K_ServiceName, serviceInfo.GetServiceName()), zap.String(logs.K_Cause, err.Error()))
		}
		return deregister, nil
	}
	return func() {}, nil
}
