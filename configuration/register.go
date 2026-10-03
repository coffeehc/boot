package configuration

import (
	"context"

	"google.golang.org/grpc"
)

// ServiceRegisterInfo 描述对外发布的服务信息，读取方不得原地修改 metadata。
type ServiceRegisterInfo struct {
	// Info 是服务身份和接口描述。
	Info ServiceInfo `json:"info"`
	// ServiceAddr 是外部可访问的服务地址，发布时必须有效。
	ServiceAddr string `json:"service_addr"`
	// ManageEndpoint 是可选管理入口。
	ManageEndpoint string `json:"manage_endpoint"`
	// Metadata 是可选服务属性，由发布方持有。
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RPCService 装配应用拥有的 RPC 客户端，初始化与停止由应用生命周期串行管理。
// discovery 只在初始化失败时关闭连接；初始化成功后由实现保存并在停止时 Close。
type RPCService interface {
	// GetRPCServiceInfo 返回服务标识和逻辑 target，调用方不得修改共享 metadata。
	GetRPCServiceInfo() ServiceInfo
	// InitRPCService 初始化客户端；ctx 仅用于本次初始化，不能保存作长期 RPC context。
	// 返回错误时自行清理其他部分初始化资源；grpcConn 由 discovery 关闭。
	// 返回 nil 时取得连接所有权，必须支持 RPC 并发调用并在停止时关闭连接。
	InitRPCService(ctx context.Context, grpcConn *grpc.ClientConn) error
}
