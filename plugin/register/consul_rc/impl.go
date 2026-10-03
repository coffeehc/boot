package consul_rc

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/grpcserver"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin/rpc"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// serviceImpl 拥有本插件登记的实例集合，串行化登记与注销，避免停止后再次登记。
type serviceImpl struct {
	// client 是并发安全的 Consul 客户端，不关闭调用方共享的 HTTP 传输。
	client *api.Client
	// config 是初始化时取得的私有只读配置快照。
	config Config
	// protocol 是已装配的 RPC 传输安全协议，空值表示明文。
	protocol string
	// mutex 保护登记请求、实例集合和停止状态。
	mutex sync.Mutex
	// registered 保存已尝试登记的实例 ID，失败或未知结果也在 Stop 时清理。
	registered map[string]struct{}
	// stopped 禁止停止后再次发起登记，Start 开始一轮生命周期时复位。
	stopped bool
}

// newService 校验配置并装配原生客户端，不进行网络访问或修改调用方配置。
func newService(ctx context.Context, provided *Config) (*serviceImpl, error) {
	config := Config{RequestTimeout: 5 * time.Second}
	if provided != nil {
		config = *provided
		config.Tags = slices.Clone(provided.Tags)
		if provided.Check != nil {
			check := *provided.Check
			config.Check = &check
		}
		if config.RequestTimeout == 0 {
			config.RequestTimeout = 5 * time.Second
		}
	}
	if config.RequestTimeout <= 0 {
		return nil, fmt.Errorf("Consul 注册请求超时必须为正数")
	}
	client := config.Client
	if client == nil {
		var err error
		client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return nil, err
		}
	}
	protocol := ""
	if creds := grpcserver.GetServerCerts(ctx); creds != nil {
		protocol = creds.Info().SecurityProtocol
	}
	return &serviceImpl{client: client, config: config, protocol: protocol, registered: make(map[string]struct{})}, nil
}

// Start 标记本轮允许登记，实际登记由 register 插件在 RPC 启动后执行。
func (impl *serviceImpl) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	impl.stopped = false
	return nil
}

// Stop 注销本插件登记的实例，失败保留 ID 供重复 Stop 重试；不关闭共享客户端。
func (impl *serviceImpl) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, impl.config.RequestTimeout)
	defer cancel()
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	impl.stopped = true
	var stopErr error
	for id := range impl.registered {
		if err := impl.client.Agent().ServiceDeregisterOpts(id, (&api.QueryOptions{}).WithContext(ctx)); err != nil {
			stopErr = errors.Join(stopErr, fmt.Errorf("注销 Consul 实例 %s 失败: %w", id, err))
		} else {
			delete(impl.registered, id)
		}
	}
	return stopErr
}

// CheckDeregister 删除健康检查；旧接口不接收 context，因此使用有界请求并报告错误。
func (impl *serviceImpl) CheckDeregister(checkId string) {
	ctx, cancel := context.WithTimeout(context.Background(), impl.config.RequestTimeout)
	defer cancel()
	if err := impl.client.Agent().CheckDeregisterOpts(checkId, (&api.QueryOptions{}).WithContext(ctx)); err != nil {
		log.Error("注销 Consul 健康检查失败", zap.String("check_id", checkId), zap.Error(err))
	}
}

// Register 登记实际可访问的 RPC 地址，保留 metadata 所有权，并记录待注销实例。
func (impl *serviceImpl) Register(ctx context.Context, serviceInfo configuration.ServiceInfo) error {
	ctx, cancel := context.WithTimeout(ctx, impl.config.RequestTimeout)
	defer cancel()
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if impl.stopped {
		return fmt.Errorf("Consul 注册服务已经停止")
	}
	if serviceInfo.ServiceName == "" {
		return fmt.Errorf("登记 Consul 服务必须提供服务名")
	}
	address := impl.config.ServiceAddress
	if address == "" {
		address = rpc.GetService().GetRPCServerAddr()
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("Consul 广播地址必须为 host:port: %w", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("Consul 广播地址端口必须为 1..65535；端口 0 须在 RPC 启动后登记")
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		if impl.config.ServiceAddress != "" {
			return fmt.Errorf("显式 Consul 广播地址不能使用通配 IP")
		}
		localIP, err := rpc.GetLocalIP()
		if err != nil {
			return err
		}
		host = localIP.String()
	}
	address = net.JoinHostPort(host, strconv.Itoa(portNumber))
	id := impl.config.ServiceID
	if id == "" {
		id = serviceInfo.ServiceName + "_" + address
	}
	metadata := maps.Clone(serviceInfo.Metadata)
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["Version"] = serviceInfo.Version
	metadata["Descriptor"] = serviceInfo.Descriptor
	metadata["APIDefine"] = serviceInfo.APIDefine
	metadata["Address"] = address
	tags := impl.config.Tags
	if tags == nil {
		tags = []string{}
		if runModel := configuration.GetRunModel(); runModel != "" {
			tags = append(tags, runModel)
		}
	}
	check := api.AgentServiceCheck{}
	if impl.config.Check != nil {
		check = *impl.config.Check
	} else {
		if impl.protocol != "" && impl.protocol != "tls" {
			return fmt.Errorf("Consul 默认 gRPC 健康检查不支持 %s，请提供自定义 Check", impl.protocol)
		}
		check.GRPC = address
		check.GRPCUseTLS = impl.protocol == "tls"
		check.Interval = "10s"
		check.Timeout = "2s"
		check.DeregisterCriticalServiceAfter = viper.GetString("register.deregisterCriticalServiceAfter")
		if check.DeregisterCriticalServiceAfter == "" {
			check.DeregisterCriticalServiceAfter = "1m"
		}
	}
	if check.CheckID == "" {
		check.CheckID = id + "_grpcHealth"
	}
	if check.Name == "" {
		check.Name = check.CheckID
	}
	registration := &api.AgentServiceRegistration{ID: id, Name: serviceInfo.ServiceName, Tags: tags, Address: host, Port: portNumber, Meta: metadata, Check: &check}
	// HTTP 失败时登记结果可能未知，Stop 仍需注销这个 ID。
	impl.registered[id] = struct{}{}
	opts := api.ServiceRegisterOpts{ReplaceExistingChecks: true}.WithContext(ctx)
	if err := impl.client.Agent().ServiceRegisterOpts(registration, opts); err != nil {
		return fmt.Errorf("登记 Consul 实例 %s 失败: %w", id, err)
	}
	return nil
}
