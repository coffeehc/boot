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

// InitConfiguration 串行初始化服务配置和身份；默认允许 ENV_ 环境覆盖。
// WithFileOnly 将本轮限制为文件和 Boot 默认值，不继承之前初始化的配置来源。
func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo, opts ...Option) {
	if serviceInfo.Metadata == nil {
		serviceInfo.Metadata = map[string]string{
			"git_rev": GitRev,
			// "build_version": BuildVersion,
			"build_time": BuildTime,
			"git_tag":    GitTag,
			"version":    Version,
		}
	}
	options := configurationOptions{}
	for _, option := range opts {
		if option != nil {
			option(&options)
		}
	}
	loadConfig(options)
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

// SetRunModel 注册 Boot 运行模式默认值；每次文件配置初始化均恢复此默认值。
func SetRunModel(model string) {
	defaultRunModel = model
	viper.SetDefault(_run_model, model)
}
