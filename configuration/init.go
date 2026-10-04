package configuration

import (
	"flag"
	"os"
	"strings"

	"github.com/coffeehc/base/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	Model_dev     = "dev"
	Model_test    = "test"
	Model_product = "prod"
)

var configFile = pflag.StringP("config", "c", "./config.yml", "配置文件路径")

var runModel = ""

// defaultRunModel 保存 Boot 拥有的默认值，文件模式重建配置时不继承上一轮 Viper 状态。
var defaultRunModel string

// Option 配置一次 InitConfiguration；初始化由启动 owner 串行执行，不影响下一轮默认行为。
type Option func(*configurationOptions)

// configurationOptions 保存本次初始化允许读取的配置来源。
type configurationOptions struct {
	// fileOnly 排除环境绑定、上一轮文件值和程序覆盖值，仅保留 Boot 默认值和本次文件。
	fileOnly bool
}

// WithFileOnly 要求本次初始化只读取配置文件和 Boot 默认值。
// config CLI 参数仍用于选择文件；资源值不接受环境变量或其他 CLI 参数覆盖。
// 初始化重建全局 Viper，应用应在初始化后注册程序默认值，不能依赖上一轮 Set/BindEnv。
func WithFileOnly() Option {
	return func(options *configurationOptions) { options.fileOnly = true }
}

// GetRunModel 返回当前初始化已经固定的运行模式，不重新读取配置来源。
func GetRunModel() string {
	return runModel
}

// loadConfig 按本轮来源策略装配配置；文件模式必须清除上轮开启的环境自动读取。
func loadConfig(options configurationOptions) {
	if options.fileOnly {
		// Viper 没有关闭 AutomaticEnv 或清除 BindEnv 的公开 API；重建防止跨轮残留。
		viper.Reset()
		if defaultRunModel != "" {
			viper.SetDefault(_run_model, defaultRunModel)
		}
	}
	viper.SetConfigType("yaml")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	if !pflag.Parsed() {
		pflag.Parse()
	}
	if !options.fileOnly {
		viper.BindPFlags(pflag.CommandLine)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.SetEnvPrefix("ENV")
		viper.AutomaticEnv()
	}
	_, err := os.ReadFile(*configFile)
	if err == nil {
		viper.SetConfigFile(*configFile)
		if err := viper.MergeInConfig(); err != nil {
			log.Warn("加载日志文件失败:", zap.Error(err))
		}
	}
	if viper.GetString(_run_model) == "" {
		log.Panic("没有指定run model")
	}
	runModel = viper.GetString(_run_model)
	log.Info("加载配置", zap.String("run model", viper.GetString(_run_model)))
}
