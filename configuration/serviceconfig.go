package configuration

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/boot/base/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var defaultServiceConfig *ServiceConfig
var currentServiceInfo ServiceInfo
var mutex = new(sync.RWMutex)
var rootCtx context.Context

func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo) {
	initServiceConfig(ctx)
	initServiceInfo(ctx, serviceInfo)
	loadRemortConfigProvider(ctx)
}

func initServiceInfo(ctx context.Context, serviceInfo ServiceInfo) {
	if rootCtx == nil {
		rootCtx = ctx
	}
	viper.SetDefault("version", "0.0.0")
	viper.SetDefault("scheme", MicorServiceProtocolScheme)
	err := viper.Unmarshal(&serviceInfo)
	if err != nil {
		log.Fatal("加载ServiceInfo失败", zap.Error(err))
	}
	if serviceInfo.ServiceName == "" {
		log.Fatal("服务名没有设置")
	}
	currentServiceInfo = serviceInfo
	viper.MergeInConfig()
	log.Debug("加载服务信息", zap.String("version", serviceInfo.Version), zap.String("scheme", serviceInfo.Scheme), zap.String("APIDefine", serviceInfo.APIDefine), zap.String("Descriptor", serviceInfo.Descriptor))
}

func initServiceConfig(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if defaultServiceConfig != nil {
		return
	}
	if rootCtx == nil {
		rootCtx = ctx
	}
	log.InitLogger(true)
	conf := &ServiceConfig{}
	err := viper.Unmarshal(conf)
	if err != nil {
		log.Fatal("不能从配置中读取服务配置", zap.Error(err))
	}
	if conf.Model == "" {
		log.Fatal("service model没有设置")
	}
	defaultServiceConfig = conf
	viper.MergeInConfig()
	log.Debug("加载基础配置", zap.String("model", conf.Model), zap.String("RemoteConfigProvide", conf.RemoteConfigProvide))
}
func loadRemortConfigProvider(ctx context.Context) error {
	if defaultServiceConfig.RemoteConfigProvide != "" {
		// TODO
	}
	viper.MergeInConfig()
	log.Debug("加载远程配置")
	return nil
}

func GetModel() string {
	return defaultServiceConfig.Model
}

func GetServiceName() string {
	return currentServiceInfo.ServiceName
}

func GetServiceInfo() ServiceInfo {
	return currentServiceInfo
}

func SetModel(model string) {
	viper.SetDefault("model", model)
}
