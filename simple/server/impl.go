package main

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"golang.org/x/net/context"
)

type _GreeterServer struct {
}

func (server *_GreeterServer) SayHello(cxt context.Context, request *simplemodel.Request) (*simplemodel.Response, error) {
	logger := logs.GetLogger(cxt)
	logger.Debug("接收到客户端的请求")
	response := new(simplemodel.Response)
	response.Message = fmt.Sprintf("%s-->%s", request.Name, time.Now().String())
	return response, nil
}
