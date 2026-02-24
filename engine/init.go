package engine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ServiceStart func(ctx context.Context, cmd *cobra.Command, args []string) (ServiceCloseCallback, error)
type ServiceCloseCallback func()

func WaitServiceStop(ctx context.Context, closeCallback func()) {
	ctx, cancelFunc := context.WithCancel(ctx)
	var sigChan = make(chan os.Signal, 1)
	go func() {
		<-ctx.Done()
		sigChan <- syscall.SIGINT
	}()
	//if runtime.GOOS != "darwin" && runtime.GOOS != "ios" {
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
	//}
	sig := <-sigChan
	log.Debug("收到指令", zap.Any("signal", sig))
	if ctx.Err() == nil && cancelFunc != nil {
		cancelFunc()
	}
	if closeCallback != nil {
		closeCallback()
	}
	log.Info("关闭程序", zap.Any("signal", sig))
}

// StartEngine 是兼容入口，保留既有调用方式和默认行为。
// 内部会复用 StartEngineWithOptions，并在不传递任何 option 时维持原有语义。
func StartEngine(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) {
	StartEngineWithOptions(ctx, serviceInfo, start)
}

// StartEngineWithOptions 是 engine 的可扩展启动入口。
// 该入口在保持内建命令行为不变的前提下，允许通过 EngineOption 安全挂载自定义命令。
func StartEngineWithOptions(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart, opts ...EngineOption) {
	serviceInfo.Version = configuration.Version
	if serviceInfo.Version == "" {
		fmt.Printf("没有指定版本号")
		os.Exit(-1)
	}

	rootCmd, err := buildRootCommand(ctx, serviceInfo, start, opts...)
	if err != nil {
		log.Error("启动错误", zap.Error(err))
		os.Exit(-1)
	}
	if err = rootCmd.ExecuteContext(ctx); err != nil {
		log.Error("启动错误", zap.Error(err))
		os.Exit(-1)
	}
}

// buildRootCommand 负责构建并返回 engine 根命令。
// 该函数会先注册内建命令，再校验并挂载扩展命令，以便单元测试直接验证命令装配行为。
func buildRootCommand(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart, opts ...EngineOption) (*cobra.Command, error) {
	engineOpts, err := collectEngineOptions(opts...)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		Use:   configuration.GetServiceName(),
		Short: fmt.Sprintf("%s 服务", configuration.GetServiceName()),
		Long:  serviceInfo.Descriptor,
		Run: func(cmd *cobra.Command, args []string) {
			configuration.PrintVersionInfo()
			cmd.Help()
		},
	}

	builtinCommands := []*cobra.Command{
		buildVersionCmd(),
		buildReStartCmd(ctx, serviceInfo, start),
		buildStartCmd(ctx, serviceInfo, start),
		buildDaemonStartCmd(ctx, serviceInfo, start),
		buildStopCmd(ctx, serviceInfo),
		buildSetupCmd(serviceInfo),
	}
	builtinCommandNames := make(map[string]struct{}, len(builtinCommands))
	for _, command := range builtinCommands {
		builtinCommandNames[command.Name()] = struct{}{}
	}
	if err := validateExtraCommands(engineOpts.extraCommands, builtinCommandNames); err != nil {
		return nil, err
	}

	rootCmd.AddCommand(builtinCommands...)
	rootCmd.AddCommand(engineOpts.extraCommands...)
	return rootCmd, nil
}
