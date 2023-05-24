package engine

import (
	"fmt"
	"github.com/coffeehc/boot/configuration"
	"github.com/sevlyar/go-daemon"
)

func getDaemonContext(serviceInfo configuration.ServiceInfo) *daemon.Context {
	return &daemon.Context{
		PidFileName: fmt.Sprintf("/var/run/%s.pid", serviceInfo.ServiceName),
		PidFilePerm: 0644,
		LogFileName: fmt.Sprintf("/var/logs/%s.log", serviceInfo.ServiceName),
		LogFilePerm: 0640,
		WorkDir:     "/",
		Umask:       027,
	}

}
