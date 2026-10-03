package internal

import (
	"context"

	"github.com/coffeehc/boot/configuration"
)

// RegisterCenter 在 RPC 启动后向注册中心登记实例，调用方通过 context 约束请求生命周期。
// 实现负责线程安全和自己登记实例的停止清理，登记错误交由 plugin 启动回滚处理。
type RegisterCenter interface {
	// Register 登记服务信息，metadata 属于调用方，不得原地修改；失败返回错误。
	Register(ctx context.Context, serviceInfo configuration.ServiceInfo) error
}
