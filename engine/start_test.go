package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin"
	"github.com/spf13/cobra"
)

func TestStartCommandReturnsRecoveredPanic(t *testing.T) {
	configuration.SetRunModel(configuration.Model_test)
	command := buildStartCmd(context.Background(), configuration.ServiceInfo{ServiceName: "panic-test"},
		func(context.Context, *cobra.Command, []string) (ServiceCloseCallback, error) {
			panic("boom")
		})

	if err := command.RunE(command, nil); err == nil {
		t.Fatal("RunE() error = nil, want recovered panic")
	}
}

func TestStartCommandSkipsServiceStartWhenParentAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := false
	command := buildStartCmd(ctx, configuration.ServiceInfo{ServiceName: "canceled-test"},
		func(context.Context, *cobra.Command, []string) (ServiceCloseCallback, error) {
			started = true
			return nil, nil
		})
	err := command.RunE(command, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunE error=%v, want context canceled", err)
	}
	if started {
		t.Fatal("ServiceStart ran after parent context was already canceled")
	}
	if GetRootContext() != nil {
		t.Fatal("runtime root remained registered after canceled startup")
	}
}

// engineLifecyclePlugin 用于子进程验证真实命令入口，不依赖外部服务。
type engineLifecyclePlugin struct {
	// name 标记输出中的插件身份。
	name string
	// startError 控制启动失败，nil 表示成功。
	startError error
	// panicOnStart 验证 engine 既有的 panic 转错误行为。
	panicOnStart bool
	// cancelOnStart 模拟启动阶段的外部取消；可为空。
	cancelOnStart context.CancelFunc
	// panicOnStop 模拟停止异常，验证其余插件仍被清理。
	panicOnStop bool
}

type engineLifecycleContextKey struct{}

func (p *engineLifecyclePlugin) Start(ctx context.Context) error {
	if GetRootContext() != ctx {
		return errors.New("plugin Start did not receive the registered root context")
	}
	fmt.Fprintf(os.Stdout, "lifecycle:start:%s\n", p.name)
	if p.panicOnStart {
		panic("plugin startup panic")
	}
	if p.cancelOnStart != nil {
		p.cancelOnStart()
	}
	return p.startError
}

func (p *engineLifecyclePlugin) Stop(ctx context.Context) error {
	root := GetRootContext()
	if ctx == root || ctx.Err() != nil {
		return errors.New("plugin Stop did not receive an independent active shutdown context")
	}
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > shutdownTimeout {
		return errors.New("plugin Stop context has no deadline")
	}
	if root == nil || !errors.Is(root.Err(), context.Canceled) {
		return errors.New("runtime root was not canceled before plugin Stop")
	}
	if ctx.Value(engineLifecycleContextKey{}) != "trace" {
		return errors.New("shutdown context did not retain runtime context values")
	}
	fmt.Fprintf(os.Stdout, "lifecycle:stop:%s\n", p.name)
	if p.panicOnStop {
		panic("plugin shutdown panic")
	}
	return nil
}

// TestEngineLifecycleProcess 隔离全局插件注册和 os.Exit，覆盖真实进程退出与取消关闭。
func TestEngineLifecycleProcess(t *testing.T) {
	for _, test := range []struct {
		mode        string
		wantFailure bool
		want        []string
	}{
		{"startup-failure", true, []string{"start:A", "start:B", "start:C", "stop:B", "stop:A", "close"}},
		{"command-error", false, []string{"start:A", "start:B", "start:C", "stop:B", "stop:A", "close"}},
		{"command-cleanup-error", false, []string{"start:A", "start:B", "start:C", "stop:B", "stop:A", "close"}},
		{"service-error", true, []string{"close"}},
		{"startup-cancel", true, []string{"start:A", "start:B", "stop:B", "stop:A", "close"}},
		{"normal-cancel", false, []string{"start:A", "start:B", "start:C", "close", "stop:C", "stop:B", "stop:A"}},
		{"signal-term", false, []string{"start:A", "start:B", "start:C", "close", "stop:C", "stop:B", "stop:A"}},
		{"signal-int", false, []string{"start:A", "start:B", "start:C", "close", "stop:C", "stop:B", "stop:A"}},
		{"plugin-panic", true, []string{"start:A", "start:B", "start:C", "stop:B", "stop:A", "close"}},
		{"handler-panic", true, []string{"start:A", "start:B", "start:C", "stop:C", "stop:B", "stop:A", "close"}},
		{"close-panic", true, []string{"start:A", "start:B", "start:C", "close", "stop:C", "stop:B", "stop:A"}},
		{"stop-panic", true, []string{"start:A", "start:B", "start:C", "close", "stop:C", "stop:B", "stop:A"}},
	} {
		t.Run(test.mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestEngineLifecycleHelper$")
			command.Env = append(os.Environ(), "BOOT_LIFECYCLE_TEST="+test.mode)
			command.Dir = t.TempDir()
			output, err := command.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("child did not exit: %v\n%s", ctx.Err(), output)
			}
			if test.wantFailure {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() <= 0 {
					t.Fatalf("want non-zero exit, got %v\n%s", err, output)
				}
			} else if err != nil {
				t.Fatalf("want clean exit, got %v\n%s", err, output)
			}
			var calls []string
			for _, line := range strings.Split(string(output), "\n") {
				if strings.HasPrefix(line, "lifecycle:") {
					calls = append(calls, strings.TrimPrefix(line, "lifecycle:"))
				}
			}
			if strings.Join(calls, ",") != strings.Join(test.want, ",") {
				t.Fatalf("calls=%v want=%v\n%s", calls, test.want, output)
			}
		})
	}
}

