package engine

import (
	"context"
	"fmt"
	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var daemonModel = pflag.Bool("daemon", false, "开启守护进程模式")

//func readPid(serviceInfo configuration.ServiceInfo) int {
//	pidFile := fmt.Sprintf("/var/run/%s.pid", serviceInfo.ServiceName)
//	pidData, _ := os.ReadFile(pidFile)
//	pid, _ := strconv.ParseInt(string(pidData), 10, 64)
//	return int(pid)
//}
//
//func savePid(serviceInfo configuration.ServiceInfo, pid int) {
//	pidFile := fmt.Sprintf("/var/run/%s.pid", serviceInfo.ServiceName)
//	os.WriteFile(pidFile, []byte(strconv.FormatInt(int64(pid), 10)), 0644)
//}

func buildStartCmd(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) *cobra.Command {
	ctx, cancelFunc := context.WithCancel(ctx)
	return &cobra.Command{
		Use:   "start",
		Short: "启动服务",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("守护模式:%t\n", *daemonModel)
			log.Debug("守护模式", zap.Bool("daemonModel", *daemonModel))
			var daemonContext *daemon.Context
			if *daemonModel {
				daemonContext = getDaemonContext(serviceInfo)
			}
			configuration.InitConfiguration(ctx, serviceInfo)
			defer plugin.StopPlugins(ctx)
			var closeCallback ServiceCloseCallback = nil
			//go func() {
			_closeCallback, err := start(ctx, cmd, args)
			if err != nil {
				log.Error("启动服务失败", zap.Error(err))
				cancelFunc()
			}
			closeCallback = _closeCallback
			defer func() {
				if e := recover(); e != nil {
					err := errors.ConverUnknowError(e)
					log.DPanic("程序捕获不能处理的异常", err.GetFieldsWithCause()...)
					cancelFunc()
				}
			}()
			plugin.StartPlugins(ctx)
			log.Info("服务启动完成")
			if daemonContext != nil {
				child, err := daemonContext.Reborn()
				if err != nil {
					log.Error("错误", zap.Error(err))
				}
				if child != nil {
					return nil
				}
				defer daemonContext.Release()
			}
			WaitServiceStop(ctx, closeCallback)
			return nil
		},
	}
}
