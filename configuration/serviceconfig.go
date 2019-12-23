package configuration

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/base/utils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var defaultServiceConfig *ServiceConfig
var currentServiceInfo ServiceInfo
var mutex = new(sync.RWMutex)
var rootCtx context.Context

func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo) {
	initLoggerConfig()
	initServiceConfig(ctx)
	initServiceInfo(ctx, serviceInfo)
	viper.MergeInConfig()
	loadRemortConfigProvider(ctx)
}

func initServiceInfo(ctx context.Context, serviceInfo ServiceInfo) {
	if rootCtx == nil {
		rootCtx = ctx
	}
	viper.SetDefault("version", "0.0.0")
	viper.SetDefault("scheme", MicroServiceProtocolScheme)
	err := viper.Unmarshal(&serviceInfo)
	if err != nil {
		log.Fatal("加载ServiceInfo失败", zap.Error(err))
	}
	if serviceInfo.ServiceName == "" {
		log.Fatal("服务名没有设置")
	}
	currentServiceInfo = serviceInfo
	localIp, err1 := utils.GetLocalIP()
	if err1 != nil {
		log.Fatal("获取本机IP失败", err1.GetFieldsWithCause()...)
	}
	log.SetBaseFields(zap.String("serviceName", serviceInfo.ServiceName), zap.String("localIp", localIp.String()))
	viper.MergeInConfig()
	log.Debug("加载服务信息", zap.Any("serviceInfo", serviceInfo))
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
	viper.RegisterAlias("model", "RUN_MODEL")
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
	log.Debug("加载基础配置", zap.String("model", conf.Model), zap.Any("RemoteConfigProvide", conf.RemoteConfigProvide))
}

func loadRemortConfigProvider(ctx context.Context) error {
	if defaultServiceConfig.RemoteConfigProvide != nil {
		log.Debug("加载远程配置")
		configProvide := defaultServiceConfig.RemoteConfigProvide
		// TODO
		viper.AddRemoteProvider(configProvide.Provider, configProvide.Endpoint, configProvide.Path)
		viper.ReadRemoteConfig()
	}
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
