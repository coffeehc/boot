package testutils

import (
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/spf13/viper"
)

func InitTestConfig() {
	configuration.SetRunModel("dev")
	viper.Set("logger", &log.Config{
		Level: "debug",
		FileConfig: log.FileLogConfig{
			Enable: false,
		},
		EnableConsole: true,
		EnableColor:   true,
		EnableSampler: true,
	})
	log.InitLogger(true)
}
