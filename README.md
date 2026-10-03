#boot

一直想把这个框架开源，今天日子比较特别，正好也没事，整理一下代码，把一些陈年烂代码该删的都删了，然后从私有库转到github

Boot的核心服务启动器，最早的时候是作为微服务框架而设计的，但随着不同项目的应用和框架设计思想的影响下，不断不断的重构成为今天这样子。

首先这个框架比较简陋，以至于啥都没有，就只管启动服务，这有点类似IOC,但是又不太一样，因为boot把服务看成插件，每个插件提供自己独特的服务，很多插件组装起来对外就是一个独立的微服务。

当然，既然是微服务，有一些东西必须是标准化的，比如传输协议，服务发现，监控等

######关于服务协议

boot使用的是GRPC作为服务协议，当然，您也可以换别的协议，毕竟，RPC也只是一个插件而已。

######关于服务注册与发现

`plugin/discovery` 提供 gRPC resolver 适配，`plugin/register` 负责服务登记。
普通域名、Kubernetes Service 和 Service Mesh 入口优先使用 gRPC 原生
`dns:///host:port`；Headless Service 可返回 Pod 地址，供客户端负载均衡。

额外提供三种能力：`ipsd` 静态多地址及动态更新，`kubernetes` 定期 DNS 刷新，
`consul_dc` Consul 健康实例阻塞查询。Consul 客户端可用官方环境变量配置或原生客户端注入，
新地址使用 `consul:///service-name`，历史 `console` scheme 继续支持。
Kubernetes 适配器不订阅 API 或 EndpointSlice；etcd、Nacos 等可通过外部 resolver 接入。

每个连接拥有独立 resolver，连接关闭会取消并等待后台查询退出。Consul 注册中心在插件
停止时注销登记实例。配置覆盖、初始化与连接所有权见 [USAGE.md](USAGE.md#2-服务发现)。

######关于插件化

各种通用的数据库处理，队列处理等可以实现plugin.Plugin，然后作为一个插件注入到服务中。可以套用以下模版来生成服务:
```go

import (
  "context"
  "github.com/coffeehc/base/log"
  "github.com/coffeehc/boot/plugin"
  "go.uber.org/zap"
  "sync"
)

var service Service
var mutex = new(sync.RWMutex)
var name = ""
var scope = zap.String("scope",name)


func GetService()Service  {
  if service == nil{
    log.Panic("Service没有初始化",scope)
  }
  return service
}

func EnablePlugin(ctx context.Context)  {
  if name == ""{
    log.Panic("插件名称没有初始化")
  }
  mutex.Lock()
  defer mutex.Unlock()
  if service!=nil{
    return
  }
  service = newService(ctx)
  plugin.RegisterPlugin(name,service)
}


type Service interface {

}

func newService(ctx context.Context)Service  {
  xxx.EnablePlugin(ctx)
  impl := &serviceImpl{
    xxxService: xxx.GetService()
  }
  return impl
}

type serviceImpl struct {
  xxxService xxx.Service

}

```

框架已经处理好了插件调用依赖，所以不会重复创建插件,但实际为了保证不重复创建插件是以上模版来保证的，之所为没有集成到Plugin里面去，主要是因为这里的服务都是单例模式，如果有的插件遇到不能使用单例的场景，或者更个性化的需求的时候，就可以通过修改模版创建的代码来实现自定义，这样灵活度更高。

那么基本上这个框架就这样，已经算比较稳定的一个版本，既然开源了，以后就尽量向下兼容来升级框架，代码很简单也很容易看懂，如果大家有兴趣的话，可以开issus提建议或者直接点rp丢过来。

说实话，我倒是很希望有一天boot能被精简成百来行代码，服务被更细粒度化，一切插件都被服务化，被弱化成less Service或者镜像化，或者成为低代码平台的组件等其他形态，那boot将会真正成为一个启动器，它将是应用组装的最后那一公里。

README写的很粗糙，慢慢改善
