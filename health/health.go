package health

import (
	"runtime"

	"git.xiagaogao.com/coffee/boot/serviceboot"
	"github.com/coffeehc/httpx"
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

func newHealth(serviceInfo serviceboot.ServiceInfo) *health {
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
