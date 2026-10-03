package kubernetes

import (
	"context"
	"slices"

	"google.golang.org/grpc/resolver"
)

// serviceImpl 提供连接级 DNS resolver 装配，不在插件启用时访问网络。
type serviceImpl struct{}

// GetResolverBuilder 返回可复用的 builder，配置错误由 gRPC 调用 Build 时返回。
func (impl *serviceImpl) GetResolverBuilder(_ context.Context, configs ...*Config) resolver.Builder {
	return &resolverBuilder{configs: slices.Clone(configs)}
}
