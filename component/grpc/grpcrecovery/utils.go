package grpcrecovery

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	Header_XRequestId      = "x-request-id"
	Header_XB3Traceid      = "x-b3-traceid"
	Header_XB3Spanid       = "x-b3-spanid"
	Header_XB3Parentspanid = "x-b3-parentspanid"
	Header_XB3Sampled      = "x-b3-sampled"
	Header_XB3Flags        = "x-b3-flags"
	Header_XotSpanContext  = "x-ot-span-context"
)

var (
	MashTracingHeaders = []string{Header_XRequestId, Header_XB3Traceid, Header_XB3Spanid, Header_XB3Parentspanid, Header_XB3Sampled, Header_XB3Flags, Header_XotSpanContext}
)

func BuildMetadataFromContext(ctx context.Context) metadata.MD {
	md := metadata.New(make(map[string]string, 7))
	for _, headerKey := range MashTracingHeaders {
		v := ctx.Value(headerKey)
		if vString, ok := v.(string); ok {
			md.Set(headerKey, vString)
		}
	}
	return md
}

func ParseMetadataToContext(ctx context.Context, md metadata.MD) context.Context {
	for _, headerKey := range MashTracingHeaders {
		v := md.Get(headerKey)
		if v != nil && len(v) > 0 {
			ctx = context.WithValue(ctx, headerKey, v[0])
		}
	}
	return ctx
}
