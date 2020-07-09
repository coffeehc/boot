package ipsd

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
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
