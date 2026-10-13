package rpc

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/coffeehc/boot/configuration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// readyFixture 使用真实构造流程和 loopback 临时端口，不连接外部服务。
func readyFixture(t *testing.T) *serviceImpl {
	t.Helper()
	impl := newService(t.Context()).(*serviceImpl)
	impl.rpcServerAddr = "127.0.0.1:0"
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := impl.Stop(ctx); err != nil {
			t.Errorf("cleanup Stop: %v", err)
		}
	})
	return impl
}

func assertRPCReady(t *testing.T, impl *serviceImpl, want bool) {
	t.Helper()
	if err := impl.Ready(t.Context()); (err == nil) != want {
		t.Fatalf("Ready = %v, want ready %t", err, want)
	}
	wantStatus := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if want {
		wantStatus = grpc_health_v1.HealthCheckResponse_SERVING
	}
	for _, serviceName := range []string{"", configuration.GetServiceInfo().ServiceName} {
		response, err := impl.healthServer.Check(t.Context(), &grpc_health_v1.HealthCheckRequest{Service: serviceName})
		if err != nil || response.GetStatus() != wantStatus {
			t.Fatalf("health(%q) = %v, %v; want %s", serviceName, response, err, wantStatus)
		}
	}
}

func assertNetworkRPCReady(t *testing.T, impl *serviceImpl) {
	t.Helper()
	connection, err := grpc.NewClient("passthrough:///"+impl.GetRPCServerAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithNoProxy())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	response, err := grpc_health_v1.NewHealthClient(connection).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil || response.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("network health = %v, %v; want SERVING", response, err)
	}
}

func TestReadyLifecycle(t *testing.T) {
	impl := readyFixture(t)
	assertRPCReady(t, impl, false)
	if err := impl.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	assertRPCReady(t, impl, true)
	assertNetworkRPCReady(t, impl)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := impl.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	assertRPCReady(t, impl, false)
}

func TestReadyHonorsContext(t *testing.T) {
	impl := readyFixture(t)
	for _, started := range []bool{false, true} {
		if started {
			if err := impl.Start(t.Context()); err != nil {
				t.Fatal(err)
			}
		}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := impl.Ready(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("Ready canceled (started=%t) = %v", started, err)
		}
		ctx, cancel = context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
		if err := impl.Ready(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Ready expired (started=%t) = %v", started, err)
		}
		cancel()
	}
}

func TestReadyAfterCanceledStart(t *testing.T) {
	impl := readyFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := impl.Start(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start = %v, want canceled", err)
	}
	assertRPCReady(t, impl, false)
}

func TestReadyAfterOccupiedPort(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	impl := readyFixture(t)
	impl.rpcServerAddr = occupied.Addr().String()
	if err := impl.Start(t.Context()); err == nil {
		t.Fatal("Start succeeded on an occupied port")
	}
	assertRPCReady(t, impl, false)
	// 绑定失败未关停健康服务，允许插件启动失败后的受控重试。
	impl.rpcServerAddr = "127.0.0.1:0"
	if err := impl.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	assertRPCReady(t, impl, true)
}

func TestReadyAfterServeFailure(t *testing.T) {
	impl := readyFixture(t)
	if err := impl.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	assertRPCReady(t, impl, true)
	assertNetworkRPCReady(t, impl)
	failed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if err := failed.Close(); err != nil {
		t.Fatal(err)
	}
	// 同一 server 的第二条监听发生真实 Accept 错误，应撤销全部健康状态并关闭 TCP。
	impl.serveWorkers.Add(1)
	impl.serve(failed)
	assertRPCReady(t, impl, false)
	connection, err := net.DialTimeout("tcp", impl.GetRPCServerAddr(), time.Second)
	if err == nil {
		connection.Close()
		t.Fatal("Serve failure left the original listener open")
	}
}

func TestReadyAfterPartialBindFailure(t *testing.T) {
	occupied, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	impl := readyFixture(t)
	impl.rpcServerAddr = occupied.LocalAddr().String()
	impl.config.DisableQUICServer = false
	if err := impl.Start(t.Context()); err == nil || !strings.Contains(err.Error(), "QUIC 地址") {
		t.Fatalf("Start = %v, want occupied QUIC port error", err)
	}
	assertRPCReady(t, impl, false)
	// 第二个协议绑定失败必须回收已经绑定的 TCP，不能留下半启动的实例。
	listener, err := net.Listen("tcp", impl.GetRPCServerAddr())
	if err != nil {
		t.Fatalf("TCP listener was not released after QUIC bind failure: %v", err)
	}
	listener.Close()
}

func TestReadyAfterStoppedServer(t *testing.T) {
	for _, beforeStart := range []bool{false, true} {
		name := "running"
		if beforeStart {
			name = "before_serve"
		}
		t.Run(name, func(t *testing.T) {
			impl := readyFixture(t)
			if beforeStart {
				impl.GetGRPCServer().Stop()
			}
			if err := impl.Start(t.Context()); err != nil {
				t.Fatal(err)
			}
			if !beforeStart {
				// 确认 Serve 已接收 RPC，独立覆盖正常停止返回 nil 的路径。
				assertNetworkRPCReady(t, impl)
				impl.GetGRPCServer().Stop()
			}
			done := make(chan struct{})
			go func() {
				impl.serveWorkers.Wait()
				close(done)
			}()
			awaitShutdownSignal(t, "Serve 已返回", done)
			assertRPCReady(t, impl, false)
		})
	}
}

func TestReadyWithdrawnBeforeCanceledStopReturns(t *testing.T) {
	impl, _, _, _, _ := shutdownFixture(t)
	assertRPCReady(t, impl, true)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := impl.Stop(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop = %v, want canceled while handler remains active", err)
	}
	assertRPCReady(t, impl, false)
	assertShutdownPending(t, "完整排空", impl.shutdownDone)
}
