package plugin

import (
	"context"
	"sync"

	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
)

var plugins = make(map[string]Plugin, 0)
var sortPlugins = make([]Plugin, 0)
var _plugins = make(map[Plugin]string, 0)
var mutex = new(sync.RWMutex)

type Plugin interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func RegisterPlugin(name string, service interface{}) {
	if service == nil {
		log.Panic("服务为空，不能注册", zap.String("name", name))
	}
	mutex.Lock()
	defer mutex.Unlock()
	if plugins[name] != nil {
		log.Warn("插件已经注册过,不能重复注册", zap.String("name", name))
	}
	var plugin Plugin = nil
	if _, ok := service.(Plugin); !ok {
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

func StartPlugins(ctx context.Context) {
	for _, plugin := range sortPlugins {
		name := _plugins[plugin]
		log.Info("开始启动插件", zap.String("pluginName", name))
		err := plugin.Start(ctx)
		if err != nil {
			log.Panic("启动插件失败", zap.String("pluginName", name), zap.Error(err))
			// continue
		}
		log.Info("启动插件成功", zap.String("pluginName", name))
	}
}

func StopPlugins(ctx context.Context) {
	for name, plugin := range plugins {
		err := plugin.Stop(ctx)
		if err != nil {
			log.Error("启动插件失败", zap.String("pluginName", name), zap.Error(err))
			continue
		}
		log.Info("关闭插件", zap.String("pluginName", name))
	}
}

//func RegisterPluginByFast(name string, start func(ctx context.Context) error, stop func(ctx context.Context) error) {
//	RegisterPlugin(name, &pluginImpl{
//		stop:  stop,
//		start: start,
//	})
//}

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
