package grpcclient

import (
	"context"
	"crypto/tls"
	"github.com/coffeehc/base/log"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/credentials"
)

const (
	perRPCCredentialsKey  = "_grpc._PerRPCCredentialsKey"
	contextKeyClientCerds = "_grpc.client.Credentials"
)

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

func SetInsecureSkipVerifyCerds(ctx context.Context) context.Context {
	tlsConfig := &tls.Config{
		NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
		InsecureSkipVerify: true,
	}
	return SetClientCerds(ctx, credentials.NewTLS(tlsConfig))
}

func SetClientCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	if ctx.Value(contextKeyClientCerds) != nil {
		log.DPanic("****已经设置了TransportCredentials,不能多次设置****")
	}
	return context.WithValue(ctx, contextKeyClientCerds, creds)
}

func GetClientCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextKeyClientCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}
