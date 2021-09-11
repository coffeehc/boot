package manage

import (
	"net/http"
	"runtime"
	"time"

	"github.com/coffeehc/boot/configuration"
	"github.com/gin-gonic/gin"
)

func (impl *serviceImpl) registerServiceRuntimeInfoEndpoint(router gin.IRouter) {
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
	router.GET("/info", func(context *gin.Context) {
		context.JSON(http.StatusOK, h)
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

func (impl *serviceImpl) registerHealthEndpoint(router gin.IRouter) {
	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"goroutine_count": runtime.NumGoroutine(),
		})
	})
}
