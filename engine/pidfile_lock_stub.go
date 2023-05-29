//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!plan9,!solaris

package daemon

import "github.com/coffeehc/base/errors"

func lockFile(fd uintptr) error {
	return errors.SystemError("操作不支持")
}

func unlockFile(fd uintptr) error {
	return errors.SystemError("操作不支持")
}
