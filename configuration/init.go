package configuration

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	Model_dev     = "dev"
	Model_test    = "test"
	Model_product = "product"
)

var configFile = pflag.StringP("config", "c", "", "配置文件路径")

func init() {
	if !pflag.Parsed() {
		pflag.Parse()
	}
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}

	// 本地配置里面如果有配置远程配置中心的也需要处理

}
