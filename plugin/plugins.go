package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
)

var serviceImpls = make(map[interface{}]interface{})

var AfterPluginStartedHandler func() error = nil

var plugins = make(map[string]Plugin, 0)
var sortPlugins = make([]Plugin, 0)
var _plugins = make(map[Plugin]string, 0)
var mutex = new(sync.RWMutex)

// startedPlugins 只保存 Start 已成功且尚未交给停止流程的插件。
// 生命周期由 engine 串行推进，mutex 保护成功集合的登记和一次性提取。
var startedPlugins []Plugin

// startupRollbackTimeout 是整轮回滚的协作式超时，不中断忽略 context 的 Stop。
const startupRollbackTimeout = 30 * time.Second

// Plugin 定义由 engine 串行调用的生命周期；Start 返回错误或 panic 的实例自行清理部分初始化资源。
type Plugin interface {
	// Start 返回 nil 后，框架才接管该实例的 Stop 责任；panic 被转换为启动错误。
	Start(ctx context.Context) error
	// Stop 释放成功启动的资源；返回错误或 panic 均不会阻止其他插件继续停止。
	Stop(ctx context.Context) error
}

func GetPluginByName(name string) interface{} {
	return plugins[name]
}

func RegisterPlugin(name string, service interface{}) {
	if service == nil {
		log.Panic("服务为空，不能注册", zap.String("name", name))
	}
	mutex.Lock()
	defer mutex.Unlock()
	if plugins[name] != nil {
		log.DPanic("插件已经注册过,不能重复注册!!!", zap.String("name", name))
	}
	var plugin Plugin = nil
	if _, ok := service.(Plugin); ok {
		plugin = service.(Plugin)
	} else {
		plugin = &pluginImpl{
			service: service,
		}
	}
	plugins[name] = plugin
	sortPlugins = append(sortPlugins, plugin)
	_plugins[plugin] = name
	log.Debug("插件注册", zap.String("plugin", name))
}

// StartPlugins 按注册顺序串行启动插件；失败时逆序回滚成功集合，保留启动和清理错误。
// 调用方须先完成注册，且不得与其他启动或停止流程并发调用。
func StartPlugins(ctx context.Context) (startErr error) {
	return startPlugins(ctx, nil)
}

// StartPluginsWithCancel 在插件启动失败时先取消由调用方持有的运行根，再逆序回滚成功集合。
// cancel 必须属于传入 ctx 对应运行根的 owner；nil 时行为与 StartPlugins 相同。
func StartPluginsWithCancel(ctx context.Context, cancel context.CancelFunc) (startErr error) {
	return startPlugins(ctx, cancel)
}

// startPlugins 驱动插件启动和失败回滚；需要先取消 owner 根的调用方通过 cancel 显式交付其取消权。
func startPlugins(ctx context.Context, cancel context.CancelFunc) (startErr error) {
	phase := "插件启动"
	defer func() {
		if recovered := recover(); recovered != nil {
			startErr = lifecyclePanicError(phase, recovered)
		}
		if startErr != nil {
			if cancel != nil {
				cancel()
			}
			// 启动取消或超时不能阻止资源清理，同时保留 context 中的链路信息。
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), startupRollbackTimeout)
			defer cancel()
			if cleanupErr := stopStartedPlugins(cleanupCtx); cleanupErr != nil {
				startErr = errors.Join(startErr, cleanupErr)
			}
		}
	}()
	for _, plugin := range sortPlugins {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("插件启动已取消: %w", err)
		}
		name := _plugins[plugin]
		phase = "启动插件 " + name
		log.Info("开始启动插件", zap.String("pluginName", name))
		err := plugin.Start(ctx)
		if err != nil {
			return fmt.Errorf("启动插件 %s: %w", name, err)
		}
		mutex.Lock()
		startedPlugins = append(startedPlugins, plugin)
		mutex.Unlock()
		log.Info("启动插件成功", zap.String("pluginName", name))
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("插件启动已取消: %w", err)
		}
	}
	if AfterPluginStartedHandler != nil {
		phase = "业务启动回调"
		err := AfterPluginStartedHandler()
		if err != nil {
			return fmt.Errorf("业务启动回调: %w", err)
		}
	}
	return nil
}

// lifecyclePanicError 将生命周期回调 panic 转为错误，并保留 error 类型的原始原因供 errors.Is 检查。
func lifecyclePanicError(phase string, recovered any) error {
	if cause, ok := recovered.(error); ok {
		return fmt.Errorf("%s panic: %w", phase, cause)
	}
	return fmt.Errorf("%s panic: %v", phase, recovered)
}

// StopPlugins 逆序停止成功启动的插件并记录清理错误；重复调用不会再次停止同一批插件。
// 该兼容接口不返回错误；需要将清理失败传回启动入口时使用 StopPluginsWithError。
func StopPlugins(ctx context.Context) {
	_ = stopStartedPlugins(ctx)
}

// StopPluginsWithError 逆序停止成功启动的插件并返回聚合清理错误。
// ctx 是调用方提供的 shutdown context，原样传给每个插件；取消和 deadline 需要各插件主动响应。
// 插件停止失败不会阻止其余插件继续停止，重复调用返回 nil 且不会再次停止同一批插件。
func StopPluginsWithError(ctx context.Context) error {
	return stopStartedPlugins(ctx)
}

// stopStartedPlugins 统一正常关闭和启动回滚，先提取清理责任，再调用插件以避免重复 Stop。
func stopStartedPlugins(ctx context.Context) error {
	mutex.Lock()
	stoppingPlugins := startedPlugins
	startedPlugins = nil
	pluginNames := make(map[Plugin]string, len(_plugins))
	for plugin, name := range _plugins {
		pluginNames[plugin] = name
	}
	mutex.Unlock()

	// 依赖插件先注册并先启动，停止时必须逆序释放使用方和依赖方。
	var cleanupErr error
	for index := len(stoppingPlugins) - 1; index >= 0; index-- {
		currentPlugin := stoppingPlugins[index]
		name := pluginNames[currentPlugin]
		err := stopPlugin(ctx, currentPlugin)
		if err != nil {
			log.Error("关闭插件失败", zap.String("pluginName", name), zap.Error(err))
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("关闭插件 %s: %w", name, err))
			continue
		}
		log.Info("关闭插件", zap.String("pluginName", name))
	}
	return cleanupErr
}

// stopPlugin 隔离单个插件的停止 panic，防止它中断其余依赖的逆序清理。
func stopPlugin(ctx context.Context, currentPlugin Plugin) (stopErr error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			stopErr = lifecyclePanicError("插件停止", recovered)
		}
	}()
	return currentPlugin.Stop(ctx)
}

type pluginImpl struct {
	service interface{}
	start   func(ctx context.Context) error
	stop    func(ctx context.Context) error
}

func (impl *pluginImpl) Start(ctx context.Context) error {
	if impl.start != nil {
		return impl.start(ctx)
	}
	return nil
}
func (impl *pluginImpl) Stop(ctx context.Context) error {
	if impl.stop != nil {
		return impl.stop(ctx)
	}
	return nil
}
