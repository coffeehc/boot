package boot

import (
	"flag"
	"os"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitFlags() {
	if !pflag.Parsed() {
		// pflag.SetInterspersed(false)
		pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
		pflag.Parse()
		viper.BindPFlags(pflag.CommandLine) // viper绑定flags
	}
}

var modelInit = new(sync.Once)

func InitModel() {
	modelInit.Do(func() {
		model, ok := os.LookupEnv("ENV_RUN_MODEL")
		if !ok {
			if *runModel != "" {
				return
			}
			panic("没有指定运行模式,请设置环境变量：ENV_RUN_MODEL或参数--run_model")
		}
		*runModel = model
	})
}

//

const (
	Model_dev     = "dev"
	Model_test    = "test"
	Model_product = "product"
)

// IsDevModule 是否是开发模式
func IsProductModel() bool {
	return *runModel == Model_product
}

func RunModel() string {
	if *runModel == "" {
		panic("runmodel没有指定")
	}
	return *runModel

}
