//go:build !linux && !windows
// +build !linux,!windows

package engine

import "github.com/coffeehc/boot/configuration"

func daemon(serviceInfo configuration.ServiceInfo) error {
	return nil
}
