package boot

import (
	"flag"
)

const (
	Ctx_Key_serviceName = "_serviceName"
	Ctx_Key_serviceInfo = "_serviceInfo"
)

var (
	devModule = flag.Bool("dev", true, "运行模式")
)

//IsDevModule 是否是开发模式
func IsDevModule() bool {
	return *devModule
}
