package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// engineLifecyclePlugin 用于子进程验证真实命令入口，不依赖外部服务。
type engineLifecyclePlugin struct {
	// name 标记输出中的插件身份。
	name string
	// startError 控制启动失败，nil 表示成功。
	startError error
	// panicOnStart 验证 engine 既有的 panic 转错误行为。
	panicOnStart bool
}

func (p *engineLifecyclePlugin) Start(context.Context) error {
	fmt.Fprintf(os.Stdout, "lifecycle:start:%s\n", p.name)
	if p.panicOnStart {
		panic("plugin startup panic")
	}
	return p.startError
}

func (p *engineLifecyclePlugin) Stop(context.Context) error {
	fmt.Fprintf(os.Stdout, "lifecycle:stop:%s\n", p.name)
	return nil
}

// TestEngineLifecycleProcess 隔离全局插件注册和 os.Exit，覆盖真实进程退出与取消关闭。
func TestEngineLifecycleProcess(t *testing.T) {
	for _, test := range []struct {
		mode        string
		wantFailure bool
		want        []string
	}{
		{"startup-failure", true, []string{"start:A", "start:B", "start:C", "stop:B", "stop:A"}},
		{"command-error", false, []string{"start:A", "start:B", "start:C", "stop:B", "stop:A"}},
		{"normal-cancel", false, []string{"start:A", "start:B", "start:C", "close", "stop:C", "stop:B", "stop:A"}},
		{"plugin-panic", true, []string{"start:A", "start:B", "start:C"}},
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	expected := errors.New("plugin startup sentinel")
	serviceInfo := configuration.ServiceInfo{ServiceName: "lifecycle-test"}
	start := func(context.Context, *cobra.Command, []string) (ServiceCloseCallback, error) {
		for _, name := range []string{"A", "B", "C"} {
			instance := &engineLifecyclePlugin{name: name}
			if name == "C" {
				switch mode {
				case "startup-failure", "command-error":
					instance.startError = expected
				case "plugin-panic":
					instance.panicOnStart = true
				}
			}
			plugin.RegisterPlugin(name, instance)
		}
		if mode == "normal-cancel" {
			plugin.AfterPluginStartedHandler = func() error { cancel(); return nil }
		}
		return func() { fmt.Fprint(os.Stdout, "lifecycle:close\n") }, nil
	}
	if mode == "command-error" {
		command, err := buildRootCommand(ctx, serviceInfo, start)
		if err != nil {
			t.Fatal(err)
		}
		command.SetArgs([]string{"start"})
		if err := command.ExecuteContext(ctx); !errors.Is(err, expected) {
			t.Fatalf("command lost startup error: %v", err)
		}
	} else {
		StartEngine(ctx, serviceInfo, start)
		if mode != "normal-cancel" {
			fmt.Fprint(os.Stdout, "lifecycle:unexpected-success\n")
			t.Fatal("startup failure returned as success")
		}
	}
	plugin.StopPlugins(context.Background())
	plugin.StopPlugins(context.Background())
}
