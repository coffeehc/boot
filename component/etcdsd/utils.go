package etcdsd

import (
	"fmt"

	"git.xiagaogao.com/coffee/boot/configuration"
)

func BuildServiceKeyPrefix() string {
	return fmt.Sprintf("/ms/registers/%s/%s/", configuration.GetServiceName(), configuration.GetModel())
}
