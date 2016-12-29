package serviceboot

import (
	"runtime"

	"github.com/coffeehc/httpx"
	"github.com/coffeehc/microserviceboot/base"
)

func (h *health) health(reply httpx.Reply) {
	h.GoRoutine = runtime.NumGoroutine()
	reply.With(h).As(httpx.DefaultRenderJSON)
}

type health struct {
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	GoVersion   string `json:"go_version"`
	GoRoutine   int    `json:"go_routine"`
	CPUNum      int    `json:"cpu_num"`
	GoRach      string `json:"go_rach"`
	GoOS        string `json:"go_os"`
}

func newHealth(serviceInfo base.ServiceInfo) *health {
	return &health{
		ServiceName: serviceInfo.GetServiceName(),
		Tag:         serviceInfo.GetServiceTag(),
		Version:     serviceInfo.GetVersion(),
		GoVersion:   runtime.Version(),
		CPUNum:      runtime.NumCPU(),
		GoRach:      runtime.GOARCH,
		GoOS:        runtime.GOOS,
	}
}
