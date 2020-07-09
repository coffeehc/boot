package internal

import (
	"context"
	"sync"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/plugin"
	"go.uber.org/zap"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "service_register_manage"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Fatal("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetRegisterCenter() RegisterCenter
	SetRegisterCenter(center RegisterCenter) errors.Error
}

func EnablePlugin(ctx context.Context) {
	if name == "" {
		log.Fatal("插件名称没有初始化")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	service = &serviceImpl{}
	plugin.RegisterPluginByFast(name, nil, nil)
}
