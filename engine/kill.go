//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || plan9
// +build darwin dragonfly freebsd linux netbsd openbsd plan9

package engine

import "syscall"

func Kill(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}
