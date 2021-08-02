package configuration

import (
	"flag"
	"strings"

	"git.xiagaogao.com/coffee/base/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	Model_dev     = "dev"
	Model_test    = "test"
	Model_product = "prod"
)

var configFile = pflag.StringP("config", "c", "./config.yml", "配置文件路径")

var runModel = ""

func GetRunModel() string {
	return runModel
}

func loadConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	if !pflag.Parsed() {
		pflag.Parse()
	}
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvPrefix("ENV")
	viper.AutomaticEnv()
	viper.SetConfigFile(*configFile)
	if err := viper.MergeInConfig(); err != nil {
		log.Warn("加载日志文件失败", zap.Error(err))
	}
	if viper.GetString(_run_model) == "" {
		log.Panic("没有指定run model")
	}
	runModel = viper.GetString(_run_model)
	log.Info("加载配置", zap.String("run model", viper.GetString(_run_model)))
}
