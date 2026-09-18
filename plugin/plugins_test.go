package plugin

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// lifecycleTestPlugin 记录插件停止调用，用于验证依赖释放顺序。
type lifecycleTestPlugin struct {
	// name 是写入停止记录的插件名称。
	name string
	// stopOrder 保存插件实际停止顺序。
	stopOrder *[]string
	// startError 是测试插件启动时返回的错误。
	startError error
	// stopError 模拟清理失败。
	stopError error
	// calls 记录完整调用顺序；可为空。
	calls *[]string
	// startCount 和 stopCount 用于断言每个实例的调用次数。
	startCount int
	stopCount  int
	// stopContext 保存清理所收到的 context，仅用于测试断言。
	stopContext context.Context
	// stopContextError 保存调用 Stop 时的取消状态。
	stopContextError error
}

func (plugin *lifecycleTestPlugin) Start(context.Context) error {
	plugin.startCount++
	if plugin.calls != nil {
		*plugin.calls = append(*plugin.calls, "start "+plugin.name)
	}
	return plugin.startError
}

func TestStartPluginsReturnsPluginError(t *testing.T) {
	isolatePluginLifecycle(t)

	expected := errors.New("dependency unavailable")
	RegisterPlugin("failing", &lifecycleTestPlugin{startError: expected})
	if err := StartPlugins(t.Context()); !errors.Is(err, expected) {
		t.Fatalf("StartPlugins() error = %v, want %v", err, expected)
	}
}

func (plugin *lifecycleTestPlugin) Stop(ctx context.Context) error {
	plugin.stopCount++
	plugin.stopContext = ctx
	plugin.stopContextError = ctx.Err()
	if plugin.stopOrder != nil {
		*plugin.stopOrder = append(*plugin.stopOrder, plugin.name)
	}
	if plugin.calls != nil {
		*plugin.calls = append(*plugin.calls, "stop "+plugin.name)
	}
	return plugin.stopError
}

// isolatePluginLifecycle 隔离包级注册和启动状态；这些测试不能并行执行。
func isolatePluginLifecycle(t *testing.T) {
	t.Helper()
	mutex.Lock()
	originalPlugins := plugins
	originalSortPlugins := sortPlugins
	originalPluginNames := _plugins
	originalStartedPlugins := startedPlugins
	originalHandler := AfterPluginStartedHandler
	plugins = make(map[string]Plugin)
	sortPlugins = make([]Plugin, 0)
	_plugins = make(map[Plugin]string)
	startedPlugins = nil
	AfterPluginStartedHandler = nil
	mutex.Unlock()
	t.Cleanup(func() {
		mutex.Lock()
		plugins = originalPlugins
		sortPlugins = originalSortPlugins
		_plugins = originalPluginNames
		startedPlugins = originalStartedPlugins
		AfterPluginStartedHandler = originalHandler
		mutex.Unlock()
	})
}

func TestStopPluginsUsesReverseRegistrationOrder(t *testing.T) {
	isolatePluginLifecycle(t)

	stopOrder := make([]string, 0, 3)
	RegisterPlugin("dependency", &lifecycleTestPlugin{name: "dependency", stopOrder: &stopOrder})
	RegisterPlugin("consumer", &lifecycleTestPlugin{name: "consumer", stopOrder: &stopOrder})
	RegisterPlugin("entrypoint", &lifecycleTestPlugin{name: "entrypoint", stopOrder: &stopOrder})

	if err := StartPlugins(t.Context()); err != nil {
		t.Fatal(err)
	}
	StopPlugins(t.Context())

	expected := []string{"entrypoint", "consumer", "dependency"}
	if !reflect.DeepEqual(stopOrder, expected) {
		t.Fatalf("插件停止顺序错误: got=%v want=%v", stopOrder, expected)
	}
}

