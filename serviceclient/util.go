package serviceclient

import (
	"fmt"
	"strings"

	"github.com/coffeehc/microserviceboot/base"
)

func buildServiceDomain(serviceName string, serviceClientDNSConfig ServiceClientConsulConfig) string {
	tag := ""
	if base.IsDevModule() {
		tag = "dev"
	}
	if len(tag) > 0 {
		tag += "."
	}
	return fmt.Sprintf("%s%s.service.%s.%s", tag, serviceName, serviceClientDNSConfig.GetDataCenter(), serviceClientDNSConfig.GetDomain())
}

func WarpUrl(restUrl string, pathParams map[string]string) string {
	//TODO 此处可以优化
	for k, v := range pathParams {
		restUrl = strings.Replace(restUrl, "{"+k+"}", v, -1)
	}
	return restUrl
}

func addMissingPort(addr string, isTLS bool) string {
	n := strings.Index(addr, ":")
	if n >= 0 {
		return addr
	}
	port := 80
	if isTLS {
		port = 443
	}
	return fmt.Sprintf("%s:%d", addr, port)
}
