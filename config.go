package boot

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// SimpleServiceInfo 简单的 ServiceInfo 配置
type ServiceInfo struct {
	ServiceName string `yaml:"service_name" json:"service_name"`
	Version     string `yaml:"version" json:"version"`
	Descriptor  string `yaml:"descriptor" json:"descriptor"`
	APIDefine   string `yaml:"api_define" json:"api_define"`
	Scheme      string `yaml:"scheme" json:"scheme"`
}

func CheckServiceInfoConfig(ctx context.Context, serviceInfo *ServiceInfo) error {
	if serviceInfo.ServiceName == "" {
		return errors.New("没有配置 ServiceName")
	}
	if serviceInfo.Version == "" {
		return errors.New("没有配置 ServiceVersion")
	}
	if serviceInfo.Scheme == "" {
		return errors.New("没有配置 ServiceScheme")
	}
	return nil
}

func PrintServiceInfo(serviceInfo *ServiceInfo, logger *zap.Logger) {
	logger.Info(fmt.Sprintf("ServiceName:%s", serviceInfo.ServiceName))
	logger.Info(fmt.Sprintf("Version:%s", serviceInfo.Version))
	logger.Info(fmt.Sprintf("Descriptor:%s", serviceInfo.Descriptor))
	logger.Info(fmt.Sprintf("APIDefine:%s", serviceInfo.APIDefine))
	logger.Info(fmt.Sprintf("Scheme:%s", serviceInfo.Scheme))
}
