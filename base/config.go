package base

import "flag"

var (
	devModule = flag.Bool("dev", false, "开发模式")
)

//IsDevModule 是否是开发模式
func IsDevModule() bool {
	return *devModule
}
