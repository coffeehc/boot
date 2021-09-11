package ipsd

import (
	"context"

	"github.com/coffeehc/base/errors"
	"google.golang.org/grpc/resolver"
)

type serviceImpl struct {
}

func (impl *serviceImpl) GetResolverBuilder(ctx context.Context, defaultSrvAddr ...string) (resolver.Builder, errors.Error) {
	rb := &resolverBuilder{
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
	}
	return rb, nil
}
