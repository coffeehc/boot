package sd

import (
	"fmt"

	"git.xiagaogao.com/coffee/boot"
)

type ServiceRegisterInfo struct {
	ServiceInfo    *boot.ServiceInfo `json:"info"`
	ServerAddr     string            `json:"server_addr"`
	ManageEndpoint string            `json:"manage_endpoint"`
	Data           string            `json:"data"`
}

func BuildServiceKeyPrefix(serviceInfo *boot.ServiceInfo) string {
	return fmt.Sprintf("/ms/registers/%s/%s/", serviceInfo.ServiceName, boot.RunModel())
}
