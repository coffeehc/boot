# Boot 框架使用说明

## 项目简介

Boot 是一个 Go 微服务核心启动框架，采用插件化架构设计。将服务视为插件，通过插件组装构建独立的微服务。

### 核心特性

- **插件化架构**：类似 IOC，每个服务作为一个插件运行
- **gRPC 协议**：默认使用 gRPC 作为服务通信协议，支持 QUIC 传输
- **服务发现**：适配 Kubernetes + Service Mesh，使用 DNS 方式
- **进程管理**：支持守护进程模式，自动 PID 管理
- **配置管理**：基于 Viper 的配置系统

## 技术栈

- Go 1.22+
- gRPC 1.71.0
- Consul（可选）
- Prometheus（监控）
- Zap（日志）
- Fiber 2.52.5

## 项目结构

```
boot/
├── component/          # 组件
│   └── grpcx/         # GRPC 相关组件
│       ├── grpcserver/   # 服务端
│       ├── grpcclient/   # 客户端
│       ├── grpcquic/     # QUIC 支持
│       └── grpcrecovery/ # 恢复和日志
├── configuration/     # 配置管理
│   ├── keys.go
│   ├── modle.go        # 数据模型
│   └── serviceconfig.go
├── engine/            # 启动引擎
│   ├── init.go         # 初始化
│   ├── start.go        # 启动服务
│   ├── stop.go         # 停止服务
│   ├── kill.go         # 强制终止
│   ├── daemon.go       # 守护进程
│   └── setup.go        # 设置
├── plugin/            # 插件系统
│   ├── plugins.go      # 插件核心
│   ├── manage/         # 管理插件
│   │   ├── health.go   # 健康检查
│   │   ├── metrics/    # 指标监控
│   │   └── service.go  # 管理服务
│   ├── rpc/            # RPC 插件
│   ├── discovery/      # 服务发现
│   │   ├── kubernetes/ # K8s 发现
│   │   ├── consul_dc/  # Consul 发现
│   │   └── ipsd/       # IPSD 发现
│   └── register/       # 服务注册
│       └── consul_rc/  # Consul 注册
└── testutils/         # 测试工具
```

## 快速开始

### 1. 创建服务主文件

```go
package main

import (
	"context"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/engine"
	"github.com/coffeehc/boot/plugin/rpc"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()
	
	// 定义服务信息
	serviceInfo := configuration.ServiceInfo{
		ServiceName: "my-service",
		Version:     "1.0.0",
		Descriptor:  "我的微服务",
	}
	
	// 启动引擎
	engine.StartEngine(ctx, serviceInfo, func(ctx context.Context, cmd *cobra.Command, args []string) (func(), error) {
		// 初始化你的插件
		myPlugin.EnablePlugin(ctx)
		
		// 返回关闭回调
		return func() {
			// 清理资源
		}, nil
	})
}
```

### 2. 创建自定义插件

```go
package myplugin

import (
	"context"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"go.uber.org/zap"
	"sync"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "myPlugin"
var scope = zap.String("scope", name)

// 定义服务接口
type Service interface {
	DoSomething() error
}

// GetService 获取服务实例（单例）
func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

// EnablePlugin 启用插件
func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	service = newService(ctx)
	plugin.RegisterPlugin(name, service)
}

// newService 创建服务实例
func newService(ctx context.Context) Service {
	// 依赖其他插件（如 RPC）
	rpc.EnablePlugin(ctx)
	rpcService := rpc.GetService()
	
	impl := &serviceImpl{
		rpcService: rpcService,
	}
	return impl
}

type serviceImpl struct {
	rpcService rpc.Service
}

func (s *serviceImpl) DoSomething() error {
	// 实现你的业务逻辑
	return nil
}
```

## 核心概念

### 1. 插件系统

每个插件需要实现 `plugin.Plugin` 接口：

