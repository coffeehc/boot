package boot

import (
	"flag"
)

var (
	devModule = flag.Bool("dev", true, "运行模式")
)

//IsDevModule 是否是开发模式
func IsDevModule() bool {
	return *devModule
}

func RunModule() string {
	if *devModule {
		return "dev"
	}
	return "product"

}
