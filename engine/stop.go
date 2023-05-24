package engine

import (
	"context"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"syscall"
)

func buildStopCmd(ctx context.Context, serviceInfo configuration.ServiceInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "关闭服务",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonContext := getDaemonContext(serviceInfo)
			pid, err := daemon.ReadPidFile(daemonContext.PidFileName)
			if err != nil {
				log.Error("错误", zap.Error(err))
				return err
			}
			log.Info("关闭服务", zap.Int("pid", pid))
			syscall.Kill(pid, syscall.SIGTERM)
			return nil
		},
	}
}
