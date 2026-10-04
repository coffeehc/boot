package configuration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestFileOnlyConfigurationExcludesEnvironmentAndPreviousInitialization(t *testing.T) {
	previousPath, previousDefault := *configFile, defaultRunModel
	t.Cleanup(func() {
		*configFile, defaultRunModel = previousPath, previousDefault
		viper.Reset()
	})
	root := t.TempDir()
	first := filepath.Join(root, "first.yml")
	second := filepath.Join(root, "second.yml")
	if err := os.WriteFile(first, []byte("run_model: test\nresource:\n  limit: 8\nold_file: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("resource:\n  limit: 12\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ENV_RUN_MODEL", "prod")
	t.Setenv("ENV_RESOURCE_LIMIT", "99")
	t.Setenv("BOOT_BOUND_RESOURCE", "88")
	SetRunModel(Model_dev)
	*configFile = first
	InitConfiguration(context.Background(), ServiceInfo{ServiceName: "file-only-test"})
	if GetRunModel() != Model_product || viper.GetInt("resource.limit") != 99 {
		t.Fatal("default configuration no longer applies ENV_ overrides")
	}
	if err := viper.BindEnv("bound_resource", "BOOT_BOUND_RESOURCE"); err != nil {
		t.Fatal(err)
	}
	viper.Set("previous_override", true)
	InitConfiguration(context.Background(), ServiceInfo{ServiceName: "file-only-test"}, WithFileOnly())
	if GetRunModel() != Model_test || viper.GetInt("resource.limit") != 8 ||
		viper.IsSet("bound_resource") || viper.IsSet("previous_override") {
		t.Fatalf("file-only configuration inherited environment or program values: %v", viper.AllSettings())
	}
	*configFile = second
	InitConfiguration(context.Background(), ServiceInfo{ServiceName: "file-only-test"}, WithFileOnly())
	if GetRunModel() != Model_dev || viper.GetInt("resource.limit") != 12 || viper.IsSet("old_file") {
		t.Fatalf("next file-only initialization inherited previous file or lost Boot defaults: %v", viper.AllSettings())
	}
	InitConfiguration(context.Background(), ServiceInfo{ServiceName: "file-only-test"})
	if GetRunModel() != Model_product || viper.GetInt("resource.limit") != 99 {
		t.Fatal("file-only option leaked into a later default initialization")
	}
}

func TestFileOnlyConfigurationDoesNotReadUnregisteredEnvironmentKeys(t *testing.T) {
	previousPath, previousDefault := *configFile, defaultRunModel
	t.Cleanup(func() {
		*configFile, defaultRunModel = previousPath, previousDefault
		viper.Reset()
	})
	*configFile = filepath.Join(t.TempDir(), "absent.yml")
	SetRunModel(Model_test)
	t.Setenv("ENV_UNDECLARED_RESOURCE", "value")
	InitConfiguration(context.Background(), ServiceInfo{ServiceName: "file-only-test"}, WithFileOnly())
	if GetRunModel() != Model_test || viper.GetString("undeclared_resource") != "" {
		t.Fatal("file-only configuration queried environment for an absent file key")
	}
}
