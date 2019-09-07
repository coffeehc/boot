package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"google.golang.org/grpc"
)

// Service 接口定义
type Service interface {
	Init(cxt context.Context, serviceKit ServiceKit) xerror.Error
	Run(cxt context.Context) xerror.Error
	Stop(cxt context.Context) xerror.Error
	RegisterServer(s *grpc.Server) xerror.Error
}