```go
type Plugin interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

#### 运行根与停机

每次执行内建 `start` 命令时，Boot 在调用 `ServiceStart` 和插件初始化前创建一个运行根
context。`ServiceStart`、插件初始化调用方传递的 context，以及每个插件的 `Start` 都共享同一
取消根。业务代码可调用 `engine.GetRootContext()` 获取当前根，并从它派生任务 context；不要
取消根本身，取消权属于 Boot。该方法只在一次 `start` 执行期间有效，未初始化或命令退出后
返回 nil；重叠的并发启动会返回错误，连续的独立启动会创建新的根。

外部运行 context 的取消以及 SIGINT / SIGTERM 都会取消运行根。插件启动失败时，Boot 先取消
运行根，再用保留 context 值、脱离取消的 30 秒回滚 context 逆序停止已启动插件。正常停机也先
取消运行根，再创建同样独立的 30 秒 shutdown context。正常停机时，无参数业务关闭回调先执行，
之后插件以该 context 逆序停止；业务回调无法接收 shutdown context，因此必须自行及时返回。插件的 `Stop(ctx)` 应
响应取消或 deadline；忽略 context 或卡在其他不可中断调用中的回调无法被 Boot 强行结束。

已有 `engine.StartEngine`、`engine.ServiceStart`、`plugin.StartPlugins` 和无返回值的
`plugin.StopPlugins(ctx)` 调用方式保持兼容。需要检查清理失败的直接调用方可改用
`plugin.StopPluginsWithError(ctx)`；engine 会将其错误返回给命令并记录。若业务服务需要访问同一
运行生命周期，可在启动期间使用 `engine.GetRootContext()`，并在服务自行启动的后台任务中遵守
context 取消。

`ServiceStart` 返回非 nil 关闭回调即交付一次业务清理责任，即使同时返回错误也会执行该回调。
插件启动返回错误、取消或 panic 时，先完成插件回滚，再执行业务关闭回调；该回调不能假定全部插件
仍然可用。业务关闭回调的 panic 会被转换为命令错误，并继续插件清理，原始启动和清理错误都保留。

或者通过 `plugin.RegisterPlugin()` 注册，框架会自动包装。

#### 启动失败与资源回滚

插件按注册顺序串行启动。`StartPlugins(ctx)` 返回首个启动错误后，不再启动后续插件；
框架仅对 `Start` 已返回 nil 的实例执行逆序 `Stop`。全部插件启动后的
`AfterPluginStartedHandler` 返回错误时也回滚成功集合。插件 `Start` 或该回调 panic 时，转换为包含
阶段信息的错误，插件 Start 的异常同时携带插件名称；panic 值为 error 时，其原始原因可通过
`errors.Is` 检查。

回滚和正常关闭复用内部停止流程：提取成功集合后立即清空，重复 `StopPlugins` 不会
重复关闭同一批实例。任一 Stop 返回错误或 panic 仍继续清理其余插件；回滚错误带插件名记录日志，
并通过 `errors.Join` 附在原始启动错误后，两者均可通过 `errors.Is` 检查。

回滚保留启动 context 的值，但通过 `context.WithoutCancel` 脱离其取消和截止时间，
再设置整轮 30 秒清理预算。该预算需要插件 Stop 遵守 context，不强制中断插件。
正常关闭由 engine 使用独立 30 秒 shutdown context，并沿用原有逆序；`StopPlugins` 保持无返回值，
新增 `StopPluginsWithError` 供需要返回错误的调用方使用。插件应遵守该 context；期限不能强制结束
忽略 context 的 Stop。

启动错误由 Cobra 命令返回到 engine，再沿用 `os.Exit(-1)` 非零退出约定；公开
`StartEngine` 签名不变。生命周期入口由 engine 串行调用，不支持重叠启动/停止。
返回错误或 panic 的失败插件仍负责自己部分初始化的资源，Boot 仅回滚已经成功启动的实例。

### 2. 服务发现

框架提供三种服务发现方式：

#### Kubernetes（推荐）

```go
import (
	"github.com/coffeehc/boot/plugin/discovery/kubernetes"
)

kubernetes.EnablePlugin(ctx)
```

#### Consul

```go
import (
	"github.com/coffeehc/boot/plugin/discovery/consul_dc"
)

consul_dc.EnablePlugin(ctx)
```

#### DNS（简单）

```go
import (
	"github.com/coffeehc/boot/plugin/discovery/ipsd"
)

ipsd.EnablePlugin(ctx)
```

### 3. RPC 服务

```go
import (
	"github.com/coffeehc/boot/plugin/rpc"
)

// 启用 RPC 插件
rpc.EnablePlugin(ctx)

