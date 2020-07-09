package internal

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/boot/configuration"
)

type RegisterCenter interface {
	Register(ctx context.Context, serviceInfo configuration.ServiceInfo) errors.Error
}
