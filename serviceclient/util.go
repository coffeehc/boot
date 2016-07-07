package serviceclient

import (
	"fmt"
	"strings"

	"github.com/coffeehc/microserviceboot/base"
)

func buildServiceDomain(serviceName string, serviceClientDNSConfig ServiceClientConsulConfig) string {
	tag := "pro"
	if base.IsDevModule() {
		tag = "dev"
	}
	return fmt.Sprintf("%s.%s.service.%s.%s", tag, serviceName, serviceClientDNSConfig.GetDataCenter(), serviceClientDNSConfig.GetDomain())
}

func WarpUrl(restUrl string, pathParams map[string]string) string {
	//TODO 此处可以优化
	for k, v := range pathParams {
		restUrl = strings.Replace(restUrl, "{"+k+"}", v, -1)
	}
	return restUrl
}
