package engine

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
)

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
