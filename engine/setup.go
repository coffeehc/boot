package engine

import (
	"fmt"
	"github.com/coffeehc/base/errors"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
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
			tl := ""
			servicePath := ""
			switch runtime.GOOS {
			case "freebsd", "netbsd":
				tl = bsdServiceTemp
				servicePath = serviceInfo.ServiceName //"/usr/local/etc/rc.d"
			case "linux":
				tl = linuxServiceTemp
				servicePath = fmt.Sprintf("%s.service", serviceInfo.ServiceName)
			default:
				return errors.MessageError("该系统不支持生成服务注册文件")
			}
			temp, err := template.New("serviceTemp").Parse(tl)
			if err != nil {
				log.Panic("解析模版错误", zap.Error(err))
				return err
			}
			serviceFile, err := os.OpenFile(path.Join(workDir, servicePath), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
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

var linuxServiceTemp = `[Unit]
Description="{{.ServiceName}}"
Requires=network-online.target
After=network-online.target

[Service]
ExecStart={{.ApplicationPath}} start
KillMode=process
Restart=always
RestartSec=5
WorkingDirectory={{.WorkDir}}
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target`

var bsdServiceTemp = `#!/bin/sh
. /etc/rc.subr

name={{.ServiceName}}
rcvar={{.ServiceName}}_enable

start_cmd="${name}_start"
stop_cmd="${name}_shutdown"

{{.ServiceName}}_start()
{
  echo "Starting supervising daemon."
  /usr/sbin/daemon -r -f -P "/var/run/${name}.pid" {{.ApplicationPath}} start --config={{.WorkDir}}/config.yml
}

{{.ServiceName}}_shutdown()
{
	if [ -e "/var/run/${name}.pid" ]; then
		echo "Stopping supervising daemon."
		kill -s TERM $(cat "/var/run/${name}.pid")
	fi
}


load_rc_config $name
run_rc_command "$1"
`
