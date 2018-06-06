package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
)

//ServiceDiscoveryRegister 服务注册接口
type ServiceDiscoveryRegister interface {
	//注册服务
	RegService(cxt context.Context, info ServiceInfo, serviceAddr string) (deregister func(), err errors.Error)
}
