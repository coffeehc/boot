package serviceboot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

/**
 *	Service 启动
 */
func ServiceLauncher(service base.Service) {
	logger.InitLogger()
	defer logger.WaitToClose()
	if service == nil {
		logger.Error("service is nil")
		return
	}
	microService, err := newMicroService(service)
	if err != nil {
		fmt.Printf("初始化微服务出错:%s\n", err)
		return
	}
	err = microService.init()
	if err != nil {
		fmt.Printf("初始化微服务出错:%s\n", err)
		return
	}
	err = startService(service)
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
		fmt.Printf("start service error,%s\n", err)
		return
	}
	err = microService.start()
	if err != nil {
		logger.Error("service start fail. %s", err)
		time.Sleep(time.Second)
		os.Exit(-1)
	}
	waitStop()
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
