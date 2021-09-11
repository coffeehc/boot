#boot

一直想把这个框架开源，今天日子比较特别，正好也没事，整理一下代码，把一些陈年烂代码该删的都删了，然后从私有库转到github

Boot的核心服务启动器，最早的时候是作为微服务框架而设计的，但随着不同项目的应用和框架设计思想的影响下，不断不断的重构成为今天这样子。

首先这个框架比较简陋，以至于啥都没有，就只管启动服务，这有点类似IOC,但是又不太一样，因为boot把服务看成插件，每个插件提供自己独特的服务，很多插件组装起来对外就是一个独立的微服务。

当然，既然是微服务，有一些东西必须是标准化的，比如传输协议，服务发现，监控等

######关于服务协议

boot使用的是GRPC作为服务协议，当然，您也可以换别的协议，毕竟，RPC也只是一个插件而已。

######关于服务注册与发现

在plugin的两个包里面可以看到有discovery和register两个包，分别是发现和注册的实现

第一版使用了Etcd来支撑服务注册和服务发现，

第二版改为consul作为服务中心，同时兼容etcd作为服务发现中心，其实里面问题挺多的，只是没时间去改，好比熔断等，直到最近整体要迁移到kubernetes中，所以做了第三版改进

第三版为了适应k8s+service mash，所以回归了最原始的dns方式，由service Mash来控制服务的访问，熔断等，其实这一版我是最满意的一版，因为结构最简单了。
（ps:在做架构的时候我习惯把流程设计得复杂一些，但是实现要求一定要简单，不要绕，毕竟开发人员在代码实现与沟通过程中还那么绕的话，代码就别想写好。）
其实serviceMash就是要消灭代码中的服务发现与注册，从系统层面来保证服务的动态变更与调用，正好我也就顺其自然的没有使用服务注册，但是服务发现还有的，不过也只是对dns的简单封装，当然如果有谁有兴趣优化一下k8s的服务发现组件，欢迎提交rp


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
  plugin.RegisterPluginByFast(name,nil,nil)
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
