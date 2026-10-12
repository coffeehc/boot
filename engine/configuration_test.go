package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestStartReturnsConfigurationSourceErrorBeforeServiceInitialization(t *testing.T) {
	flag := pflag.Lookup("config")
	previous := flag.Value.String()
	t.Cleanup(func() { _ = flag.Value.Set(previous) })
	configuration.SetRunModel(configuration.Model_test)
	for _, body := range []string{"", "resource: [ do-not-leak-secret-value\n"} {
		path := filepath.Join(t.TempDir(), "explicit.yml")
		if body != "" {
			if err := os.WriteFile(path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
		if err := flag.Value.Set(path); err != nil {
			t.Fatal(err)
		}
		started := false
		command := buildStartCmd(t.Context(), configuration.ServiceInfo{ServiceName: "source-error-start"}, func(context.Context, *cobra.Command, []string) (ServiceCloseCallback, error) {
			started = true
			return nil, nil
		}, configuration.WithFileOnly())
		err := command.RunE(command, nil)
		if err == nil || strings.Contains(err.Error(), "do-not-leak") || started {
			t.Fatal("source initialization failure did not reach engine safely", err, started)
		}
		if GetRootContext() != nil {
			t.Fatal("failed configuration startup retained runtime root")
		}
	}
}
