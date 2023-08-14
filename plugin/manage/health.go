package manage

import (
	"github.com/gofiber/fiber/v2"
	"runtime"
	"time"

	"github.com/coffeehc/boot/configuration"
)

func RegisterServiceRuntimeInfoEndpoint(router *fiber.App) {
	serviceInfo := configuration.GetServiceInfo()
	h := &serviceRuntimeInfo{
		ServiceName: serviceInfo.ServiceName,
		Descriptor:  serviceInfo.Descriptor,
		Version:     serviceInfo.Version,
		GoVersion:   runtime.Version(),
		CPUNum:      runtime.NumCPU(),
		GoRach:      runtime.GOARCH,
		GoOS:        runtime.GOOS,
		StartTime:   time.Now(),
		Model:       configuration.GetRunModel(),
	}
	router.Get("/info", func(c *fiber.Ctx) error {
		return c.Format(h)
		//context.JSON(http.StatusOK, h)
	})
}

type serviceRuntimeInfo struct {
	ServiceName string    `json:"service_name"`
	Version     string    `json:"version"`
	Descriptor  string    `json:"descriptor"`
	GoVersion   string    `json:"go_version"`
	GoRach      string    `json:"go_rach"`
	GoOS        string    `json:"go_os"`
	CPUNum      int       `json:"cpu_num"`
	StartTime   time.Time `json:"start_time" time_format:"2006-01-02 15:04:05.999"`
	Model       string    `json:"model"`
}
