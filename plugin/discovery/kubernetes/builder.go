package kubernetes

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strconv"
	"time"

	"google.golang.org/grpc/resolver"
)

// ServiceProtocolScheme 是定期刷新 DNS 的 Kubernetes target scheme。
const ServiceProtocolScheme = "kubernetes"

// resolverBuilder 为每个连接独立装配 DNS 刷新器，不保留装配 context。
type resolverBuilder struct {
	// configs 为可选的只读配置，最多一个；Build 校验并取得运行快照。
	configs []*Config
}

// Build 校验 target 和配置，启动由当前连接 Close 管理的 DNS 刷新任务。
func (impl *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	if target.URL.Host != "" {
		return nil, fmt.Errorf("Kubernetes target 必须使用 kubernetes:///host:port")
	}
	host, port, err := net.SplitHostPort(target.Endpoint())
	if err != nil || host == "" {
		return nil, fmt.Errorf("Kubernetes target 必须包含有效的 host:port")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("Kubernetes target 的端口必须为 1..65535")
	}
	if len(impl.configs) > 1 {
		return nil, fmt.Errorf("DNS resolver 最多接受一个配置")
	}
	config := Config{RefreshInterval: 30 * time.Second, MinRefreshInterval: time.Second, LookupTimeout: 5 * time.Second, Resolver: net.DefaultResolver}
	if len(impl.configs) == 1 && impl.configs[0] != nil {
		provided := impl.configs[0]
		if provided.RefreshInterval != 0 {
			config.RefreshInterval = provided.RefreshInterval
		}
		if provided.MinRefreshInterval != 0 {
			config.MinRefreshInterval = provided.MinRefreshInterval
		}
		if provided.LookupTimeout != 0 {
			config.LookupTimeout = provided.LookupTimeout
		}
		if provided.Resolver != nil {
			config.Resolver = provided.Resolver
		}
	}
	if config.RefreshInterval <= 0 || config.MinRefreshInterval <= 0 || config.LookupTimeout <= 0 || config.MinRefreshInterval > config.RefreshInterval {
		return nil, fmt.Errorf("DNS 刷新间隔和查询超时必须为正数，最小刷新间隔不得超过定期刷新间隔")
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &kubernetesResolver{
		cc: cc, cancel: cancel, host: host, port: strconv.Itoa(portNumber), config: config,
		refresh: make(chan struct{}, 1), done: make(chan struct{}),
	}
	go r.watch(ctx)
	return r, nil
}

// Scheme 返回当前 builder 使用的 target scheme。
func (impl *resolverBuilder) Scheme() string {
	return ServiceProtocolScheme
}

// kubernetesResolver 串行查询 DNS，合并并限流 ResolveNow；Close 等待后台退出。
type kubernetesResolver struct {
	// cc 是当前连接的地址发布入口。
	cc resolver.ClientConn
	// cancel 取消后台任务和正在执行的 DNS 查询。
	cancel context.CancelFunc
	// host 是经过校验的服务域名或 IP。
	host string
	// port 是经过校验的十进制端口。
	port string
	// config 是本连接的只读运行配置快照。
	config Config
	// refresh 合并主动刷新提示，不关闭此 channel。
	refresh chan struct{}
	// done 在后台退出后关闭。
	done chan struct{}
}

// ResolveNow 请求主动刷新，实际频率受 MinRefreshInterval 限制；可并发调用。
func (impl *kubernetesResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case impl.refresh <- struct{}{}:
	default:
	}
}

// Close 取消查询并等待后台退出，可重复调用。
func (impl *kubernetesResolver) Close() {
	impl.cancel()
	<-impl.done
}

// watch 同时处理定期刷新和主动刷新；错误通知 gRPC，并保留最后一次成功的地址。
func (impl *kubernetesResolver) watch(ctx context.Context) {
	defer close(impl.done)
	var previous []string
	published := false
	backoff := impl.config.MinRefreshInterval
	for ctx.Err() == nil {
		started := time.Now()
		refreshDelay := impl.config.RefreshInterval
		earliestRefresh := started.Add(impl.config.MinRefreshInterval)
		queryCtx, cancel := context.WithTimeout(ctx, impl.config.LookupTimeout)
		addresses, err := impl.config.Resolver.LookupHost(queryCtx, impl.host)
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err == nil && len(addresses) == 0 {
			err = fmt.Errorf("DNS 没有返回服务 %s 的地址", impl.host)
		}
		if err != nil {
			impl.cc.ReportError(err)
			published = false
		} else {
			seen := make(map[string]struct{}, len(addresses))
			unique := make([]string, 0, len(addresses))
			for _, address := range addresses {
				if _, exists := seen[address]; !exists {
					seen[address] = struct{}{}
					unique = append(unique, address)
				}
			}
			addresses = unique
			if !published || !slices.Equal(previous, addresses) {
				state := resolver.State{}
				for _, address := range addresses {
					addr := resolver.Address{Addr: net.JoinHostPort(address, impl.port)}
					state.Addresses = append(state.Addresses, addr)
					state.Endpoints = append(state.Endpoints, resolver.Endpoint{Addresses: []resolver.Address{addr}})
				}
				if err = impl.cc.UpdateState(state); err != nil {
					impl.cc.ReportError(err)
					published = false
				} else {
					previous, published = addresses, true
				}
			}
		}
		if err != nil {
			refreshDelay = min(impl.config.RefreshInterval, backoff)
			earliestRefresh = time.Now().Add(refreshDelay)
			backoff = min(impl.config.RefreshInterval, backoff*2)
		} else {
			backoff = impl.config.MinRefreshInterval
		}
		timer := time.NewTimer(refreshDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		case <-impl.refresh:
			timer.Stop()
			// 连续错误或重复提示不能形成紧循环。
			if wait := time.Until(earliestRefresh); wait > 0 {
				timer.Reset(wait)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		}
	}
}
