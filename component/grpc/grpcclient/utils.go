package grpcclient

import (
	"context"
	"google.golang.org/grpc/credentials"
)

const (
	contextKeyServerCerds = "_grpc.clientCredentials"
	perRPCCredentialsKey  = "_grpc._PerRPCCredentialsKey"
)

func SetCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	return context.WithValue(ctx, contextKeyServerCerds, creds)
}

func getCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextKeyServerCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}

func SetAuthService(ctx context.Context, prc credentials.PerRPCCredentials) context.Context {
	return context.WithValue(ctx, perRPCCredentialsKey, prc)
}
