package engine

import (
	"context"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
)

func TestWaitServiceStopContextCancellation(t *testing.T) {
	for range 20 {
		ctx, cancel := context.WithCancel(t.Context())
		var calls atomic.Int64
		done := make(chan struct{})
		go func() {
			WaitServiceStop(ctx, func() { calls.Add(1) })
			close(done)
		}()
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("调用方取消后信号等待未结束")
		}
		if calls.Load() != 1 {
			t.Fatalf("关闭回调调用次数=%d，期望一次", calls.Load())
		}
	}
}

func TestBuildRootCommand_ExecuteExtraCommand(t *testing.T) {
	serviceInfo := configuration.ServiceInfo{
		ServiceName: "boot-test",
		Descriptor:  "boot test service",
	}
	executed := false
	extraCommand := &cobra.Command{
		Use: "cli",
		Run: func(cmd *cobra.Command, args []string) {
			executed = true
		},
	}

	rootCmd, err := buildRootCommand(context.Background(), serviceInfo, testServiceStart, WithExtraCommands(extraCommand))
	if err != nil {
		t.Fatalf("buildRootCommand returned error: %v", err)
	}

	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs([]string{"cli"})
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext returned error: %v", err)
	}
	if !executed {
		t.Fatalf("extra command was not executed")
	}
}

func TestBuildRootCommand_ConflictWithBuiltinCommand(t *testing.T) {
	serviceInfo := configuration.ServiceInfo{
		ServiceName: "boot-test",
		Descriptor:  "boot test service",
	}
	_, err := buildRootCommand(context.Background(), serviceInfo, testServiceStart, WithExtraCommands(&cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {},
	}))
	if err == nil {
		t.Fatalf("expected conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "内建命令重名") {
		t.Fatalf("expected conflict error message, got: %v", err)
	}
}

func TestBuildRootCommand_DefaultCommandsUnchanged(t *testing.T) {
	serviceInfo := configuration.ServiceInfo{
		ServiceName: "boot-test",
		Descriptor:  "boot test service",
	}
	rootCmd, err := buildRootCommand(context.Background(), serviceInfo, testServiceStart)
	if err != nil {
		t.Fatalf("buildRootCommand returned error: %v", err)
	}

	commandNames := make(map[string]struct{})
	for _, command := range rootCmd.Commands() {
		commandNames[command.Name()] = struct{}{}
	}

	builtinCommands := []string{"version", "restart", "start", "daemonStart", "stop", "setup"}
	for _, commandName := range builtinCommands {
		if _, exists := commandNames[commandName]; !exists {
			t.Fatalf("builtin command %q not found", commandName)
		}
	}
}

func testServiceStart(ctx context.Context, cmd *cobra.Command, args []string) (ServiceCloseCallback, error) {
	return nil, nil
}
