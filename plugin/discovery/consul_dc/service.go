package consul_dc

import (
	"context"
	"sync"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

var service Service
var mutex sync.RWMutex
var name = "consul_discovery"
var scope = zap.String("scope", name)

// GetService 返回已启用的服务，未启用时终止启动，可并发读取。
func GetService() Service {
	mutex.RLock()
	defer mutex.RUnlock()
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

// Service 装配持续发现健康实例的 resolver；方法可并发调用。
type Service interface {
	// GetResolverBuilder 接受最多一个只读原生查询配置，nil 使用默认配置。
	// ctx 仅用于装配；查询 Context 和 WaitIndex 由各连接管理，Close 取消后台查询。
	// 默认仅查询当前 run_model 标签，target 的 tag 参数可覆盖，tag= 禁用标签过滤。
	// target 中的 dc、ns、partition、peer、filter、near 覆盖原生查询配置。
	// 非法配置由 Build 返回错误，不修改调用方配置。
	GetResolverBuilder(ctx context.Context, queries ...*api.QueryOptions) resolver.Builder
}

// EnablePlugin 初始化客户端并注册 consul 和旧 console scheme，创建连接前串行调用。
// 不传客户端时使用 api.DefaultConfig，支持 Consul 官方环境变量；可注入一个非 nil 客户端覆盖 ACL/TLS 等配置。
// 客户端装配失败终止启动；重复调用不替换已注册的客户端。
func EnablePlugin(ctx context.Context, clients ...*api.Client) {
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	if len(clients) > 1 || (len(clients) == 1 && clients[0] == nil) {
		log.Panic("Consul 发现只能注入一个非 nil 客户端", scope)
	}
	var client *api.Client
	if len(clients) == 1 {
		client = clients[0]
	} else {
		var err error
		client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Panic("初始化 Consul 发现客户端失败", zap.Error(err), scope)
		}
	}
	impl := &serviceImpl{client: client}
	resolver.Register(impl.GetResolverBuilder(ctx))
	resolver.Register(&resolverBuilder{client: client, scheme: legacyScheme})
	service = impl
	plugin.RegisterPlugin(name, impl)
}
