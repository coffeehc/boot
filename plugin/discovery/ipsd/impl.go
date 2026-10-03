package ipsd

import (
	"context"

	"google.golang.org/grpc/resolver"
)

// GetResolverBuilder 校验初始地址并创建可复用的 builder；空列表可等待后续 UpdateAddress。
// ctx 是装配上下文，订阅生命周期由各连接的 Close 管理。
func GetResolverBuilder(_ context.Context, defaultSrvAddr ...string) (ResolverBuilder, error) {
	rb := &IpResolverBuilder{}
	if err := rb.UpdateAddress(defaultSrvAddr); err != nil {
		return nil, err
	}
	return rb, nil
}

// ResolverBuilder 为静态多地址连接提供解析和更新能力，所有方法均支持并发调用。
type ResolverBuilder interface {
	resolver.Builder
	// UpdateAddress 替换地址快照；非法输入返回错误并保留原地址。
	UpdateAddress(addresses []string) error
}
