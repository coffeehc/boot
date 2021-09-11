package internal

import (
	"context"
	"sync"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "service_register_manage"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetRegisterCenter() RegisterCenter
	SetRegisterCenter(center RegisterCenter) errors.Error
}

func EnablePlugin(ctx context.Context) {
	if name == "" {
		log.Panic("插件名称没有初始化")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	service = &serviceImpl{}
	plugin.RegisterPluginByFast(name, nil, nil)
}
