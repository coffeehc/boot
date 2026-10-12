package rpc

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// shutdownHandler 在测试主动释放前永不结束，用于区分预算返回和真实排空。
type shutdownHandler struct {
	grpc_health_v1.UnimplementedHealthServer
	// entered 确认真实 RPC 已进入 handler。
	entered chan struct{}
	// canceled 确认连接取消已经传播，不代表 handler 已退出。
	canceled chan struct{}
	// release 仅由测试在断言后主动释放。
	release chan struct{}
	// finished 记录 handler 的实际返回。
	finished chan struct{}
}

func (handler *shutdownHandler) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	defer close(handler.finished)
	close(handler.entered)
	select {
	case <-ctx.Done():
		close(handler.canceled)
		// 模拟收到取消后仍在不可中断 I/O 中的业务 handler。
		<-handler.release
	case <-handler.release:
	}
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

// shutdownFixture 仅使用临时 loopback 端口，不读取凭据或连接任何外部服务。
func shutdownFixture(t *testing.T) (*serviceImpl, *shutdownHandler, *grpc.ClientConn, <-chan error, func()) {
	t.Helper()
	handler := &shutdownHandler{
		entered: make(chan struct{}), canceled: make(chan struct{}),
		release: make(chan struct{}), finished: make(chan struct{}),
	}
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(handler.release) }) }
	impl := &serviceImpl{
		server: grpc.NewServer(), healthServer: health.NewServer(),
		config: &RpcConfig{DisableQUICServer: true}, rpcServerAddr: "127.0.0.1:0",
	}
	grpc_health_v1.RegisterHealthServer(impl.server, handler)
	if err := impl.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	var connection *grpc.ClientConn
	t.Cleanup(func() {
		// 所有停机预算断言完成后才释放 handler，不能靠提前释放掩盖卡死。
		release()
		if connection != nil {
			_ = connection.Close()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := impl.Stop(ctx); err != nil {
			t.Errorf("cleanup Stop: %v", err)
		}
		awaitShutdownSignal(t, "在途请求及监听协程最终排空", impl.shutdownDone)
		if impl.forceStopDone != nil {
			awaitShutdownSignal(t, "后台强制关闭最终返回", impl.forceStopDone)
		}
	})
	var err error
	connection, err = grpc.NewClient("passthrough:///"+impl.GetRPCServerAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithNoProxy())
	if err != nil {
		t.Fatal(err)
	}
	callCtx, cancelCall := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancelCall)
	callDone := make(chan error, 1)
	clientFinished := make(chan struct{})
	go func() {
		defer close(clientFinished)
		_, err := grpc_health_v1.NewHealthClient(connection).Check(callCtx, &grpc_health_v1.HealthCheckRequest{})
		callDone <- err
	}()
	awaitShutdownSignal(t, "handler 开始", handler.entered)
	t.Cleanup(func() {
		release()
		awaitShutdownSignal(t, "handler 最终退出", handler.finished)
		awaitShutdownSignal(t, "客户端 RPC 协程最终退出", clientFinished)
	})
	return impl, handler, connection, callDone, release
}

func awaitShutdownSignal(t *testing.T, label string, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
		t.Log(label)
	case <-time.After(5 * time.Second):
		t.Errorf("%s 未在测试清理预算内完成", label)
	}
}

func assertShutdownPending(t *testing.T, label string, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
		t.Fatalf("%s 提前报告完成", label)
	default:
	}
}

