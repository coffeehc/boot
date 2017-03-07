package base

import (
	"github.com/coffeehc/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const errScopeLoadConfig = "loadConfig"

//LoadConfig load the config from config path
func LoadConfig(configPath string, config interface{}) Error {
	logger.Debug("load config file %s", configPath)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return NewError(ErrCodeBaseSystemConfig, errScopeLoadConfig, err.Error())
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return NewError(ErrCodeBaseSystemConfig, errScopeLoadConfig, err.Error())
	}
	return nil
}
