# gRPC 就绪状态与监听退出不一致

## 问题

基线 `ad3e34222f235286b9e7b380160a1cc73573dedf` 的 RPC 插件已有同一个
`health.Server`，构造时为 `NOT_SERVING`，全部监听绑定成功后改为 `SERVING`。
未启动、context 取消导致启动失败，以及 TCP/QUIC 绑定失败时的状态原本正确。
业务缺少一个不建立连接的公开就绪读取入口，同时原有生命周期有两处遗漏：

1. `serve` 只在非 `grpc.ErrServerStopped` 的错误时撤销健康状态。
   gRPC 正常停止后 `Serve` 返回 nil，或者停止早于 `Serve` 时返回
   `grpc.ErrServerStopped`，这两种路径都可能留下 `SERVING`。
2. `Stop` 把 `healthServer.Shutdown()` 放在后台 `gracefulStop` 中。
   已取消的 shutdown context 可使 `Stop` 先返回，此时后台协程尚未执行，
   就绪检查仍可能成功。停机超时不能被解释成实例仍可接收流量。

## 复现与回归

所有测试只使用合成 handler 和 loopback 临时端口，无外部服务。

- `TestReadyAfterStoppedServer/running`：先完成一次真实 gRPC health RPC，
  再停止 server，等 `Serve` 完全返回后检查健康状态。
- `TestReadyAfterStoppedServer/before_serve`：先停止 server，再进入 `Serve`，
  等监听协程结束后检查健康状态，覆盖 `grpc.ErrServerStopped`。
- `TestReadyWithdrawnBeforeCanceledStopReturns`：让 handler 保持未结束，用已取消
  的 context 调用 `Stop`，在其返回后立即检查健康状态，最后才释放 handler。

验证时保留新 `Ready` 读取方法和测试，临时恢复基线的 `serve`、`Stop` 和
`gracefulStop` 行为。在 `GOMAXPROCS=1` 下运行以下回归，三个分支均复现
`Ready = <nil>, want ready false`；恢复最小修复后通过：

```sh
GOMAXPROCS=1 go test -p 1 -count=1 ./plugin/rpc \
  -run 'TestReadyAfterStoppedServer|TestReadyWithdrawnBeforeCanceledStopReturns'
```

另有测试覆盖未启动、取消和到期、启动成功的真实 health RPC、TCP 占用、
TCP 成功而 QUIC 占用时释放已绑定 TCP、真实 `Accept` 失败后停止所有监听，
以及正常停止。保留原有 handler 取消、在途请求排空和停机预算回归。

## 最小修复

- `Service.Ready(context.Context) error` 直接调用本实例现有 `health.Server.Check`，
  只在 `SERVING` 时成功；预先检查 context 取消或到期，不等待启动、不探测网络。
- 任一 `Serve` 返回时都调用同一个 `healthServer.Shutdown()`；异常路径仍按原逻辑
  停止整个 gRPC server。`Shutdown` 会阻止后续状态写入重新发布 `SERVING`。
- 在 `Stop` 的 `shutdownOnce` 中同步撤销健康状态，再调度原有后台排空流程。
  关闭预算、强制关闭和最终排空的所有权及语义保持不变。

就绪状态只表示该实例的 RPC 监听状态，不替代业务依赖检查。业务仍在插件启动前
注册服务，生命周期仍由插件管理。没有新增 server、连接探针或生命周期管理者。
