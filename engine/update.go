package engine

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func buildUpdateCmd(serviceInfo configuration.ServiceInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "升级服务程序",
		Long:  "自动升级服务程序",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("自动升级暂时不可用")
			// workDir, _ := os.Getwd()
			// applicationPath,err:= filepath.Abs(os.Args[0])
			// if err!=nil{
			//   log.Panic("转化程序路径失败",zap.Error(err))
			//   return err
			// }
			// params := map[string]string{
			//   "ApplicationPath": applicationPath,
			//   "ServiceName":     serviceInfo.ServiceName,
			//   "WorkDir":         workDir,
			// }
			// temp, err := template.New("serviceTemp").Parse(serviceTemp)
			// if err != nil {
			//   log.Panic("解析模版错误", zap.Error(err))
			//   return err
			// }
			// serviceFile, err := os.OpenFile(path.Join(workDir, fmt.Sprintf("%s.service", serviceInfo.ServiceName)), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
			// if err != nil {
			//   log.Panic("创建service文件失败", zap.Error(err))
			//   return err
			// }
			// err = temp.Execute(serviceFile, params)
			// if err != nil {
			//   log.Panic("写入service文件失败", zap.Error(err))
			// }
			// return err
			return nil
		},
	}
}

func UpdateSalf(serviceInfo configuration.ServiceInfo, downloadUrl string) errors.Error {
	serviceSoftPath, e := filepath.Abs(os.Args[0])
	if e != nil {
		log.Error("获取程序绝对路径失败", zap.Error(e))
		return errors.MessageError("获取程序绝对路径失败")
	}
	baseDir := filepath.Dir(serviceSoftPath)
	updateDir := filepath.Join(baseDir, "updates")
	e = os.MkdirAll(updateDir, 0666)
	if e != nil {
		log.Error("创建更新目录失败", zap.Error(e))
		return errors.MessageError("创建更新目录失败")
	}
	oldFiles := make([]string, 0)
	// TODO 开始遍历文件夹
	fs.WalkDir(os.DirFS(updateDir), ".", func(path string, d fs.DirEntry, err error) error {
		oldFiles = append(oldFiles, path)
		return nil
	})
	if len(oldFiles) > 5 {
		for i, path := range oldFiles {
			if i < 3 {
				os.Remove(path)
			}
		}
	}
	log.Debug("创建了升级目录", zap.String("updtaDir", updateDir))
	updateFileName := fmt.Sprintf("%s%s%s", updateDir, filepath.Separator, fmt.Sprintf("%s_%d", serviceInfo.ServiceName, time.Now().UnixNano()))
	log.Debug("创建新版本文件", zap.String("updateFileName", updateFileName))
	updateFile, e := os.OpenFile(updateFileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	if e != nil {
		log.Error("创建下载文件失败", zap.Error(e))
		return errors.MessageError("创建下载文件失败")
	}
	log.Debug("开始下载程序")
	resp, e := http.Get(downloadUrl)
	if e != nil {
		log.Error("请求下载地址失败", zap.Error(e))
		return errors.MessageError("请求下载地址失败")
	}
	log.Debug("开始保存程序")
	n, e := io.Copy(updateFile, resp.Body)
	resp.Body.Close()
	updateFile.Sync()
	updateFile.Close()
	if e != nil {
		log.Error("写入文件失败", zap.Error(e))
		return errors.MessageError("写入文件失败")
	}
	log.Debug("写入文件大小", zap.Int64("size", n))
	log.Debug("开始转储程序")
	os.Rename(serviceSoftPath, filepath.Join(updateDir, fmt.Sprintf("%s_%s", serviceInfo.ServiceName, strings.ReplaceAll(serviceInfo.Version, ".", "_"))))
	os.Rename(updateFileName, serviceSoftPath)
	log.Debug("开始启动后台守护进程")
	daemon(serviceInfo)
	return nil
}
