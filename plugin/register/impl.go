package register

import (
	"context"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin/register/internal"
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
