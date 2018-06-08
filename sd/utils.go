package sd

import (
	"fmt"

	"git.xiagaogao.com/coffee/boot"
)

type ServiceRegisterInfo struct {
	ServiceInfo *boot.SimpleServiceInfo `json:"info"`
	ServerAddr  string                  `json:"server_addr"`
}

func BuildServiceKeyPrefix(serviceInfo boot.ServiceInfo) string {
	return fmt.Sprintf("/ms/registers/%s/%s/", serviceInfo.GetServiceName(), serviceInfo.GetServiceTag())
}
