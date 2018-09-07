package boot

import (
	"flag"
	"os"
	"sync"
)

var modelInit = new(sync.Once)

var runModel = flag.String("run_model", "", "运行模式,必填（dev，test，product或其他）")

func InitModel() {
	modelInit.Do(func() {
		if !flag.Parsed() {
			flag.Parse()
		}
		model, ok := os.LookupEnv("ENV_RUN_MODEL")
		if !ok {
			if *runModel != "" {
				return
			}
			panic("没有指定运行模式")
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

//IsDevModule 是否是开发模式
func IsProductModel() bool {
	return *runModel == Model_product
}

func RunModel() string {
	return *runModel

}