// awaitGRPCShutdownStack 确认本次真实 gRPC 关闭已进入指定等待，避免依赖固定 sleep。
func awaitGRPCShutdownStack(t *testing.T, function, wait string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		buffer := make([]byte, 1<<20)
		stacks := string(buffer[:runtime.Stack(buffer, true)])
		for _, stack := range strings.Split(stacks, "\n\n") {
			if strings.Contains(stack, "(*serviceImpl)."+function) &&
				strings.Contains(stack, "google.golang.org/grpc.(*Server).stop") && strings.Contains(stack, wait) {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("未观察到 %s 的 %s 等待:\n%s", function, wait, stacks)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestStopBoundsDisconnectedHandler(t *testing.T) {
	for _, mode := range []string{"deadline", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			impl, handler, connection, callDone, _ := shutdownFixture(t)
			if err := connection.Close(); err != nil {
				t.Fatal(err)
			}
			awaitShutdownSignal(t, "连接关闭已传播到 handler context", handler.canceled)
			select {
			case err := <-callDone:
				if err == nil {
					t.Fatal("断连 RPC 意外成功")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("客户端 RPC 未结束")
			}
			ctx, cancel := context.WithCancel(context.Background())
			wantErr := context.Canceled
			if mode == "deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 200*time.Millisecond)
				wantErr = context.DeadlineExceeded
			}
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- impl.Stop(ctx) }()
			awaitGRPCShutdownStack(t, "gracefulStop", "sync.(*WaitGroup).Wait")
			if mode == "canceled" {
				cancel()
			}
			<-ctx.Done()
			budgetExpired := time.Now()
			select {
			case err := <-result:
				if !errors.Is(err, wantErr) || !strings.Contains(err.Error(), "未排空") {
					t.Fatalf("Stop error = %v, want 未排空 and %v", err, wantErr)
				}
			case <-time.After(250 * time.Millisecond):
				t.Fatal("Stop 在预算结束后仍阻塞；handler 此时从未释放")
			}
			t.Logf("Stop 在预算结束后 %s 返回；handler 仍未释放", time.Since(budgetExpired))
			assertShutdownPending(t, "handler", handler.finished)
			assertShutdownPending(t, "完整排空", impl.shutdownDone)
			// v1.84.0 中强制 Stop 确实被 GracefulStop 的互斥锁挡住。
			awaitGRPCShutdownStack(t, "forceStop", "sync.(*Mutex).Lock")
			assertShutdownPending(t, "强制关闭", impl.forceStopDone)
			firstForceDone := impl.forceStopDone
			if err := impl.Stop(ctx); !errors.Is(err, wantErr) {
				t.Fatalf("重复 Stop error = %v, want %v", err, wantErr)
			}
			if impl.forceStopDone != firstForceDone {
				t.Fatal("重复 Stop 启动了新的后台强制关闭")
			}
		})
	}
}

func TestStopForcesConnectedHandlerCancellation(t *testing.T) {
	impl, handler, _, _, _ := shutdownFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- impl.Stop(ctx) }()
	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Stop error = %v, want deadline exceeded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Stop 未遵守停机预算")
	}
	awaitShutdownSignal(t, "强制关闭已传播取消到已连接的 handler", handler.canceled)
	assertShutdownPending(t, "handler", handler.finished)
	assertShutdownPending(t, "完整排空", impl.shutdownDone)
}

func TestStopGracefullyDrainsHandler(t *testing.T) {
	impl, handler, _, callDone, release := shutdownFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- impl.Stop(ctx) }()
	select {
	case err := <-result:
		t.Fatalf("handler 释放前 Stop 已返回: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	assertShutdownPending(t, "正常请求取消", handler.canceled)
	release()
	select {
	case err := <-callDone:
		if err != nil {
			t.Fatalf("优雅关闭截断了正常请求: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("正常 RPC 未完成")
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Stop error = %v", err)
		}
	case <-ctx.Done():
		t.Fatal("排空后 Stop 未返回")
	}
	awaitShutdownSignal(t, "正常关闭在途请求及监听协程已退出", impl.shutdownDone)
	if impl.forceStopDone != nil {
		t.Fatal("正常优雅关闭意外发起强制关闭")
	}
	cancel()
	if err := impl.Stop(ctx); err != nil {
		t.Fatalf("排空后重复 Stop error = %v", err)
	}
}

func TestStopBoundsServeWorkersWait(t *testing.T) {
	impl := &serviceImpl{server: grpc.NewServer(), healthServer: health.NewServer()}
	// 模拟监听 goroutine 尚未退出：gRPC 已结束也不能无预算等待 serveWorkers。
	impl.serveWorkers.Add(1)
	t.Cleanup(func() {
		impl.serveWorkers.Done()
		awaitShutdownSignal(t, "监听协程最终排空", impl.shutdownDone)
		awaitShutdownSignal(t, "后台强制关闭最终返回", impl.forceStopDone)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- impl.Stop(ctx) }()
	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "未排空") {
			t.Fatalf("Stop error = %v, want 未排空 and deadline exceeded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("等待监听协程导致 Stop 超出停机预算")
	}
	assertShutdownPending(t, "监听协程排空", impl.shutdownDone)
}
