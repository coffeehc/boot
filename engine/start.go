package engine

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func buildServiceCmd(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "启动服务",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancelFunc := context.WithCancel(ctx)
			configuration.InitConfiguration(ctx, serviceInfo)
			defer plugin.StopPlugins(ctx)
			var closeCallback ServiceCloseCallback = nil
			go func() {
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
			}()
			WaitServiceStop(ctx, cancelFunc, closeCallback)
			return nil
		},
	}
}

// func StartEngine1(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) {
//   var rootCmd = &cobra.Command{
//     Use:   "", // configuration.GetServiceName(),
//     Short: "",//fmt.Sprintf("%s 服务", configuration.GetServiceName()),
//     Long:  serviceInfo.Descriptor,
//     RunE: func(cmd *cobra.Command, args []string) error {
//       ctx, cancelFunc := context.WithCancel(ctx)
//       initService(ctx, serviceInfo)
//       defer plugin.StopPlugins(ctx)
//       var closeCallback ServiceCloseCallback = nil
//       go func() {
//         defer func() {
//           if e := recover(); e != nil {
//             err := errors.ConverUnknowError(e)
//             log.DPanic("程序捕获不能处理的异常", err.GetFieldsWithCause()...)
//             cancelFunc()
//           }
//         }()
//         _closeCallback, err := start(ctx, cmd, args)
//         if err != nil {
//           log.Error("启动服务失败", zap.Error(err))
//           cancelFunc()
//         }
//         closeCallback = _closeCallback
//         plugin.StartPlugins(ctx)
//         log.Info("服务启动完成")
//       }()
//       WaitServiceStop(ctx, cancelFunc, closeCallback)
//       return nil
//     },
//   }
//   rootCmd.AddCommand(versionCmd, helpCmd)
//   _err := rootCmd.Execute()
//   if _err != nil {
//     log.Error("启动错误", zap.Error(_err))
//     os.Exit(-1)
//   }
// }
