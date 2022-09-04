package configuration

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	_run_model = "run_model"
)

func UnmarshalKeyFormJson(key string, body interface{}) error {
	return viper.UnmarshalKey(key, body, func(config *mapstructure.DecoderConfig) {
		config.TagName = "json"
	})
}
