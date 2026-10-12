package testutils

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer.
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestGrpc(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	serveDone := make(chan error, 1)
	go func() { serveDone <- s.Serve(l) }()
	t.Cleanup(func() {
		s.Stop()
		select {
		case err := <-serveDone:
			if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				t.Errorf("serve: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("gRPC server did not stop")
		}
	})
	c, err := grpc.NewClient("passthrough:///"+l.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithNoProxy(),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})
	client := pb.NewGreeterClient(c)
	for _, name := range []string{"coffee", "", "咖啡"} {
		t.Run("name="+name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			r, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
			if err != nil {
				t.Fatalf("SayHello: %v", err)
			}
			if got, want := r.GetMessage(), "Hello "+name; got != want {
				t.Errorf("SayHello message = %q, want %q", got, want)
			}
		})
	}
}
