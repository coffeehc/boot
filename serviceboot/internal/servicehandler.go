package internal

import (
	"fmt"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
)

const err_scope_service_handler = "startService"

//StopService 停止服务
func StopService(service base.Service) {
	if service != nil && service.Stop != nil {
		stopErr := service.Stop()
		if stopErr != nil {
			logger.Error("关闭服务失败,%s\n", stopErr)
		}
	}
}

//StartService 启动服务
func StartService(service base.Service) (err base.Error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			if e, ok := err1.(base.Error); ok {
				err = e
				return
			}
			err = base.NewError(base.ErrCodeBaseSystemUnknown, err_scope_service_handler, fmt.Sprintf("service crash,cause is %s", err1))
		}
	}()
	if service == nil {
		panic(base.NewError(base.ErrCodeBaseSystemInit, err_scope_service_handler, "没有 Service 的实例"))
	}
	if service.Run == nil {
		panic(base.NewError(base.ErrCodeBaseSystemInit, err_scope_service_handler, "没有指定Run方法"))
	}
	err1 := service.Run()
	if err1 != nil {
		panic(err1)
	}
	logger.Info("服务已正常启动")
	return
}
