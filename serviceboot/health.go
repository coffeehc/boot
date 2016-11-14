package serviceboot

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"runtime"
)

func (this *Health) health(reply web.Reply) {
	this.GoRoutine = runtime.NumGoroutine()
	reply.With(this).As(web.Default_Render_Json)
}

type Health struct {
	ServiceName string   `json:"service_name"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
	GoVersion   string   `json:"go_version"`
	GoRoutine   int      `json:"go_routine"`
	CpuNum      int      `json:"cpu_num"`
	GoRach      string   `json:"go_rach"`
	GoOS        string   `json:"go_os"`
}

func newHealth(serviceInfo base.ServiceInfo) *Health {
	return &Health{
		ServiceName: serviceInfo.GetServiceName(),
		Tags:        serviceInfo.GetServiceTags(),
		Version:     serviceInfo.GetVersion(),
		GoVersion:   runtime.Version(),
		CpuNum:      runtime.NumCPU(),
		GoRach:      runtime.GOARCH,
		GoOS:        runtime.GOOS,
	}
}
