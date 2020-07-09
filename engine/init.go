package engine

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"git.xiagaogao.com/coffee/base/log"
	"go.uber.org/zap"

	"git.xiagaogao.com/coffee/boot/configuration"
)

func initService(ctx context.Context, serviceInfo configuration.ServiceInfo) {
	configuration.InitConfiguration(ctx, serviceInfo)
}

func WaitServiceStop(ctx context.Context, cancelFunc context.CancelFunc, closeCallback func()) {
	var sigChan = make(chan os.Signal, 3)
	go func() {
		<-ctx.Done()
		sigChan <- syscall.SIGINT
	}()
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-sigChan
	if ctx.Err() == nil {
		cancelFunc()
	}
	if closeCallback != nil {
		closeCallback()
	}
	log.Info("关闭程序", zap.Any("signal", sig))
}
