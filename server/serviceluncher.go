package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/common"
)

/**
 *	Service 启动
 */
func ServiceLauncher(service common.Service, serviceDiscoveryRegedit ServiceDiscoveryRegister) {
	logger.InitLogger()
	defer logger.WaitToClose()
	if service == nil {
		logger.Error("service is nil")
		return
	}
	startService(service)
	micorService, err := newMicorService(service, serviceDiscoveryRegedit)
	if err != nil {

	}
	err = micorService.Start()
	if err != nil {
		logger.Error("启动微服务出错:%s", err)
		return
	}
	waitStop()
}

func startService(service common.Service) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("service crash,cause is %s", err)
		}
		if service != nil && service.Stop != nil {
			stopErr := service.Stop()
			if stopErr != nil {
				logger.Error("关闭服务失败,%s", stopErr)
			}
		}
	}()
	if service == nil {
		panic("没有 Service 的实例")
	}
	if service.Run == nil {
		panic("没有指定Run方法")
	}
	err := service.Run()
	if err != nil {
		panic(logger.Error("服务运行错误:%s", err))
	}
	logger.Info("服务已正常启动")
}

//func stop()  {
//
//}

/*
	wait,一般是可执行函数的最后用于阻止程序退出
*/
func waitStop() {
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-sigChan
	logger.Debug("接收到指令:%s,立即关闭程序", sig)
}