// 获取 gRPC Server
rpcService := rpc.GetService()
server := rpcService.GetGRPCServer()

// 注册你的 gRPC 服务
pb.RegisterYourServiceServer(server, &yourServiceImpl{})
```

## 配置

配置使用 YAML 格式（默认）：

```yaml
grpc:
  rpc_server_addr: "0.0.0.0:8888"
  max_concurrent_streams: 100
  max_msg_size: 8388608
  max_connection_idle: 30m
  disable_tcp_server: false
  disable_quic_server: true
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `grpc.rpc_server_addr` | RPC 服务地址 | `0.0.0.0:8888` |
| `grpc.max_concurrent_streams` | 每条 HTTP/2 连接的并发流上限 | `100` |
| `grpc.max_msg_size` | 单条收发消息的最大编码字节数 | `8388608` (8 MiB) |
| `grpc.max_connection_idle` | 没有活动 RPC 的连接保留时长；`0` 不限制 | `30m` |
| `grpc.disable_tcp_server` | 关闭 TCP 监听；至少保留一种协议 | `false` |
| `grpc.disable_quic_server` | 关闭 QUIC 监听 | `true` |

### gRPC server/client 默认行为

`component/grpcx` 是 Boot 自维护的 gRPC 封装，当前依赖 gRPC `v1.84.0` 和 Protobuf `v1.36.12`。`grpc/examples` 固定在 gRPC `v1.84.0` 的同一提交，只供原有示例测试使用。

| 配置 | 默认行为 |
|------|----------|
| Client 负载均衡 | `round_robin`，显式注册策略，允许 resolver 提供服务配置 |
| Client 业务重试 | 不提供全方法重试策略；保留 gRPC 自带透明重试，幂等方法可自行配置重试策略 |
| Client 等待就绪 | 使用 gRPC 默认 fail-fast；需要等待时按调用传入 `grpc.WaitForReady(true)` |
| Client 消息大小 | 收发均为 8 MiB，可通过 CallOption 覆盖 |
| Unary deadline | 调用方未设置时默认一分钟；显式 deadline 原样保留 |
| Streaming deadline | 由调用方设置或取消，不自动限制长连接时长 |
| Client keepalive | 有活动 RPC 时每 60 秒检查，20 秒超时；无活动 RPC 不发送 PING |
| Server keepalive | 保留 gRPC 的 2 小时检查、20 秒超时；允许的客户端最小 PING 间隔为一分钟，不允许空闲 PING |
| 窗口、连接退避、读写 buffer、client idle、代理 | 使用 gRPC 默认值，启用动态流控；不强制静态大窗口或禁用代理 |
| 压缩 | 注册 gzip 支持，默认不压缩；需要时按调用传入 `grpc.UseCompressor("gzip")` |
| 连接凭据 | Client 必须显式配置 TLS、ALTS 或 `insecure.NewCredentials()`；Server 未设置凭据时使用明文 TCP |

`grpcserver.NewServer(ctx, nil)` 读取 `grpc` 配置段；传入非 nil `GRPCServerConfig` 时使用显式配置，不修改该对象。`grpc.MaxConnectionIdle` 旧配置键仍支持，新的 `grpc.max_connection_idle` 优先。配置解析失败由 `NewServer` 返回错误。

插件只在 `Start` 绑定端口，保留配置中的监听 IP；端口为 `0` 时，`Start` 成功后 `GetRPCServerAddr()` 返回实际端口。默认仅启动 TCP。正常 `Stop` 先结束健康报告并调用 `GracefulStop`；shutdown context 到期后调用 `Stop` 关闭连接并返回取消错误。业务 handler 仍需响应 RPC context 取消。

### 客户端创建与定制

```go
ctx = grpcclient.SetClientCerds(ctx, credentials.NewTLS(&tls.Config{
    MinVersion: tls.VersionTLS12,
}))
conn, err := grpcclient.NewClientConn(ctx, "dns:///api.example.com:443", "apiService")
if err != nil {
    return err
}
defer conn.Close()

callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
// 使用 callCtx 调用生成的 gRPC client。
```

构造连接使用 `grpc.NewClient`，成功仅表示目标和选项被接受，实际连接在第一次 RPC 时建立。构造时的 context 用于凭据装配，不负责连接生命周期；连接由调用方关闭，每次 RPC 使用自己的 context。

