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
  max_concurrent_streams: 100000
  max_msg_size: 4194304
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `grpc.rpc_server_addr` | RPC 服务地址 | `0.0.0.0:8888` |
| `grpc.max_concurrent_streams` | 最大并发流 | `100000` |
| `grpc.max_msg_size` | 最大消息大小（字节） | `4194304` (4MB) |

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
