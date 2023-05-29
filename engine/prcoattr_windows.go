//go:build windows
// +build windows

package engine

import "syscall"

func NewSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func GetLogFile(serviceName string) string {
	return fmt.Sprintf("./%s.log", serviceName)
}

func GetPidFilePath(serviceName string) string {
	return fmt.Sprintf("./%s.pid", serviceName)
}
