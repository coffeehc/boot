package grpcx

import (
	"context"
	"crypto/tls"
	"github.com/coffeehc/base/log"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/credentials"
)

const (
	contextKeyCerds = "_grpc.Credentials"
)

func SetInsecureSkipVerifyCerds(ctx context.Context) context.Context {
	tlsConfig := &tls.Config{
		NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
		InsecureSkipVerify: true,
	}
	return SetCerds(ctx, credentials.NewTLS(tlsConfig))
}

func SetCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	if ctx.Value(contextKeyCerds) != nil {
		log.DPanic("****已经设置了TransportCredentials,不能多次设置****")
	}
	return context.WithValue(ctx, contextKeyCerds, creds)
}

func GetCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextKeyCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}
