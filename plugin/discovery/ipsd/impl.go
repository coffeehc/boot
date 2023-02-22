package ipsd

import (
	"context"

	"google.golang.org/grpc/resolver"
)

func GetResolverBuilder(ctx context.Context, defaultSrvAddr ...string) (resolver.Builder, error) {
	rb := &resolverBuilder{
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
	}
	return rb, nil
}

type ResolverBuilder interface {
	UpdateAddress(addresses []string)
}
