package discovery

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/grpcclient"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin/discovery/ipsd"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

// RPCServiceInitializationByResolverBuilder 使用连接级 resolver 初始化服务。
// 默认回调超时为 5 秒，调用方提供的 deadline 优先；成功不代表服务已连接。
// 失败或 panic 关闭连接，成功后由 RPCService 保存连接并在 Stop 中 Close。
func RPCServiceInitializationByResolverBuilder(ctx context.Context, rpcService configuration.RPCService, resolverBuilders ...resolver.Builder) error {
	conn, err := grpcclient.NewClientConnByResolverBuilder(ctx, rpcService.GetRPCServiceInfo(), resolverBuilders...)
	if err != nil {
		return err
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}
	return initializeRPCService(ctx, rpcService, conn)
}

// RPCServiceInitializationByAddresses 使用静态多地址初始化服务，并返回可更新地址的 builder。
// 输入为空时保留空地址列表，等待 UpdateAddress；非法地址返回错误。
// 非 ip scheme 的 TargetUrl 由服务名构造的 ip target 替换，保证实际使用传入地址。
// 连接失败清理和成功后的所有权约定与 RPCServiceInitialization 相同。
func RPCServiceInitializationByAddresses(ctx context.Context, rpcService configuration.RPCService, serverAddr ...string) (ipsd.ResolverBuilder, error) {
	builder, err := ipsd.GetResolverBuilder(ctx, serverAddr...)
	if err != nil {
		return nil, err
	}
	info := rpcService.GetRPCServiceInfo()
	if !strings.HasPrefix(info.TargetUrl, ipsd.ServiceProtocolScheme+":") {
		name := info.ServiceName
		if name == "" {
			name = "addresses"
		}
		info.TargetUrl = ipsd.ServiceProtocolScheme + ":///" + url.PathEscape(name)
	}
	conn, err := grpcclient.NewClientConnByResolverBuilder(ctx, info, builder)
	if err != nil {
		return builder, err
	}
	return builder, initializeRPCService(ctx, rpcService, conn)
}

// RPCServiceInitializationByAddress 通过显式地址初始化服务，RPC 使用各自的调用 context。
// 成功后由 RPCService 持有并关闭连接；失败或 panic 由本入口清理连接。
func RPCServiceInitializationByAddress(ctx context.Context, rpcService configuration.RPCService, serverAddr string) error {
	conn, err := grpcclient.NewClientConn(ctx, serverAddr, rpcService.GetRPCServiceInfo().ServiceName)
	if err != nil {
		return err
	}
	return initializeRPCService(ctx, rpcService, conn)
}

// RPCServiceInitialization 根据 ServiceInfo.TargetUrl 初始化服务，遵循原生 scheme 选择规则。
// 创建连接是惰性的，不保证目标可用；ctx 仅用于装配和回调，不能保存作长期 RPC context。
// 成功后由 RPCService 保存连接并在停止时 Close；失败或 panic 由本入口关闭连接。
func RPCServiceInitialization(ctx context.Context, rpcService configuration.RPCService) error {
	conn, err := grpcclient.NewClientConnByServiceInfo(ctx, rpcService.GetRPCServiceInfo())
	if err != nil {
		return err
	}
	return initializeRPCService(ctx, rpcService, conn)
}

// initializeRPCService 在回调成功前持有连接，错误、取消和 panic 路径都释放连接。
func initializeRPCService(ctx context.Context, rpcService configuration.RPCService, conn *grpc.ClientConn) (initErr error) {
	transferred := false
	defer func() {
		if !transferred {
			if closeErr := conn.Close(); closeErr != nil {
				initErr = errors.Join(initErr, closeErr)
			}
		}
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := rpcService.InitRPCService(ctx, conn); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	log.Debug("初始化 RPC 客户端成功", zap.String("service", rpcService.GetRPCServiceInfo().ServiceName))
	transferred = true
	return nil
}
