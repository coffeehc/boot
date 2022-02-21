package engine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/coffeehc/base/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/coffeehc/boot/configuration"
)

type ServiceStart func(ctx context.Context, cmd *cobra.Command, args []string) (ServiceCloseCallback, error)
type ServiceCloseCallback func()

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

func StartEngine(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) {
	var rootCmd = &cobra.Command{
		Use:   configuration.GetServiceName(),
		Short: fmt.Sprintf("%s 服务", configuration.GetServiceName()),
		Long:  serviceInfo.Descriptor,
		Run: func(cmd *cobra.Command, args []string) {
			configuration.PrintVersionInfo(serviceInfo)
			fmt.Println()
			cmd.Help()
		},
	}
	rootCmd.AddCommand(
		buildVersionCmd(serviceInfo),
		buildStartCmd(ctx, serviceInfo, start),
		buildSetupCmd(serviceInfo),
		buildUpdateCmd(serviceInfo),
	)
	_err := rootCmd.ExecuteContext(ctx)
	if _err != nil {
		log.Error("启动错误", zap.Error(_err))
		os.Exit(-1)
	}
}
