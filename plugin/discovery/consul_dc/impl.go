package consul_dc

import (
	"context"
	"slices"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

// serviceImpl 持有并发安全的 Consul 客户端，供各连接独立执行阻塞查询。
type serviceImpl struct {
	// client 由默认配置构建或由调用方注入，插件不修改其配置或关闭共享传输。
	client *api.Client
}

// GetResolverBuilder 为连接装配查询选项，配置错误由 gRPC 调用 Build 时返回。
func (impl *serviceImpl) GetResolverBuilder(_ context.Context, queries ...*api.QueryOptions) resolver.Builder {
	return &resolverBuilder{client: impl.client, queries: slices.Clone(queries), scheme: ServiceProtocolScheme}
}