func TestStartPluginsLifecycle(t *testing.T) {
	startupErr := errors.New("startup failure")
	cleanupErr := errors.New("cleanup failure")
	for _, test := range []struct {
		name      string
		count     int
		failAt    int
		stopFails bool
		want      []string
	}{
		{"all succeed", 3, -1, false, []string{"start A", "start B", "start C"}},
		{"first fails", 3, 0, false, []string{"start A"}},
		{"middle fails", 3, 1, false, []string{"start A", "start B", "stop A"}},
		{"last fails", 3, 2, false, []string{"start A", "start B", "start C", "stop B", "stop A"}},
		{"reverse rollback", 4, 3, false, []string{"start A", "start B", "start C", "start D", "stop C", "stop B", "stop A"}},
		{"cleanup also fails", 3, 2, true, []string{"start A", "start B", "start C", "stop B", "stop A"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			isolatePluginLifecycle(t)
			var calls []string
			instances := make([]*lifecycleTestPlugin, test.count)
			for i := range instances {
				instances[i] = &lifecycleTestPlugin{name: string(rune('A' + i)), calls: &calls}
				if i == test.failAt {
					instances[i].startError = startupErr
				}
				if test.stopFails && i < test.failAt {
					instances[i].stopError = cleanupErr
				}
				RegisterPlugin(instances[i].name, instances[i])
			}
			handlerCalls := 0
			AfterPluginStartedHandler = func() error { handlerCalls++; return nil }
			err := StartPlugins(t.Context())
			if test.failAt == -1 {
				if err != nil || handlerCalls != 1 {
					t.Fatalf("success: error=%v handlerCalls=%d", err, handlerCalls)
				}
			} else if !errors.Is(err, startupErr) || handlerCalls != 0 {
				t.Fatalf("failure: error=%v handlerCalls=%d", err, handlerCalls)
			}
			if test.stopFails {
				if !errors.Is(err, cleanupErr) || !strings.HasPrefix(err.Error(), "启动插件 C: startup failure") {
					t.Fatalf("startup/cleanup errors not preserved: %v", err)
				}
				for _, name := range []string{"A", "B"} {
					if !strings.Contains(err.Error(), "关闭插件 "+name+": cleanup failure") {
						t.Fatalf("missing cleanup diagnostic for %s: %v", name, err)
					}
				}
			}
			if !reflect.DeepEqual(calls, test.want) {
				t.Fatalf("calls=%v want=%v", calls, test.want)
			}
			want := append([]string(nil), test.want...)
			if test.failAt == -1 {
				want = append(want, "stop C", "stop B", "stop A")
			}
			StopPlugins(t.Context())
			StopPlugins(t.Context())
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("shutdown/double stop: calls=%v want=%v", calls, want)
			}
			for i, instance := range instances {
				wantStart, wantStop := 1, 1
				if test.failAt >= 0 && i >= test.failAt {
					wantStop = 0
					if i > test.failAt {
						wantStart = 0
					}
				}
				if instance.startCount != wantStart || instance.stopCount != wantStop {
					t.Fatalf("%s: starts=%d stops=%d, want %d/%d", instance.name, instance.startCount, instance.stopCount, wantStart, wantStop)
				}
			}
		})
	}
}

func TestStartPluginsHandlerFailureRollsBack(t *testing.T) {
	isolatePluginLifecycle(t)
	var calls []string
	RegisterPlugin("A", &lifecycleTestPlugin{name: "A", calls: &calls})
	RegisterPlugin("B", &lifecycleTestPlugin{name: "B", calls: &calls})
	expected := errors.New("after-start failure")
	AfterPluginStartedHandler = func() error { return expected }
	if err := StartPlugins(t.Context()); !errors.Is(err, expected) {
		t.Fatalf("error=%v want=%v", err, expected)
	}
	StopPlugins(t.Context())
	if want := []string{"start A", "start B", "stop B", "stop A"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%v want=%v", calls, want)
	}
}

func TestRollbackContextSurvivesStartupCancellation(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(map[bool]string{false: "canceled", true: "deadline"}[expired], func(t *testing.T) {
			isolatePluginLifecycle(t)
			type contextKey struct{}
			parent := context.WithValue(t.Context(), contextKey{}, "trace")
			ctx, cancel := context.WithCancel(parent)
			cancel()
			if expired {
				ctx, cancel = context.WithDeadline(parent, time.Now().Add(-time.Second))
				defer cancel()
			}
			started := &lifecycleTestPlugin{}
			RegisterPlugin("A", started)
			RegisterPlugin("B", &lifecycleTestPlugin{startError: ctx.Err()})
			if err := StartPlugins(ctx); !errors.Is(err, ctx.Err()) {
				t.Fatal(err)
			}
			if started.stopCount != 1 || started.stopContextError != nil {
				t.Fatalf("stops=%d context error=%v", started.stopCount, started.stopContextError)
			}
			if started.stopContext.Value(contextKey{}) != "trace" {
				t.Fatal("lost context value")
			}
			deadline, ok := started.stopContext.Deadline()
			if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > startupRollbackTimeout {
				t.Fatalf("unexpected cleanup deadline: %v", deadline)
			}
		})
	}
}

func TestStopPluginsSkipsUnstartedPlugins(t *testing.T) {
	isolatePluginLifecycle(t)
	instance := &lifecycleTestPlugin{}
	RegisterPlugin("unstarted", instance)
	StopPlugins(t.Context())
	if instance.stopCount != 0 {
		t.Fatalf("unstarted plugin stopped %d times", instance.stopCount)
	}
}

func TestStopPluginsContinuesAfterCleanupError(t *testing.T) {
	isolatePluginLifecycle(t)
	var calls []string
	RegisterPlugin("A", &lifecycleTestPlugin{name: "A", calls: &calls})
	RegisterPlugin("B", &lifecycleTestPlugin{name: "B", calls: &calls, stopError: errors.New("cleanup failure")})
	if err := StartPlugins(t.Context()); err != nil {
		t.Fatal(err)
	}
	StopPlugins(t.Context())
	StopPlugins(t.Context())
	if want := []string{"start A", "start B", "stop B", "stop A"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%v want=%v", calls, want)
	}
}
