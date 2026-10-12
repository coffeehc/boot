package configuration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestConfigurationSourceErrorsAndDebounceOption(t *testing.T) {
	path, _, o := configTest(t, "resource: {limit: 1}\n")
	if o.delay != 15*time.Millisecond {
		t.Fatal("test option ignored")
	}
	for _, test := range []struct {
		name, body string
		missing    bool
	}{{"missing", "", true}, {"parse", "resource: [ do-not-leak-secret-value\n", false}, {"logger", "logger: {level: do-not-leak-secret-value}\n", false}} {
		t.Run(test.name, func(t *testing.T) {
			if test.missing {
				*configFile = filepath.Join(filepath.Dir(path), "explicit-missing.yml")
			} else {
				*configFile = path
				rewrite(t, path, test.body)
			}
			err := InitConfiguration(context.Background(), ServiceInfo{ServiceName: "invalid-source"}, WithFileOnly())
			if err == nil || strings.Contains(err.Error(), "do-not-leak") {
				t.Fatal("source error ignored or leaked", err)
			}
			if configurationState.owner != nil {
				t.Fatal("failed initialization left usable owner")
			}
		})
	}
	*configFile = path
	rewrite(t, path, "resource: {limit: 1}\n")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := InitConfiguration(ctx, ServiceInfo{ServiceName: "default-delay"}, WithFileOnly()); err != nil {
		t.Fatal(err)
	}
	if configurationState.owner.delay != 5*time.Second {
		t.Fatal("default debounce changed")
	}
	if err := InitConfiguration(ctx, ServiceInfo{ServiceName: "configured-delay"}, WithFileOnly(), WithReloadDebounce(10*time.Second)); err != nil {
		t.Fatal(err)
	}
	if configurationState.owner.delay != 10*time.Second {
		t.Fatal("configured debounce ignored")
	}
	if err := InitConfiguration(ctx, ServiceInfo{ServiceName: "invalid-delay"}, WithReloadDebounce(0)); err == nil {
		t.Fatal("nonpositive debounce accepted")
	}
	// Sole parser remains the Viper used by ordinary Boot initialization.
	if viper.GetViper() == nil {
		t.Fatal("Viper missing")
	}
}

func TestBindConfigWatcherFailureDoesNotFillOrSubscribe(t *testing.T) {
	path, _, o := configTest(t, "resource: {limit: 2}\n")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	c := boundConfig{Limit: 81}
	if err := BindConfig("resource", &c, func(any, error) {}); err == nil {
		t.Fatal("missing watcher directory accepted")
	}
	if c.Limit != 81 || len(o.bindings) != 0 || o.watcher != nil {
		t.Fatal("watcher setup failure polluted first binding")
	}
}
