# 测试

项目要求 Go 1.26 或更新版本。在仓库根目录运行：

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

默认测试不需要外部 gRPC 服务、内网地址或 Consul agent，也不需要 ACL 凭据。

- `testutils/TestGrpc` 在 `127.0.0.1` 临时端口启动 gRPC 服务，使用显式明文测试凭据，
  验证普通、空字符串和中文请求的响应。每次 RPC 有 5 秒超时，测试结束关闭连接和服务。
- `plugin/register/consul_rc/TestService` 使用临时本地 HTTP 服务模拟 Consul API，
  通过真实 Consul Go 客户端验证登记内容、健康检查配置、原始 metadata 不被修改、
  停止时注销和重复停止。HTTP 503 和请求超时必须返回错误，登记结果未知时仍须清理。
  测试隔离 Consul ACL token 环境变量，所有网络请求都有有限超时。

Consul 测试验证客户端协议和注册生命周期，不代表真实 Consul 集群、ACL/TLS 配置或
健康检查执行已经通过集成验证。测试不会连接或修改任何现有 Consul 环境。
