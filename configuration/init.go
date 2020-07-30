package configuration

import (
	"flag"

	"git.xiagaogao.com/coffee/base/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	Model_dev     = "dev"
	Model_test    = "test"
	Model_product = "product"
)

var configFile = pflag.StringP("config", "c", "./cofnig.yml", "配置文件路径")

func registerAlias() {
	viper.RegisterAlias(_run_model, "RUN_MODEL")
}

func loadConfig() {
	registerAlias()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	if !pflag.Parsed() {
		pflag.Parse()
	}
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvPrefix("ENV")
	viper.AutomaticEnv()
	if *configFile == "" {
		log.Warn("没有指定config文件路径")
		return
	}
	viper.SetConfigFile(*configFile)
	if err := viper.MergeInConfig(); err != nil {
		log.Warn("加载日志文件失败", zap.Error(err))
	}
	if viper.GetString(_run_model) == "" {
		log.Fatal("没有指定run model")
	}
	log.Info("加载配置", zap.String("run model", viper.GetString(_run_model)))
}

func initDefaultLoggerConfig() {
	viper.SetDefault("logger", &log.Config{
		Level: "info",
		FileConfig: &log.FileLogConfig{
			FileName:   "./logs/service.log",
			Disable:    false,
			Maxsize:    100,
			MaxBackups: 10,
			MaxAge:     7,
			Compress:   false,
		},
		EnableConsole: false,
		EnableColor:   false,
		EnableSampler: false,
	})
	log.InitLogger(true)
	log.WatchLevel()
}