// TestEngineLifecycleHelper 仅在父测试指定的独立进程中运行 engine。
func TestEngineLifecycleHelper(t *testing.T) {
	mode := os.Getenv("BOOT_LIFECYCLE_TEST")
	if mode == "" {
		return
	}
	t.Setenv(Env_DaemonMode, "false")
	configuration.SetRunModel(configuration.Model_test)
	configuration.Version = "lifecycle-test"
	os.Args = []string{os.Args[0], "start"}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), engineLifecycleContextKey{}, "trace"))
	defer cancel()
	expected := errors.New("plugin startup sentinel")
	closeExpected := errors.New("business cleanup sentinel")
	serviceInfo := configuration.ServiceInfo{ServiceName: "lifecycle-test"}
	start := func(runCtx context.Context, _ *cobra.Command, _ []string) (ServiceCloseCallback, error) {
		for _, name := range []string{"A", "B", "C"} {
			instance := &engineLifecyclePlugin{name: name}
			if name == "C" {
				switch mode {
				case "startup-failure", "command-error", "command-cleanup-error":
					instance.startError = expected
				case "plugin-panic":
					instance.panicOnStart = true
				}
			}
			if name == "B" {
				if mode == "startup-cancel" {
					instance.cancelOnStart = cancel
				}
				instance.panicOnStop = mode == "stop-panic"
			}
			plugin.RegisterPlugin(name, instance)
		}
		if mode == "normal-cancel" || mode == "close-panic" || mode == "stop-panic" {
			plugin.AfterPluginStartedHandler = func() error { cancel(); return nil }
		}
		if mode == "signal-term" {
			plugin.AfterPluginStartedHandler = func() error { return syscall.Kill(os.Getpid(), syscall.SIGTERM) }
		}
		if mode == "signal-int" {
			plugin.AfterPluginStartedHandler = func() error { return syscall.Kill(os.Getpid(), syscall.SIGINT) }
		}
		if mode == "handler-panic" {
			plugin.AfterPluginStartedHandler = func() error { panic(expected) }
		}
		if GetRootContext() != runCtx {
			return nil, errors.New("ServiceStart did not receive the registered root context")
		}
		closeCallback := func() {
			if root := GetRootContext(); root == nil || !errors.Is(root.Err(), context.Canceled) {
				fmt.Fprint(os.Stdout, "lifecycle:root-not-canceled-at-close\n")
			}
			fmt.Fprint(os.Stdout, "lifecycle:close\n")
			if mode == "close-panic" || mode == "command-cleanup-error" {
				panic(closeExpected)
			}
		}
		if mode == "service-error" {
			return closeCallback, expected
		}
		return closeCallback, nil
	}
	if mode == "command-error" || mode == "command-cleanup-error" {
		command, err := buildRootCommand(ctx, serviceInfo, start)
		if err != nil {
			t.Fatal(err)
		}
		command.SetArgs([]string{"start"})
		err = command.ExecuteContext(ctx)
		if !errors.Is(err, expected) {
			t.Fatalf("command lost startup error: %v", err)
		}
		if mode == "command-cleanup-error" && !errors.Is(err, closeExpected) {
			t.Fatalf("command lost cleanup panic cause: %v", err)
		}
	} else {
		StartEngine(ctx, serviceInfo, start)
		if mode != "normal-cancel" && mode != "signal-term" && mode != "signal-int" {
			fmt.Fprint(os.Stdout, "lifecycle:unexpected-success\n")
			t.Fatal("startup failure returned as success")
		}
	}
	if GetRootContext() != nil {
		t.Fatal("runtime root remained registered after start command returned")
	}
	plugin.StopPlugins(context.Background())
	plugin.StopPlugins(context.Background())
}
