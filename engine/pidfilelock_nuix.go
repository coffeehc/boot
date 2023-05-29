//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || plan9
// +build darwin dragonfly freebsd linux netbsd openbsd plan9

package engine

import (
	"github.com/coffeehc/base/errors"
	"syscall"
)

func lockFile(fd uintptr) error {
	err := syscall.Flock(int(fd), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == syscall.EWOULDBLOCK {
		err = errors.SystemError("文件已经被锁定")
	}
	return err
}

func unlockFile(fd uintptr) error {
	err := syscall.Flock(int(fd), syscall.LOCK_UN)
	if err == syscall.EWOULDBLOCK {
		err = errors.SystemError("文件已经被锁定")
	}
	return err
}
