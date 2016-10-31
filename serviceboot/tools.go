package serviceboot

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

func LoadConfigPath(serviceConfig interface{}) string {
	*configPath = base.GetDefaultConfigPath(*configPath)
	err := base.LoadConfig(*configPath, serviceConfig)
	if err != nil {
		logger.Warn("加载服务器配置[%s]失败,%s", *configPath, err)
	}
	logger.Debug("serviceboot Config is %#v", serviceConfig)
	return *configPath
}
