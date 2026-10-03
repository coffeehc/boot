package engine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ServiceStart func(ctx context.Context, cmd *cobra.Command, args []string) (ServiceCloseCallback, error)
type ServiceCloseCallback func()

// shutdownTimeout 是正常停机清理的协作式期限；不响应 context 的回调无法被强制中断。
const shutdownTimeout = 30 * time.Second

var rootContextState struct {
	sync.RWMutex
	ctx context.Context
}

// GetRootContext 返回当前 start 命令运行期间 Boot 持有的运行根 context。
// ServiceStart、插件初始化与插件 Start 收到的均是此 context。业务方只能基于它派生任务 context，
// 不应自行取消此根。未处于已开始的服务运行期间时返回 nil；重复并发启动会被拒绝。
// 命令退出后该运行根不再可获取；下一次独立启动会创建新的根。
func GetRootContext() context.Context {
	rootContextState.RLock()
	defer rootContextState.RUnlock()
	return rootContextState.ctx
}

// beginRunContext 创建并登记一次服务运行根；有另一轮运行仍在进行时返回错误。
func beginRunContext(parent context.Context) (context.Context, context.CancelFunc, func(), error) {
	rootContextState.Lock()
	defer rootContextState.Unlock()
	if rootContextState.ctx != nil {
		return nil, nil, nil, fmt.Errorf("服务已经在运行")
	}
	ctx, cancel := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	rootContextState.ctx = ctx
	release := func() {
		cancel()
		rootContextState.Lock()
		if rootContextState.ctx == ctx {
			rootContextState.ctx = nil
		}
		rootContextState.Unlock()
	}
	return ctx, cancel, release, nil
}

// WaitServiceStop 等待信号或调用方取消，执行关闭回调后注销本次信号通知。
func WaitServiceStop(ctx context.Context, closeCallback func()) {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer signal.Stop(sigChan)
	var sig any
	select {
	case sig = <-sigChan:
	case <-ctx.Done():
		sig = ctx.Err()
	}
	log.Debug("收到指令", zap.Any("signal", sig))
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
