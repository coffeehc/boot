package internal

import (
	"context"

	"github.com/coffeehc/boot/configuration"
)

type RegisterCenter interface {
	Register(ctx context.Context, serviceInfo configuration.ServiceInfo) error
}
