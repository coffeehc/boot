package grpcboot

import (
	"github.com/coffeehc/web"
	"google.golang.org/grpc"
)

type grpcFilter struct {
	server *grpc.Server
}

func (this *grpcFilter) filter(reply web.Reply, chain web.FilterChain) {
	//TODO web 改版后直接使用 NotFountHandler
	request := reply.GetRequest()
	if request.ProtoMajor == 2 && request.Header.Get("content-type") {
		reply.AdapterHttpHandler(true)
		this.server.ServeHTTP(reply.GetResponseWriter(), reply.GetRequest())
		return
	}
	chain(reply)
}
