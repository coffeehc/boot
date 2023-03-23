package testutils

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestGrpc(t *testing.T) {
	InitTestConfig()
	s := grpc.NewServer()
	l, err := net.Listen("tcp", ":8899")
	if err != nil {
		log.Error("启动端口失败", zap.Error(err))
	}
	pb.RegisterGreeterServer(s, &server{})
	go func() {
		err = s.Serve(l)
		if err != nil {
			log.Error("启动GRPC Server失败", zap.Error(err))
		}
	}()
	time.Sleep(time.Second * 3)
	c, err := grpc.Dial("192.168.3.128:8899", grpc.WithNoProxy())
	if err != nil {
		log.Error("创建链接失败", zap.Error(err))
	}
	client := pb.NewGreeterClient(c)
	for {
		r, err := client.SayHello(context.Background(), &pb.HelloRequest{
			Name: "coffee",
		})
		if err != nil {
			log.Error("错误", zap.Error(err))
			t.FailNow()
		}
		log.Debug("r", zap.String("message", r.Message))
		time.Sleep(time.Second * 5)
	}

}
