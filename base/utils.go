package base

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/coffeehc/commons"
	"github.com/coffeehc/logger"
	"gopkg.in/yaml.v2"
)

const ERR_SCOPE_LOADCONFIG = "loadConfig"

func GetDefaultConfigPath(configPath string) string {
	if configPath == "" {
		configPath = path.Join(commons.GetAppDir(), "config.yml")
		_, err := os.Open(configPath)
		if err != nil {
			//logger.Error("%s 不存在", configPath)
			dir, err := os.Getwd()
			if err != nil {
				logger.Error("获取不到工作目录")
				return ""
			}
			configPath = path.Join(dir, "config.yml")
			_, err = os.Open(configPath)
			if err != nil {
				//logger.Error("%s 不存在", configPath)
				return ""
			}
		}
	}
	return configPath
}

func LoadConfig(configPath string, config interface{}) Error {
	logger.Debug("load config file %s", configPath)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return NewError(ERRCODE_BASE_SYSTEM_CONFIG_ERROR, ERR_SCOPE_LOADCONFIG, err.Error())
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return NewError(ERRCODE_BASE_SYSTEM_CONFIG_ERROR, ERR_SCOPE_LOADCONFIG, err.Error())
	}
	return nil
}
