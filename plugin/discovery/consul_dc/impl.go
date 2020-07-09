package consul_dc

import (
	"context"

	"google.golang.org/grpc/resolver"
)

type serviceImpl struct {
}

func newService() *serviceImpl {
	impl := &serviceImpl{}
	return impl
}

func (impl *serviceImpl) GetResolverBuilder(ctx context.Context) resolver.Builder {
	return &resolverBuilder{
		ctx: ctx,
	}
}
