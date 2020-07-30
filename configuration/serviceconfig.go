package configuration

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/base/utils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var remoteConfigProvide *RemoteConfigProvide
var currentServiceInfo ServiceInfo
var mutex = new(sync.RWMutex)
var rootCtx context.Context

func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo) {
	loadConfig()
	initDefaultLoggerConfig()
	initServiceInfo(ctx, serviceInfo)
	initRemoteConfigProvide(ctx)
}

func initServiceInfo(ctx context.Context, serviceInfo ServiceInfo) {
	if rootCtx == nil {
		rootCtx = ctx
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
	log.Debug("加载服务信息", zap.Any("serviceInfo", serviceInfo))
}

func initRemoteConfigProvide(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if remoteConfigProvide != nil {
		return
	}
	if rootCtx == nil {
		rootCtx = ctx
	}
	if !viper.IsSet("RemoteConfigProvide") {
		return
	}
	provider := &RemoteConfigProvide{}
	err := viper.UnmarshalKey("RemoteConfigProvide", provider)
	if err != nil {
		log.Fatal("不能从配置中读取远程服务配置", zap.Error(err))
	}
	if provider.Endpoint == "" || provider.Path == "" || provider.Provider == "" {
		log.Fatal("没有配置中读取远程服务属性", zap.Error(err))
	}
	remoteConfigProvide = provider
	viper.AddRemoteProvider(provider.Provider, provider.Endpoint, provider.Path)
	viper.ReadRemoteConfig()
	viper.WatchRemoteConfig()
}

func GetRunModel() string {
	// return defaultServiceConfig.RunModel
	return viper.GetString(_run_model)
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
