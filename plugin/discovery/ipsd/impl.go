package ipsd

import (
	"context"

	"google.golang.org/grpc/resolver"
)

type serviceImpl struct {
}

func (impl *serviceImpl) GetResolverBuilder(ctx context.Context, defaultSrvAddr ...string) (resolver.Builder, error) {
	rb := &resolverBuilder{
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
	}
	return rb, nil
}
