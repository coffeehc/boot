//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !plan9 && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!plan9,!solaris

package engine

import (
	"os"
	"syscall"
)

func Kill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	return nil
}
