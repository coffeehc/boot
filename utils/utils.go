package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

/*
	wait,一般是可执行函数的最后用于阻止程序退出
*/
func WaitStop() {
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("接收到指令:%s,立即关闭程序", sig)
}