需要调整 keepalive、消息上限或按方法配置重试时，可使用 `grpcclient.BuildDialOption(ctx, serviceName)`，追加原生 `grpc.DialOption` 后调用 `grpc.NewClient`。Server 同样可使用 `grpcserver.BuildGRPCServerOptions` 追加原生选项。并发流上限按连接计算；长流或高并发服务需要根据实际容量设置，不能替代整个进程的并发控制。

### 升级影响

- 默认并发流上限从 100000 收紧到每条连接 100；已有显式配置不变，长流较多的服务需配置适合自身容量的值。
- RPC 插件默认消息上限从 4 MiB 调整为 8 MiB；client 发送上限从 2 MiB 调整为 8 MiB。
- 旧 client 每 10 秒发送空闲 PING，新 server 会拒绝该模式；应先升级 client，再升级 server，或在迁移期间显式覆盖 server enforcement policy。
- 默认取消全方法重试、等待就绪和强制 gzip；依赖这些行为的调用方需按具体方法显式配置。
- 标准 gRPC 状态码、取消和 deadline 错误原样保留；Boot `base/errors` 仍使用已有的自定义状态码 18 编解码，stream 接收错误也会解码，正常 EOF 原样返回。
- unary 和 streaming 都保留调用方的 outgoing metadata，并向 handler 传递追踪和鉴权 context。
- 代理 codec 保持原有接口和协议名，解码后的 payload 拥有独立 buffer；同时支持新旧 protobuf 消息接口。
- 自维护 ALTS 同步畸形短帧校验，Clone 保留服务身份；认证中心连接使用 gRPC 的默认连接管理。
- 同一服务创建多个 client/server 时复用已注册的指标 collector，避免后续连接的指标不可见。

## 命令行使用

构建完成后，服务支持以下命令：

```bash
# 启动服务（前台）
./your-service start

# 启动守护进程
./your-service daemonStart

# 重启服务
./your-service restart

# 停止服务
./your-service stop

# 查看版本
./your-service version

# 设置初始化
./your-service setup
```

## 依赖管理

框架已处理插件调用依赖，不会重复创建插件。通过单例模式保证唯一性。

### 插件依赖示例

```go
func newService(ctx context.Context) Service {
	// 先启动依赖的插件
	dbPlugin.EnablePlugin(ctx)
	cachePlugin.EnablePlugin(ctx)
	
	// 依赖的服务会自动初始化
	impl := &serviceImpl{
		db:    dbPlugin.GetService(),
		cache: cachePlugin.GetService(),
	}
	return impl
}
```

## 构建项目

使用提供的构建脚本：

```bash
./build.sh
```

或在你的项目中：

```bash
go build -o your-service main.go
```

## 注意事项

1. **插件名称唯一**：确保每个插件名称唯一，避免重复注册
2. **单例保证**：插件应使用单例模式，参考模板代码
3. **资源清理**：在 Stop 方法中正确释放资源
4. **错误处理**：框架会捕获 panic，但应合理处理错误
5. **版本控制**：设置正确的服务版本号

## 高级功能

### 自定义配置变更监听

```go
configuration.RegisterOnConfigChange(func() {
	// 配置变更时的处理
})
```

### 自定义插件启动钩子

```go
plugin.AfterPluginStartedHandler = func() error {
	// 所有插件启动后的逻辑
	return nil
}
```

### 健康检查

```go
import (
	"github.com/coffeehc/boot/plugin/manage"
)

manage.EnablePlugin(ctx)
```

### 指标监控

```go
import (
	"github.com/coffeehc/boot/plugin/manage/metrics"
)

metrics.EnablePlugin(ctx)
```

## 架构演进

- **第一版**：Etcd 作为服务注册与发现中心
- **第二版**：Consul 作为服务中心，兼容 Etcd 发现
- **第三版**：回归 DNS 方式，适配 K8s + Service Mesh，由系统层面保证服务调用

## 未来规划

项目致力于将框架精简为百来行代码，服务更加细粒度化，真正成为应用组装的"最后一公里"。

## 贡献

欢迎提交 Issue 和 Pull Request 优化和改进框架。
