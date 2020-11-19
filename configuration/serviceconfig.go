package configuration

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/base/utils"
	"git.xiagaogao.com/coffee/boot/component/consul"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var onConfigChanges = make([]func(), 0)
var currentServiceInfo ServiceInfo
var mutex = new(sync.RWMutex)
var rootCtx context.Context

func EnableRemoteConfig() {
	viper.Set("remote_config.enable", true)
}

func RegisterOnConfigChange(onConfigChange func()) {
	onConfigChanges = append(onConfigChanges, onConfigChange)
}

func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo) {
	viper.SetConfigType("yaml")
	loadConfig()
	initServiceInfo(ctx, serviceInfo)
	loadRemoteConfig(ctx, serviceInfo)
	log.InitLogger(true)
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
	log.ResetLogger(zap.String("serviceName", serviceInfo.ServiceName), zap.String("localIp", localIp.String()))
	log.Info("加载服务信息", zap.Any("serviceInfo", serviceInfo))
}

func loadRemoteConfig(ctx context.Context, serviceInfo ServiceInfo) {
	log.Info("远程配置开关", zap.Bool("enable", viper.GetBool("remote_config.enable")))
	if !viper.GetBool("remote_config.enable") {
		return
	}
	consul.EnablePlugin(ctx)
	path := fmt.Sprintf("configs/%s/config_%s.yaml", serviceInfo.ServiceName, GetRunModel())
	consulService := consul.GetService()
	kv := consulService.GetConsulClient().KV()
	opts := &api.QueryOptions{
		WaitIndex: 0,
	}
	opts = opts.WithContext(ctx)
	err := readRemotConfig(ctx, path, kv, opts)
	if err != nil {
		log.Fatal("读取远程配置失败", err.GetFieldsWithCause()...)
	}
	go func() {
		for {
			err := readRemotConfig(ctx, path, kv, opts)
			if err != nil {
				log.Error("读取远程配置失败", err.GetFieldsWithCause()...)
				time.Sleep(time.Second * 5)
			}
		}
	}()
}

func readRemotConfig(ctx context.Context, path string, kv *api.KV, opts *api.QueryOptions) errors.Error {
	if ctx.Err() != nil {
		return errors.ConverError(ctx.Err())
	}
	kvpair, meta, err := kv.Get(path, opts)
	if kvpair == nil && err == nil {
		log.Warn("找不到对应的key", zap.String("path", path))
		return errors.SystemError("找不到对应的配置Key")
	}
	if err != nil {
		log.Error("获取远程配置失败", zap.Error(err))
		return errors.SystemError("获取远程配置失败")
	}
	opts.WaitIndex = meta.LastIndex
	err = viper.MergeConfig(bytes.NewReader(kvpair.Value))
	if err != nil {
		log.Fatal("读取远程配置失败", zap.Error(err), zap.String("path", path))
	}
	log.Info("远程配置已变更，需要重新加载配置", zap.Uint64("lastIndex", meta.LastIndex), zap.String("raw", string(kvpair.Value)))
	log.LoadConfig()
	for _, onConfigChange := range onConfigChanges {
		onConfigChange()
	}
	return nil
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
