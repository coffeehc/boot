package engine

import (
	"context"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func buildStopCmd(ctx context.Context, serviceInfo configuration.ServiceInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "关闭服务",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			pid := ReadPid(serviceInfo)
			log.Info("关闭服务", zap.Int("pid", pid))
			return Kill(pid)
		},
	}
}
