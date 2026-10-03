package consul_rc

import (
	"context"
	"sync"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/plugin"
	"github.com/coffeehc/boot/plugin/register/internal"
	"github.com/coffeehc/boot/plugin/rpc"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// Config 配置 Consul 登记的连接、广播地址和健康检查；EnablePlugin 后不可修改。
type Config struct {
	// Client 是可共享的原生客户端，nil 使用 api.DefaultConfig，包括官方 ACL/TLS 环境变量。
	Client *api.Client
	// ServiceAddress 是 Consul 可访问的 host:port，空值使用实际 RPC 监听地址并转换通配 IP。
	ServiceAddress string
	// ServiceID 是唯一实例标识，空值使用服务名和广播地址（含端口）构造。
	ServiceID string
	// Tags 是登记标签；nil 使用当前 run_model，非 nil 空列表禁用标签。
	Tags []string
	// Check 是只读原生健康检查，nil 默认 gRPC 检查；ALTS/mTLS 可用自定义检查。
	Check *api.AgentServiceCheck
	// RequestTimeout 是登记、注销请求的超时，零值为 5 秒；调用方更短的 deadline 仍生效。
	RequestTimeout time.Duration
}

var service Service
var mutex sync.RWMutex
var name = "consul_registercenter"
var scope = zap.String("scope", name)

// GetService 返回已经启用的注册服务，未启用时终止启动，可并发读取。
func GetService() Service {
	mutex.RLock()
	defer mutex.RUnlock()
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

// Service 登记 Consul 实例，并在插件 Stop 时注销已登记的服务；方法串行化保护状态。
type Service interface {
	internal.RegisterCenter
	// CheckDeregister 删除指定健康检查；保留原签名，请求受 RequestTimeout 限制，失败写错误日志。
	CheckDeregister(checkId string)
}

// EnablePlugin 在启动装配阶段创建客户端并设置注册中心，不执行登记请求。
// 可提供最多一个只读配置；错误终止启动，重复调用不替换配置。
// rpc 和本插件先于 register 插件装配，保证登记时已监听、注销时 RPC 仍可用。
func EnablePlugin(ctx context.Context, configs ...*Config) {
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	if len(configs) > 1 {
		log.Panic("Consul 注册只能设置一个配置", scope)
	}
	var config *Config
	if len(configs) == 1 {
		config = configs[0]
	}
	impl, err := newService(ctx, config)
	if err != nil {
		log.Panic("初始化 Consul 注册客户端失败", zap.Error(err), scope)
	}
	rpc.EnablePlugin(ctx)
	internal.EnablePlugin(ctx)
	if err := internal.GetService().SetRegisterCenter(impl); err != nil {
		log.Panic("添加注册中心失败", zap.Error(err), scope)
	}
	service = impl
	plugin.RegisterPlugin(name, impl)
}
