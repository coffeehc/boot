package register

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin/register/internal"
)

type serviceImpl struct {
	registerManage internal.Service
}

func (impl *serviceImpl) Start(ctx context.Context) errors.Error {
	register := impl.registerManage.GetRegisterCenter()
	if register == nil {
		log.Panic("没有可用的注册中心")
	}
	serviceInfo := configuration.GetServiceInfo()
	return register.Register(ctx, serviceInfo)
}
func (impl *serviceImpl) Stop(ctx context.Context) errors.Error {
	return nil
}
