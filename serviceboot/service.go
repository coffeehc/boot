package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"google.golang.org/grpc"
)

// Service 接口定义
type Service interface {
	Init(cxt context.Context, serviceBoot ServiceBoot) errors.Error
	Run(cxt context.Context) errors.Error
	Stop(cxt context.Context) errors.Error
	RegisterServer(s *grpc.Server) errors.Error
}
