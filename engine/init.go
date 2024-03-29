package engine

import (
	"context"
	"fmt"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type ServiceStart func(ctx context.Context, cmd *cobra.Command, args []string) (ServiceCloseCallback, error)
type ServiceCloseCallback func()

func WaitServiceStop(ctx context.Context, closeCallback func()) {
	ctx, cancelFunc := context.WithCancel(ctx)
	var sigChan = make(chan os.Signal, 1)
	go func() {
		<-ctx.Done()
		sigChan <- syscall.SIGINT
	}()
	//if runtime.GOOS != "darwin" && runtime.GOOS != "ios" {
	signal.Notify(sigChan,
		//syscall.SIGHUP,
		//syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		//syscall.SIGQUIT
	)
	//}
	sig := <-sigChan
	log.Debug("收到指令", zap.Any("signal", sig))
	if ctx.Err() == nil && cancelFunc != nil {
		cancelFunc()
	}
	if closeCallback != nil {
		closeCallback()
	}
	log.Info("关闭程序", zap.Any("signal", sig))
}

func StartEngine(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) {
	serviceInfo.Version = configuration.Version
	if serviceInfo.Version == "" {
		fmt.Printf("没有指定版本号")
		os.Exit(-1)
	}

	var rootCmd = &cobra.Command{
		Use:   configuration.GetServiceName(),
		Short: fmt.Sprintf("%s 服务", configuration.GetServiceName()),
		Long:  serviceInfo.Descriptor,
		Run: func(cmd *cobra.Command, args []string) {
			configuration.PrintVersionInfo()
			cmd.Help()
		},
	}
	rootCmd.AddCommand(
		buildVersionCmd(),
		buildReStartCmd(ctx, serviceInfo, start),
		buildStartCmd(ctx, serviceInfo, start),
		buildDaemonStartCmd(ctx, serviceInfo, start),
		buildStopCmd(ctx, serviceInfo),
		buildSetupCmd(serviceInfo),
	)
	_err := rootCmd.ExecuteContext(ctx)
	if _err != nil {
		log.Error("启动错误", zap.Error(_err))
		os.Exit(-1)
	}
}
