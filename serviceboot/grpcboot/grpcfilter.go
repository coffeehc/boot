package grpcboot

import (
	"strings"

	"github.com/coffeehc/httpx"
	"google.golang.org/grpc"
)

type grpcFilter struct {
	server *grpc.Server
}

func (gf *grpcFilter) filter(reply httpx.Reply, chain httpx.FilterChain) {
	request := reply.GetRequest()
	if request.ProtoMajor == 2 && strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
		reply.AdapterHTTPHandler(true)
		gf.server.ServeHTTP(reply.GetResponseWriter(), request)
		return
	}
	chain(reply)
}
