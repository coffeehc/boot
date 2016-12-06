package grpcboot

import (
	"github.com/coffeehc/web"
	"google.golang.org/grpc"
	"strings"
)

type grpcFilter struct {
	server *grpc.Server
}

func (this *grpcFilter) filter(reply web.Reply, chain web.FilterChain) {
	request := reply.GetRequest()
	if request.ProtoMajor == 2 && strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
		reply.AdapterHttpHandler(true)
		this.server.ServeHTTP(reply.GetResponseWriter(), request)
		return
	}
	chain(reply)
}
