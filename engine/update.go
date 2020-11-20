package engine

import (
	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/spf13/cobra"
)

func buildUpdataCmd(serviceInfo configuration.ServiceInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "升级服务程序",
		Long:  "自动升级服务程序",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("自动升级暂时不可用")
			// workDir, _ := os.Getwd()
			// applicationPath,err:= filepath.Abs(os.Args[0])
			// if err!=nil{
			//   log.Fatal("转化程序路径失败",zap.Error(err))
			//   return err
			// }
			// params := map[string]string{
			//   "ApplicationPath": applicationPath,
			//   "ServiceName":     serviceInfo.ServiceName,
			//   "WorkDir":         workDir,
			// }
			// temp, err := template.New("serviceTemp").Parse(serviceTemp)
			// if err != nil {
			//   log.Fatal("解析模版错误", zap.Error(err))
			//   return err
			// }
			// serviceFile, err := os.OpenFile(path.Join(workDir, fmt.Sprintf("%s.service", serviceInfo.ServiceName)), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
			// if err != nil {
			//   log.Fatal("创建service文件失败", zap.Error(err))
			//   return err
			// }
			// err = temp.Execute(serviceFile, params)
			// if err != nil {
			//   log.Fatal("写入service文件失败", zap.Error(err))
			// }
			// return err
			return nil
		},
	}
}
