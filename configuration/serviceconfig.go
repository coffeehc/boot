package configuration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/base/utils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var onConfigChanges = make([]func(), 0)
var currentServiceInfo ServiceInfo
var rootCtx context.Context
var EnableRemoteLog = false

var configurationState struct {
	sync.RWMutex
	owner *configOwner
}

func RegisterOnConfigChange(onConfigChange func()) {
	configurationState.Lock()
	defer configurationState.Unlock()
	onConfigChanges = append(onConfigChanges, onConfigChange)
}

// InitConfiguration 串行初始化服务配置和身份；默认允许 ENV_ 环境覆盖。
// WithFileOnly 将本轮限制为文件和 Boot 默认值，不继承之前初始化的配置来源。
func InitConfiguration(ctx context.Context, serviceInfo ServiceInfo, opts ...Option) error {
	configurationState.Lock()
	defer configurationState.Unlock()
	if configurationState.owner != nil {
		configurationState.owner.stop()
		configurationState.owner = nil
	}
	if ctx == nil {
		return fmt.Errorf("配置初始化 context 不能为空")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if serviceInfo.Metadata == nil {
		serviceInfo.Metadata = map[string]string{
			"git_rev": GitRev,
			// "build_version": BuildVersion,
			"build_time": BuildTime,
			"git_tag":    GitTag,
			"version":    Version,
		}
	}
	options := configurationOptions{debounce: 5 * time.Second}
	for _, option := range opts {
		if option != nil {
			option(&options)
		}
	}
	if options.debounce <= 0 {
		return fmt.Errorf("配置防抖窗口必须大于零")
	}
	path, err := loadConfig(options)
	if err != nil {
		return err
	}
	// 先用临时日志值检查类型，避免 base/log 的旧错误日志带出原始配置值。
	var loggerConfig log.Config
	if err := viper.UnmarshalKey("logger", &loggerConfig); err != nil {
		return nodeError("logger", "解码失败")
	}
	switch loggerConfig.Level {
	case "", "debug", "info", "warn", "warning", "error", "dpanic", "panic", "fatal":
	default:
		return nodeError("logger", "日志级别无效")
	}
	configurationState.owner = newConfigOwner(ctx, viper.GetViper(), path, options.debounce)
	initServiceInfo(ctx, serviceInfo)
	// loadRemoteConfig(ctx, serviceInfo)
	// base/log 已在其包初始化中建立 logger；InitLogger(true) 会另开 Viper
	// 原生 watcher 并提前读取文件，破坏本包防抖和串行所有权。只同步加载日志值。
	log.LoadConfig()
	return nil
}

func initServiceInfo(ctx context.Context, serviceInfo ServiceInfo) {
	rootCtx = ctx
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
	configurationState.RLock()
	defer configurationState.RUnlock()
	return currentServiceInfo.ServiceName
}

func GetServiceInfo() ServiceInfo {
	configurationState.RLock()
	defer configurationState.RUnlock()
	return currentServiceInfo
}

// SetRunModel 注册 Boot 运行模式默认值；每次文件配置初始化均恢复此默认值。
func SetRunModel(model string) {
	configurationState.Lock()
	defer configurationState.Unlock()
	if o := configurationState.owner; o != nil {
		o.mu.Lock()
		defer o.mu.Unlock()
	}
	defaultRunModel = model
	viper.SetDefault(_run_model, model)
}
