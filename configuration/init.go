package configuration

import (
	"flag"
	"os"

	"git.xiagaogao.com/coffee/boot/base/log"
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
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
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
