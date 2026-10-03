package consul_dc

import (
	"context"
	"fmt"
	"maps"
	"math/rand/v2"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

// ServiceProtocolScheme 是 Consul 服务发现使用的 target scheme。
const ServiceProtocolScheme = "consul"

// legacyScheme 保留已公开的 console target 地址兼容性。
const legacyScheme = "console"

// resolverBuilder 持有只读客户端和查询配置，为每个连接创建独立查询循环。
type resolverBuilder struct {
	// client 是通过插件装配的并发安全 Consul 客户端。
	client *api.Client
	// queries 是最多一个的只读查询配置，Build 取得独立快照。
	queries []*api.QueryOptions
	// scheme 是 consul 或历史 console 别名。
	scheme string
}

// Build 校验服务名和查询选项，启动由当前连接 Close 管理的阻塞查询。
func (impl *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	serviceName := target.Endpoint()
	if target.URL.Host != "" || serviceName == "" || strings.Contains(serviceName, "/") {
		return nil, fmt.Errorf("Consul target 必须使用 consul:///service-name")
	}
	if len(impl.queries) > 1 {
		return nil, fmt.Errorf("Consul resolver 最多接受一个查询配置")
	}
	query := api.QueryOptions{}
	if len(impl.queries) == 1 && impl.queries[0] != nil {
		query = *impl.queries[0]
		query.NodeMeta = maps.Clone(query.NodeMeta)
	}
	if query.WaitTime == 0 {
		query.WaitTime = time.Minute
	}
	if query.WaitTime < 0 || query.WaitTime > 10*time.Minute || query.WaitHash != "" {
		return nil, fmt.Errorf("Consul 服务查询 WaitTime 必须在 0..10 分钟内，且不支持 WaitHash")
	}
	params, err := url.ParseQuery(target.URL.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("Consul target 查询参数无效: %w", err)
	}
	tags := []string{}
	if runModel := configuration.GetRunModel(); runModel != "" {
		tags = append(tags, runModel)
	}
	for key, values := range params {
		if key != "tag" && len(values) != 1 {
			return nil, fmt.Errorf("Consul 查询参数 %s 只能指定一次", key)
		}
		switch key {
		case "tag":
			tags = nil
			for _, value := range values {
				if value != "" {
					tags = append(tags, value)
				}
			}
		case "dc":
			query.Datacenter = values[0]
		case "ns":
			query.Namespace = values[0]
		case "partition":
			query.Partition = values[0]
		case "peer":
			query.Peer = values[0]
		case "filter":
			query.Filter = values[0]
		case "near":
			query.Near = values[0]
		default:
			return nil, fmt.Errorf("不支持的 Consul 查询参数 %s", key)
		}
	}
	query.WaitIndex = 0
	ctx, cancel := context.WithCancel(context.Background())
	r := &consulResolver{cc: cc, client: impl.client, cancel: cancel, serviceName: serviceName, tags: tags, query: query, done: make(chan struct{})}
	go r.watch(ctx)
	return r, nil
}

// Scheme 返回当前 builder 注册的 target scheme。
func (impl *resolverBuilder) Scheme() string {
	return impl.scheme
}

// consulResolver 通过阻塞查询跟踪健康后端；只有 watch 修改查询索引和已发布地址。
type consulResolver struct {
	// cc 是当前连接的地址发布入口。
	cc resolver.ClientConn
	// client 是并发安全的共享客户端，不由 resolver 关闭。
	client *api.Client
	// cancel 取消后台查询和退避等待。
	cancel context.CancelFunc
	// serviceName 是经过校验的 Consul 服务名。
	serviceName string
	// tags 是当前连接的标签过滤，空列表不限制标签。
	tags []string
	// query 是本连接私有的查询选项，Context 和 WaitIndex 由 watch 管理。
	query api.QueryOptions
	// done 在后台退出后关闭。
	done chan struct{}
}

// ResolveNow 无需额外触发；阻塞查询持续接收地址变化，错误后自动退避重试。
func (impl *consulResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close 取消进行中的 HTTP 查询并等待退出，可重复调用。
func (impl *consulResolver) Close() {
	impl.cancel()
	<-impl.done
}

// watch 处理索引回退、零索引、快速变化限流和错误退避；网络错误保留已有地址。
func (impl *consulResolver) watch(ctx context.Context) {
	defer close(impl.done)
	var previous []string
	published := false
	backoff := time.Second
	tokens := 2.0
	refilled := time.Now()
	for ctx.Err() == nil {
		// 每秒补一个令牌，允许两次连续响应；正常长轮询不增加等待。
		now := time.Now()
		tokens = min(2, tokens+now.Sub(refilled).Seconds())
		refilled = now
		if tokens < 1 {
			timer := time.NewTimer(time.Duration((1 - tokens) * float64(time.Second)))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			tokens = min(2, tokens+time.Since(refilled).Seconds())
			refilled = time.Now()
		}
		tokens--
		// Consul 最多增加 wait/16 抖动，再预留 15 秒网络余量，避免失联服务器永久占住查询。
		queryCtx, cancel := context.WithTimeout(ctx, impl.query.WaitTime+impl.query.WaitTime/16+15*time.Second)
		services, meta, err := impl.client.Health().ServiceMultipleTags(impl.serviceName, impl.tags, true, impl.query.WithContext(queryCtx))
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err == nil && meta == nil {
			err = fmt.Errorf("Consul 查询没有返回索引信息")
		}
		if err != nil {
			impl.cc.ReportError(err)
			published = false
			// 指数退避上限 30 秒，附加最多 20% 抖动，避免同时重试。
			timer := time.NewTimer(backoff + time.Duration(rand.Int64N(int64(backoff/5)+1)))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			backoff = min(30*time.Second, backoff*2)
			continue
		}
		backoff = time.Second
		if meta.LastIndex < impl.query.WaitIndex {
			impl.query.WaitIndex = 0
		} else {
			impl.query.WaitIndex = max(1, meta.LastIndex)
		}
		addresses := make([]string, 0, len(services))
		seen := make(map[string]struct{}, len(services))
		for _, entry := range services {
			if entry == nil || entry.Service == nil {
				log.Warn("忽略缺少服务信息的 Consul 实例", zap.String("service", impl.serviceName))
				continue
			}
			host := entry.Service.Address
			if host == "" && entry.Node != nil {
				host = entry.Node.Address
			}
			if host == "" || entry.Service.Port < 1 || entry.Service.Port > 65535 {
				log.Warn("忽略地址或端口无效的 Consul 实例", zap.String("service", impl.serviceName), zap.String("id", entry.Service.ID))
				continue
			}
			address := net.JoinHostPort(host, strconv.Itoa(entry.Service.Port))
			if _, exists := seen[address]; !exists {
				seen[address] = struct{}{}
				addresses = append(addresses, address)
			}
		}
		if !published || !slices.Equal(previous, addresses) {
			state := resolver.State{}
			for _, address := range addresses {
				addr := resolver.Address{Addr: address}
				state.Addresses = append(state.Addresses, addr)
				state.Endpoints = append(state.Endpoints, resolver.Endpoint{Addresses: []resolver.Address{addr}})
			}
			// watch 会继续处理未来变更；空地址也必须发布以撤下旧后端。
			_ = impl.cc.UpdateState(state)
			previous, published = addresses, true
		}
	}
}
