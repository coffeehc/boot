package engine

import (
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
)

func buildVersionCmd(serviceInfo configuration.ServiceInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "版本信息",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			configuration.PrintVersionInfo(serviceInfo)
		},
	}
}
