package base

import "flag"

var (
	devmodule = flag.Bool("devmodule", false, "开发模式")
)

func IsDevModule() bool {
	return *devmodule
}
