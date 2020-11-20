package engine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.xiagaogao.com/coffee/base/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"git.xiagaogao.com/coffee/boot/configuration"
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
			configuration.PrintVersionInfo()
			fmt.Println()
			cmd.Help()

		},
	}
	rootCmd.AddCommand(
		versionCmd,
		buildServiceCmd(ctx, serviceInfo, start),
		buildSetupCmd(serviceInfo),
		buildUpdataCmd(serviceInfo),
	)
	_err := rootCmd.ExecuteContext(ctx)
	if _err != nil {
		log.Error("启动错误", zap.Error(_err))
		os.Exit(-1)
	}
}
