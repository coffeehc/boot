package server

import (
	"os"
	"os/signal"
	"syscall"

	"flag"
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

var (
	port = flag.Int("port", 8888, "服务端口")
)

/**
 *	Service 启动
 */
func ServiceLauncher(service base.Service, serviceDiscoveryRegedit ServiceDiscoveryRegister) {
	logger.InitLogger()
	defer logger.WaitToClose()
	if service == nil {
		logger.Error("service is nil")
		return
	}
	err := startService(service)
	defer func(service base.Service) {
		if service != nil && service.Stop != nil {
			stopErr := service.Stop()
			if stopErr != nil {
				fmt.Printf("关闭服务失败,%s\n", stopErr)
				os.Exit(-1)
			}
		}
	}(service)
	if err != nil {
		return
	}
	//if serviceDiscoveryRegedit == nil{
	//	serviceDiscoveryRegedit,_ = consultool.NewConsulServiceRegister(nil)
	//}
	micorService, err := newMicorService(service, serviceDiscoveryRegedit)
	err = micorService.Start()
	if err != nil {
		logger.Error("启动微服务出错:%s", err)
		return
	}
	waitStop()
	if service != nil && service.Stop != nil {
		stopErr := service.Stop()
		if stopErr != nil {
			fmt.Printf("关闭服务失败,%s\n", stopErr)
		}
	}
}

func startService(service base.Service) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			fmt.Printf("service crash,cause is %s\n", err1)
			err = fmt.Errorf("service crash,cause is %s", err1)
		}
	}()
	if service == nil {
		panic("没有 Service 的实例")
	}
	if service.Run == nil {
		panic("没有指定Run方法")
	}
	err1 := service.Run()
	if err1 != nil {
		panic(logger.Error("服务运行错误:%s", err1))
	}
	logger.Info("服务已正常启动")
	return
}

/*
	wait,一般是可执行函数的最后用于阻止程序退出
*/
func waitStop() {
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-sigChan
	logger.Debug("接收到指令:%s,立即关闭程序", sig)
}
