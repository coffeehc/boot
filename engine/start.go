package engine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func StartEngine(ctx context.Context, serviceInfo configuration.ServiceInfo, loadPlugins func(ctx context.Context), start func(ctx context.Context, cmd *cobra.Command, args []string) error) {
	InitService(ctx, serviceInfo)
	var rootCmd = &cobra.Command{
		Use:   configuration.GetServiceName(),
		Short: fmt.Sprintf("%s 服务", configuration.GetServiceName()),
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			var sigChan = make(chan os.Signal, 3)
			loadPlugins(ctx)
			defer plugin.StopPlugins(ctx)
			go func() {
				defer func() {
					if e := recover(); e != nil {
						err := errors.ConverUnknowError(e)
						log.DPanic("程序捕获不能处理的异常", err.GetFieldsWithCause()...)
						sigChan <- syscall.SIGKILL
					}
				}()
				start(ctx, cmd, args)
				plugin.StartPlugins(ctx)
				log.Info("服务启动完成")
			}()
			go func() {
				<-ctx.Done()
				sigChan <- syscall.SIGINT
			}()
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
			sig := <-sigChan
			log.Info("关闭程序", zap.Any("signal", sig))
			return nil
		},
	}
	_err := rootCmd.Execute()
	if _err != nil {
		log.Error("启动错误", zap.Error(_err))
		os.Exit(-1)
	}
}
