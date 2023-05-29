//go:build !windows && !plan9
// +build !windows,!plan9

package engine

import (
	"fmt"
	"syscall"
)

func NewSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}

func GetLogFile(serviceName string) string {
	return fmt.Sprintf("/var/log/%s.log", serviceName)
}

func GetPidFilePath(serviceName string) string {
	return fmt.Sprintf("/var/run/%s.pid", serviceName)
}
