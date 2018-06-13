package main

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type _GreeterServer struct {
	logger       *zap.Logger
	errorService errors.Service
}

func (server *_GreeterServer) SayHello(cxt context.Context, request *simplemodel.Request) (*simplemodel.Response, error) {
	server.logger.Debug("接收到客户端的请求")
	if request.GetId()%3 == 0 {
		server.logger.Debug("发送一条系统错误")
		panic(server.errorService.MessageError("系统错误"))
	}
	if request.GetId()%2 == 0 {
		server.logger.Debug("发送一条消息错误")
		return nil, server.errorService.MessageError("消息错误")
	}
	response := new(simplemodel.Response)
	response.Message = fmt.Sprintf("%s-->%s", request.Name, time.Now().String())
	return response, nil
}
