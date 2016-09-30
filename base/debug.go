package base

import (
	"runtime/debug"

	"github.com/coffeehc/logger"
)

func DebugPanic(printStick bool) {
	if err := recover(); err != nil {
		logger.Error("发生错误:%#v", err)
		if printStick {
			debug.PrintStack()
		}
		panic(err)
	}
}
