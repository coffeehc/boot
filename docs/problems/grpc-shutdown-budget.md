# gRPC 停机预算失效记录

## 状态与范围

- 发现时间：2026-10-12 UTC。
- 复现环境：Go 1.27.1、`google.golang.org/grpc v1.84.0`。
- Boot 基线：`master` 提交 `91fdda2`。
- 修复范围：`plugin/rpc` 的关闭编排、对应回归测试和文档；不修改 gRPC 库、
  Boot 的 Service 接口或其他框架功能。
- 当前处理：确保 RPC 插件的 `Stop(ctx)` 在预算结束后返回明确的未排空错误，
  后台继续尝试关闭。无法强行终止不响应取消的 Go handler 或不可中断 I/O。

## 现象与影响

客户端已经断开，服务端 handler 也确认收到了 RPC context 取消，但继续等待合成 I/O。
此时 `GracefulStop` 会在持有 gRPC server 内部互斥锁的情况下等待 handler 退出。
随后同步调用 `Server.Stop` 也会等待同一把锁，因此“优雅关闭超时后同步强制关闭”
不能保证停机预算。

Boot 原先在 shutdown context 到期后依次同步调用 `server.Stop()` 和
`serveWorkers.Wait()`。任意一个等待不返回，都可能拖住整个插件停机链，阻止后续
插件清理及进程关闭。即使 gRPC 的优雅关闭已返回，监听协程的等待也必须受预算限制。

## 最小复现与证据

复现只使用本地临时 TCP 端口、标准 gRPC health RPC 和合成 handler，无真实业务数据：

1. handler 进入后等待 `ctx.Done()`，确认收到取消，再等待手动释放的 channel。
2. 客户端关闭连接，确认 RPC 已失败且 handler 已收到取消。
3. 启动 `GracefulStop`，确认它在等待 handler。
4. 并发调用 `Server.Stop`，在接下来的 250 ms 内不释放 handler。
5. 记录超时和 goroutine 栈。仅为测试清理释放 handler，分别等待两种关闭和 Serve 结束。

原始独立复现的输出摘要：

```text
grpc_version=v1.84.0
client_result=rpc error: code = Canceled desc = grpc: the client connection is closing
client_disconnected=true handler_finished=false
forced_stop_exceeded_250ms=true elapsed=250.592591ms handler_finished=false
after_handler_release_both_stops_returned=true total_force_elapsed=257.253671ms reproduced=true
```

同一窗口的关键栈：

```text
强制 Stop:
  sync.(*Mutex).Lock
  google.golang.org/grpc.(*Server).stop(..., 0x0)  server.go:1692

GracefulStop:
  sync.(*WaitGroup).Wait
  google.golang.org/grpc.(*Server).stop(..., 0x1)  server.go:1722

handler:
  blockedHealth.Check 等待 release channel；RPC context 取消已确认送达
```

gRPC v1.84.0 的 `Server.stop` 在最终 `s.mu.Lock()` 后延迟解锁；
`handlersWG.Wait()` 发生在解锁前。上述栈和源码一致。强制关闭直到测试主动释放
handler 才返回，不能把背景 goroutine 留存描述成请求已被终止。

仓库内的可重复复现兼回归：

```sh
go test ./plugin/rpc -run TestStopBoundsDisconnectedHandler -v -count=1
go test -race ./plugin/rpc -count=10
```

测试分别覆盖 deadline 和显式取消；在预算断言窗口内从不释放 handler，并观察真实
gRPC 的 WaitGroup 与 Mutex 等待。只在测试清理时释放，再记录 handler、后台
强制关闭及全部监听协程最终退出。该真实栈断言依赖当前 gRPC 实现，升级依赖时需复核。

## 已尝试与修复方式

1. 同步 `GracefulStop` 超时后同步 `Stop`：已用上述最小复现证实仍可能卡住。
2. 仅依赖 RPC context 取消：复现确认取消已传播；不合作 handler 仍未退出，不能保证排空。
3. 把健康关闭、优雅关闭和监听等待放在同一后台流程。`Stop(ctx)` 只等待
   最终排空信号或 context，超时后只启动一次后台强制关闭，不再同步等待内部锁。
4. 超时返回包含“未排空”的错误，并保留 `context.Canceled` 或
   `context.DeadlineExceeded` 包装链。重复停止复用同一轮关闭，不积累新的关闭 goroutine。
5. 仅在优雅关闭与监听等待均真实完成后关闭排空信号、记录排空日志；强制关闭返回
   使用独立完成信号，不能冒充 handler 已退出。

## 验证结果

2026-10-12 UTC，在上述 Go / gRPC 版本上完成：

- `CGO_ENABLED=0 GOMAXPROCS=2 go test -p 1 ./...` 通过。
- `CGO_ENABLED=0 GOMAXPROCS=2 go vet -p 1 ./...` 通过。
- `CGO_ENABLED=0 GOMAXPROCS=2 go build -p 1 ./...` 通过。
- `CGO_ENABLED=1 GOMAXPROCS=2 go test -race -p 1 ./plugin/rpc -count=10` 通过。
- 首次定向运行中，断连 handler 的 deadline 和显式取消两种场景均在预算结束后
  约 32 µs 返回；测试允许 250 ms 调度余量，并确认此时 handler 与强制关闭均未退出。
  该测量是测试环境实测，不构成实时调度保证。
- 测试清理后分别确认 handler、客户端 RPC 协程、后台强制关闭、完整排空信号已结束。
  全仓 race 未在本轮执行；race 重复覆盖本次修改的 RPC 包。

## 下一步与边界

- 业务 handler 必须及时响应取消；不可中断的文件操作等需要业务层保证有界或采用进程隔离。
- 插件预算返回后，调用方应保留并报告停机错误，不应宣称全部请求和后台资源已经终止。
- gRPC 库未在本次修改；未来升级版本时重跑真实竞争回归并检查内部等待行为。
- 回归额外覆盖正常请求优雅完成、仍连接请求收到强制取消，以及独立的监听协程等待超时。
