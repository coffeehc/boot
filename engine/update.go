package engine

import (
	"fmt"
	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
)

func buildUpdateCmd(serviceInfo configuration.ServiceInfo) *cobra.Command {
	var downloadUrl = ""
	cmd := &cobra.Command{
		Use:   "update",
		Short: "升级服务程序",
		Long:  "自动升级服务程序",
		RunE: func(cmd *cobra.Command, args []string) error {
			applicationPath, err := os.Executable()
			if err != nil {
				log.Panic("转化程序路径失败", zap.Error(err))
				return err
			}
			resp, e := http.Get(downloadUrl)
			if e != nil {
				log.Error("请求下载地址失败", zap.Error(e))
				return errors.MessageError("请求下载地址失败")
			}
			backPath := fmt.Sprintf("%s.bak", applicationPath)
			tmpPath := fmt.Sprintf("%s.tmp", applicationPath)
			updateFile, e := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
			if e != nil {
				log.Error("创建下载文件失败", zap.Error(e))
				return errors.MessageError("创建下载文件失败")
			}
			_, e = io.Copy(updateFile, resp.Body)
			if e != nil {
				updateFile.Close()
				os.Remove(tmpPath)
				return e
			}
			resp.Body.Close()
			updateFile.Sync()
			updateFile.Close()
			os.Rename(applicationPath, backPath)
			os.Rename(tmpPath, applicationPath)
			pid := ReadPid(serviceInfo)
			Kill(pid)
			log.Info("安装新的程序完成")
			return err
		},
	}
	cmd.Flags().StringVarP(&downloadUrl, "download_url", "url", "", "升级文件下载地址")
	return cmd
}
