package main

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"golang.org/x/net/context"
)

type _GreeterServer struct {
}

func (server *_GreeterServer) SayHello(cxt context.Context, request *simplemodel.Request) (*simplemodel.Response, error) {
	response := new(simplemodel.Response)
	response.Message = fmt.Sprintf("%s-->%s", request.Name, time.Now().String())
	return response, nil
}
