package ipsd

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strconv"
	"sync"

	"google.golang.org/grpc/resolver"
)

// ServiceProtocolScheme 是静态地址列表使用的 gRPC target scheme。
const ServiceProtocolScheme = "ip"

// IpResolverBuilder 管理地址列表，并为每个连接创建独立 resolver；零值可用。
// Build、UpdateAddress 和连接关闭可并发执行，地址变更会广播到全部存活连接。
type IpResolverBuilder struct {
	// mutex 保护地址列表和存活 resolver 集合。
	mutex sync.Mutex
	// addresses 是由 builder 持有的地址快照，空列表表示没有可用后端。
	addresses []string
	// resolvers 保存每个连接独立的订阅，关闭后删除。
	resolvers map[*ipResolver]struct{}
}

// UpdateAddress 校验并替换全部地址；空列表撤下全部后端，非法地址保留原列表并返回错误。
// 输入必须是 host:port，端口范围为 1..65535；调用返回后可安全修改原切片。
func (impl *IpResolverBuilder) UpdateAddress(addresses []string) error {
	next := make([]string, 0, len(addresses))
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		host, port, err := net.SplitHostPort(address)
		if err != nil || host == "" {
			return fmt.Errorf("无效的 RPC 地址 %q，必须为 host:port", address)
		}
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return fmt.Errorf("RPC 地址 %q 的端口必须为 1..65535", address)
		}
		address = net.JoinHostPort(host, strconv.Itoa(portNumber))
		if _, exists := seen[address]; !exists {
			seen[address] = struct{}{}
			next = append(next, address)
		}
	}
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	if slices.Equal(impl.addresses, next) {
		return nil
	}
	impl.addresses = next
	for r := range impl.resolvers {
		select {
		case r.updates <- struct{}{}:
		default:
		}
	}
	return nil
}

// Build 为当前连接建立独立订阅；后台发布由该连接的 Close 取消并等待退出。
func (impl *IpResolverBuilder) Build(_ resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &ipResolver{
		cc: cc, builder: impl, cancel: cancel,
		updates: make(chan struct{}, 1), done: make(chan struct{}),
	}
	impl.mutex.Lock()
	if impl.resolvers == nil {
		impl.resolvers = make(map[*ipResolver]struct{})
	}
	impl.resolvers[r] = struct{}{}
	impl.mutex.Unlock()
	go r.watch(ctx)
	return r, nil
}

// Scheme 返回该 builder 支持的 target scheme。
func (impl *IpResolverBuilder) Scheme() string {
	return ServiceProtocolScheme
}

// ipResolver 串行向一个连接发布地址，不在持有 builder 锁时调用 gRPC。
type ipResolver struct {
	// cc 是当前连接的地址发布入口。
	cc resolver.ClientConn
	// builder 持有共享地址事实。
	builder *IpResolverBuilder
	// cancel 取消当前连接的后台发布。
	cancel context.CancelFunc
	// updates 合并尚未处理的地址更新和重新解析提示，不关闭此 channel。
	updates chan struct{}
	// done 在后台发布退出后关闭。
	done chan struct{}
}

// ResolveNow 非阻塞请求重新发布当前地址，可并发调用。
func (impl *ipResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case impl.updates <- struct{}{}:
	default:
	}
}

// Close 等待当前订阅退出并删除订阅，不影响其他连接。
func (impl *ipResolver) Close() {
	impl.cancel()
	<-impl.done
	impl.builder.mutex.Lock()
	delete(impl.builder.resolvers, impl)
	impl.builder.mutex.Unlock()
}

// watch 发布初始状态及后续地址快照；更新提示保证不会漏掉最新变更。
func (impl *ipResolver) watch(ctx context.Context) {
	defer close(impl.done)
	var previous []string
	published := false
	for ctx.Err() == nil {
		impl.builder.mutex.Lock()
		addresses := slices.Clone(impl.builder.addresses)
		impl.builder.mutex.Unlock()
		if !published || !slices.Equal(previous, addresses) {
			state := resolver.State{}
			for _, address := range addresses {
				addr := resolver.Address{Addr: address}
				state.Addresses = append(state.Addresses, addr)
				state.Endpoints = append(state.Endpoints, resolver.Endpoint{Addresses: []resolver.Address{addr}})
			}
			// 地址只会由 builder 更新；不重发相同的失败状态，避免重解析紧循环。
			_ = impl.cc.UpdateState(state)
			previous, published = addresses, true
		}
		select {
		case <-ctx.Done():
			return
		case <-impl.updates:
		}
	}
}
