package configuration

import (
	"flag"
	"git.xiagaogao.com/coffee/base/log"
	"github.com/json-iterator/go/extra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	Model_dev     = "dev"
	Model_test    = "test"
	Model_product = "product"
)

var configFile = pflag.StringP("config", "c", "", "配置文件路径")

func init() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	if !pflag.Parsed() {
		pflag.Parse()
	}
	viper.BindPFlags(pflag.CommandLine)
	viper.AddConfigPath(".")
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
	} else {
		viper.SetConfigName("config")
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Warn("加载日志文件失败", zap.Error(err))
	}
	viper.SetEnvPrefix("ENV")
	viper.AutomaticEnv()
	// 本地配置里面如果有配置远程配置中心的也需要处理
	log.WatchLevel()
}
func initJsonConifg() {
	extra.RegisterFuzzyDecoders()
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
}

func initLoggerConfig() {
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
}
