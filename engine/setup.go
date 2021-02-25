package engine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func buildSetupCmd(serviceInfo configuration.ServiceInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "安装服务,适用与systemd",
		Long:  "生成对应的systemc服务描述文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, _ := os.Getwd()
			applicationPath, err := filepath.Abs(os.Args[0])
			if err != nil {
				log.Panic("转化程序路径失败", zap.Error(err))
				return err
			}
			params := map[string]string{
				"ApplicationPath": applicationPath,
				"ServiceName":     serviceInfo.ServiceName,
				"WorkDir":         workDir,
			}
			temp, err := template.New("serviceTemp").Parse(serviceTemp)
			if err != nil {
				log.Panic("解析模版错误", zap.Error(err))
				return err
			}
			serviceFile, err := os.OpenFile(path.Join(workDir, fmt.Sprintf("%s.service", serviceInfo.ServiceName)), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
			if err != nil {
				log.Panic("创建service文件失败", zap.Error(err))
				return err
			}
			err = temp.Execute(serviceFile, params)
			if err != nil {
				log.Panic("写入service文件失败", zap.Error(err))
			}
			return err
		},
	}
}

var serviceTemp = `[Unit]
Description="{{.ServiceName}}"
Requires=network-online.target
After=network-online.target

[Service]
ExecStart={{.ApplicationPath}} start
#ExecReload={{.ApplicationPath}} reload
EnvironmentFile=/data/.env_setting
KillMode=process
Restart=always
RestartSec=5
WorkingDirectory={{.WorkDir}}
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target`
