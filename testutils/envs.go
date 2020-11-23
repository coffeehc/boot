package testutils

import (
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
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
}
