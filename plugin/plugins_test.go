package plugin

import (
	"context"
	"reflect"
	"testing"
)

// lifecycleTestPlugin 记录插件停止调用，用于验证依赖释放顺序。
type lifecycleTestPlugin struct {
	// name 是写入停止记录的插件名称。
	name string
	// stopOrder 保存插件实际停止顺序。
	stopOrder *[]string
}

func (plugin *lifecycleTestPlugin) Start(context.Context) error {
	return nil
}

func (plugin *lifecycleTestPlugin) Stop(context.Context) error {
	*plugin.stopOrder = append(*plugin.stopOrder, plugin.name)
	return nil
}

func TestStopPluginsUsesReverseRegistrationOrder(t *testing.T) {
	mutex.Lock()
	originalPlugins := plugins
	originalSortPlugins := sortPlugins
	originalPluginNames := _plugins
	plugins = make(map[string]Plugin)
	sortPlugins = make([]Plugin, 0)
	_plugins = make(map[Plugin]string)
	mutex.Unlock()
	t.Cleanup(func() {
		mutex.Lock()
		plugins = originalPlugins
		sortPlugins = originalSortPlugins
		_plugins = originalPluginNames
		mutex.Unlock()
	})

	stopOrder := make([]string, 0, 3)
	RegisterPlugin("dependency", &lifecycleTestPlugin{name: "dependency", stopOrder: &stopOrder})
	RegisterPlugin("consumer", &lifecycleTestPlugin{name: "consumer", stopOrder: &stopOrder})
	RegisterPlugin("entrypoint", &lifecycleTestPlugin{name: "entrypoint", stopOrder: &stopOrder})

	StopPlugins(t.Context())

	expected := []string{"entrypoint", "consumer", "dependency"}
	if !reflect.DeepEqual(stopOrder, expected) {
		t.Fatalf("插件停止顺序错误: got=%v want=%v", stopOrder, expected)
	}
}
