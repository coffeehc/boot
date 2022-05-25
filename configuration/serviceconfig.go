package configuration

import (
	"context"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/base/utils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var onConfigChanges = make([]func(), 0)
var currentServiceInfo ServiceInfo
var rootCtx context.Context
var EnableRemoteLog = false

func RegisterOnConfigChange(onConfigChange func()) {
	onConfigChanges = append(onConfigChanges, onConfigChange)
}

func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo) {
	if serviceInfo.Metadata == nil {
		serviceInfo.Metadata = map[string]string{
			"git_rev": GitRev,
			// "build_version": BuildVersion,
			"build_time": BuildTime,
			"git_tag":    GitTag,
			"version":    Version,
		}
	}
	viper.SetConfigType("yaml")
	// 默认开启远程配置
	// viper.SetDefault(enableRemoteConfigKey, false)
	loadConfig()
	initServiceInfo(ctx, serviceInfo)
	// loadRemoteConfig(ctx, serviceInfo)
	log.InitLogger(true)
}

func initServiceInfo(ctx context.Context, serviceInfo ServiceInfo) {
	if rootCtx == nil {
		rootCtx = ctx
	}
	if serviceInfo.ServiceName == "" {
		log.Panic("服务名没有设置")
	}
	currentServiceInfo = serviceInfo
	if EnableRemoteLog {
		localIp, err1 := utils.GetLocalIP()
		if err1 != nil {
			log.Panic("获取本机IP失败", err1.GetFieldsWithCause()...)
		}
		log.ResetLogger(zap.String("serviceName", serviceInfo.ServiceName), zap.String("localIp", localIp.String()))
	}
	log.Info("加载服务信息", zap.Any("serviceInfo", serviceInfo))
}

func GetServiceName() string {
	return currentServiceInfo.ServiceName
}

func GetServiceInfo() ServiceInfo {
	return currentServiceInfo
}

func SetRunModel(runModel string) {
	viper.SetDefault(_run_model, runModel)
}
