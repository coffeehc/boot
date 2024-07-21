package grpcclient

import (
	"context"
	"google.golang.org/grpc/credentials"
)

const perRPCCredentialsKey = "_grpc._PerRPCCredentialsKey"

func SetPerRPCCredentials(ctx context.Context, prc credentials.PerRPCCredentials) context.Context {
	return context.WithValue(ctx, perRPCCredentialsKey, prc)
}

func GetPerRPCCredentials(ctx context.Context) credentials.PerRPCCredentials {
	v := ctx.Value(perRPCCredentialsKey)
	if v == nil {
		return nil
	}
	return v.(credentials.PerRPCCredentials)
}
