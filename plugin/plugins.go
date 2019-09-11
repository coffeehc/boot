package plugin

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"go.uber.org/zap"
)

var plugins = make(map[string]Plugin, 0)
var mutex = new(sync.RWMutex)

type Plugin interface {
	Start(ctx context.Context) errors.Error
	Stop(ctx context.Context) errors.Error
}

func RegisterPlugin(name string, plugin Plugin) {
	mutex.Lock()
	defer mutex.Unlock()
	if plugins[name] != nil {
		log.Warn("插件已经注册过,不能重复注册", zap.String("name", name))
	}
	plugins[name] = plugin
	log.Debug("插件注册", zap.String("plugin", name))
}

func StartPlugins(ctx context.Context) {
	for name, plugin := range plugins {
		err := plugin.Start(ctx)
		if err != nil {
			log.Error("启动插件失败", err.GetFieldsWithCause(zap.String("pluginName", name))...)
			continue
		}
		log.Info("启动插件成功", zap.String("pluginName", name))
	}
}

func StopPlugins(ctx context.Context) {
	for name, plugin := range plugins {
		err := plugin.Stop(ctx)
		if err != nil {
			log.Error("启动插件失败", err.GetFieldsWithCause(zap.String("pluginName", name))...)
			continue
		}
		log.Info("关闭插件", zap.String("pluginName", name))
	}
}

func RegisterPluginByFast(name string, start func(ctx context.Context) errors.Error, stop func(ctx context.Context) errors.Error) {
	RegisterPlugin(name, &pluginImpl{
		stop:  stop,
		start: start,
	})
}

type pluginImpl struct {
	start func(ctx context.Context) errors.Error
	stop  func(ctx context.Context) errors.Error
}

func (impl *pluginImpl) Start(ctx context.Context) errors.Error {
	if impl.start != nil {
		return impl.start(ctx)
	}
	return nil
}
func (impl *pluginImpl) Stop(ctx context.Context) errors.Error {
	if impl.stop != nil {
		return impl.stop(ctx)
	}
	return nil
}
