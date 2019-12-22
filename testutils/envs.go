package testutils

import (
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/spf13/viper"
)

func InitTestConfig() {
	configuration.SetModel("dev")
	viper.Set("logger", &log.Config{
		Level: "debug",
		FileConfig: &log.FileLogConfig{
			Disable: true,
		},
		EnableConsole: true,
		EnableColor:   true,
		EnableSampler: true,
	})
}
