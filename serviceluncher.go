package microserviceboot

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/common"
	"github.com/coffeehc/web"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/**
 *	Service 启动
 */
func ServiceLauncher(service common.Service, config *web.ServerConfig) {
	logger.InitLogger()
	if service == nil {
		logger.Error("service is nil")
		os.Exit(-1)
	}
	startService(service)
	server := web.NewServer(config)
	regeditEndpoints()
	server.Start()
	waitStop()
}

func regeditEndpoints() {

}

func startService(service common.Service) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("service crash,cause is %s", err)
		}
		if service.Stop != nil {
			stopErr := service.Stop()
			if stopErr != nil {
				logger.Error("关闭服务失败,%s", stopErr)
			}
		}
		time.Sleep(time.Second)
	}()
	if service.Run == nil {
		panic("没有指定Run方法")
	}
	err := service.Run()
	if err != nil {
		panic(logger.Error("服务运行错误:%s", err))
	}
	logger.Info("服务已正常启动")
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
