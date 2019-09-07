package serviceboot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

// serviceLaunch Service 启动
func ServiceLaunch(ctx context.Context, service Service) {
	ctx, cancelFunc := context.WithCancel(ctx)
	boot.InitFlags()
	boot.InitModel()
	time.FixedZone("CST", 8*3600)
	if pflag.Lookup("help") != nil {
		pflag.PrintDefaults()
		os.Exit(0)
	}
	logger, _ := zap.NewDevelopment()
	err1 := boot.CheckServiceInfoConfig(ctx, serviceInfo)
	if err1 != nil {
		logger.Error("校验服务信息失败", zap.Error(err1))
		return
	}

	ctx = boot.SetServiceName(ctx, serviceInfo.ServiceName)
	logger = xlog.GetLogger()
	errorService := xerror.NewService(serviceInfo.ServiceName)
	boot.PrintServiceInfo(serviceInfo, logger)
	ctx = boot.SetServiceName(ctx, serviceInfo.ServiceName)
	serviceConfig, configPath, err := loadServiceConfig(ctx, errorService, logger)
	if err != nil {
		logger.DPanic(err.Error(), err.GetFields()...)
		return
	}
	logger.Info("运行模式", zap.String("model", boot.RunModel()))
	microService, err := Launch(ctx, service, serviceInfo, serviceConfig, configPath)
	if err != nil {
		logger.DPanic(err.Error(), err.GetFields()...)
		return
	}
	defer func() {
		microService.Stop(ctx)
		logger.Info("服务关闭")
	}()
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	cancelFunc()
}

// Launch 纯粹的启动微服务,不做系统信令监听
func Launch(ctx context.Context, service Service, serviceInfo *boot.ServiceInfo, serviceConfig *ServiceConfig, configPath string) (MicroService, xerror.Error) {
	startTime := time.Now()
	if service == nil {
		return nil, xerror.SystemError("serviceboot is nil")
	}
	microService, err := newMicroService(ctx, service, serviceInfo, configPath, errorService, logger, loggerService)
	if err != nil {
		return nil, err
	}
	err = microService.Start(ctx, serviceConfig)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("核心服务启动成功,服务地址:%s,启动耗时:%s", serviceConfig.GetServiceEndpoint(), time.Since(startTime)))
	return microService, nil
}
