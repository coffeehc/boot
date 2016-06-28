package serviceboot

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"log"
	"fmt"
)

/**
 *	Service 启动
 */
func ServiceLauncher(service base.Service) {
	log.SetFlags(log.Ldate|log.Ltime|log.Llongfile)
	logger.InitLogger()
	defer logger.WaitToClose()
	if service == nil {
		logger.Error("service is nil")
		return
	}
	microService, err := newMicroService(service)
	if err != nil {
		log.Printf("初始化微服务出错:%s\n", err)
		return
	}
	err = microService.init()
	if err != nil {
		log.Printf("初始化微服务出错:%s\n", err)
		return
	}
	err = startService(service)
	defer func(service base.Service) {
		if service != nil && service.Stop != nil {
			stopErr := service.Stop()
			if stopErr != nil {
				log.Printf("关闭服务失败,%s\n", stopErr)
				os.Exit(-1)
			}
		}
	}(service)
	if err != nil {
		log.Printf("start service error,%s\n", err)
		return
	}
	err = microService.start()
	if err != nil {
		logger.Error("service start fail. %s", err)
		time.Sleep(time.Second)
		os.Exit(-1)
	}
	defer base.DebugPanic(true)
	defer func() {
		fmt.Printf("服务[%s]关闭\n",service.GetServiceInfo().GetServiceName())
	}()
	waitStop()
}
func startService(service base.Service) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
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
	fmt.Printf("接收到指令:%s,立即关闭程序", sig)
}
