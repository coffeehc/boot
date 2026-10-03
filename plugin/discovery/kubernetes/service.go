package kubernetes

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

// Config 配置单个连接的 DNS 查询和刷新；调用方在 builder 使用期间保持只读。
type Config struct {
	// RefreshInterval 是成功或失败查询后的定期刷新间隔，零值为 30 秒。
	RefreshInterval time.Duration
	// MinRefreshInterval 是主动刷新提示的最小间隔，零值为 1 秒，不得大于 RefreshInterval。
	MinRefreshInterval time.Duration
	// LookupTimeout 是每次 DNS 查询的超时，零值为 5 秒；查询可被连接关闭取消。
	LookupTimeout time.Duration
	// Resolver 是支持并发调用的 DNS 客户端，nil 使用 net.DefaultResolver。
	Resolver *net.Resolver
}

var service Service
var mutex sync.RWMutex
var name = "kubernetes_discovery"
var scope = zap.String("scope", name)

// GetService 返回已启用的发现服务；未启用时终止启动，可并发读取。
func GetService() Service {
	mutex.RLock()
	defer mutex.RUnlock()
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

// Service 装配支持 DNS 定期刷新和主动刷新的 resolver，方法可并发调用。
type Service interface {
	// GetResolverBuilder 接受最多一个只读配置；ctx 仅用于装配，连接 Close 管理运行生命周期。
	// 无配置或 nil 使用默认值；非法配置和 target 由 Build 返回错误。
	GetResolverBuilder(ctx context.Context, configs ...*Config) resolver.Builder
}

// EnablePlugin 在创建 gRPC 连接前注册全局 kubernetes scheme；重复调用无副作用。
// 必须在启动装配阶段串行调用，符合 gRPC 全局 resolver 注册约束。
func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	impl := &serviceImpl{}
	resolver.Register(impl.GetResolverBuilder(ctx))
	service = impl
	plugin.RegisterPlugin(name, impl)
}
