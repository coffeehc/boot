package engine

import (
	"fmt"
	"strings"

	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
)

// EngineOption 用于配置 StartEngineWithOptions 的可选能力。
// 通过函数式参数可以在不破坏 StartEngine 现有调用方式的前提下，
// 逐步扩展 engine 的启动行为（例如挂载额外 CLI 子命令）。
type EngineOption func(*engineOptions) error

// WithExtraCommands 用于向 engine 注入额外的根级 CLI 子命令。
// 该能力适用于业务方按需扩展命令入口（例如注入 cli 子命令），
// 不会改变任何内建命令的注册与执行逻辑。
func WithExtraCommands(extraCommands ...*cobra.Command) EngineOption {
	return func(options *engineOptions) error {
		options.extraCommands = append(options.extraCommands, extraCommands...)
		return nil
	}
}

// WithFileOnlyConfiguration 要求每轮 start 初始化仅从配置文件读取资源值。
// 配置文件仍由 config CLI 参数选择；未启用此选项的应用保留 ENV_ 覆盖行为。
func WithFileOnlyConfiguration() EngineOption {
	return func(options *engineOptions) error {
		options.configurationOptions = append(options.configurationOptions, configuration.WithFileOnly())
		return nil
	}
}

// engineOptions 保存 StartEngineWithOptions 的内部配置结果。
// 选项绑定到本次命令装配，避免文件模式影响后续默认 engine。
type engineOptions struct {
	// extraCommands 是应用显式挂载的额外根命令。
	extraCommands []*cobra.Command
	// configurationOptions 是每轮 start 交给配置 owner 的初始化策略。
	configurationOptions []configuration.Option
}

// collectEngineOptions 按顺序执行所有 EngineOption，并聚合最终配置。
// 若任一 option 返回错误，则立即中止并返回该错误。
func collectEngineOptions(opts ...EngineOption) (engineOptions, error) {
	options := engineOptions{
		extraCommands: make([]*cobra.Command, 0),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&options); err != nil {
			return engineOptions{}, err
		}
	}
	return options, nil
}

// validateExtraCommands 校验扩展命令是否可安全挂载。
// 校验规则包含：
// 1) 不能为 nil；
// 2) 命令名不能为空；
// 3) 不能与内建命令重名；
// 4) 扩展命令之间不能重名。
func validateExtraCommands(extraCommands []*cobra.Command, builtinCommandNames map[string]struct{}) error {
	extraCommandNames := make(map[string]struct{}, len(extraCommands))
	for _, extraCommand := range extraCommands {
		if extraCommand == nil {
			return fmt.Errorf("扩展命令不能为空")
		}
		name := strings.TrimSpace(extraCommand.Name())
		if name == "" {
			return fmt.Errorf("扩展命令 %q 没有有效命令名", extraCommand.Use)
		}
		if _, exists := builtinCommandNames[name]; exists {
			return fmt.Errorf("扩展命令 %q 与内建命令重名", name)
		}
		if _, exists := extraCommandNames[name]; exists {
			return fmt.Errorf("扩展命令 %q 重复定义", name)
		}
		extraCommandNames[name] = struct{}{}
	}
	return nil
}
