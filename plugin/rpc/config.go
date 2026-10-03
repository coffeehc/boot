package rpc

// RpcConfig 是 RPC 插件的启动配置，须在 EnablePlugin 前设置。
type RpcConfig struct {
	// MaxMsgSize 是单条收发消息的最大字节数，默认 8 MiB。
	MaxMsgSize int `mapstructure:"max_msg_size,omitempty" json:"max_msg_size,omitempty"`
	// MaxConcurrentStreams 是每条连接的并发流上限，默认 100。
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams,omitempty" json:"max_concurrent_streams,omitempty"`
	// RPCServerAddr 是监听地址，默认 0.0.0.0:8888；端口 0 由 Start 分配。
	RPCServerAddr string `mapstructure:"rpc_server_addr,omitempty" json:"rpc_server_addr,omitempty"`
	// DisableTCPServer 关闭 TCP 监听，默认 false；至少保留一种监听协议。
	DisableTCPServer bool `mapstructure:"disable_tcp_server,omitempty" json:"disable_tcp_server,omitempty"`
	// DisableQUICServer 关闭 QUIC 监听，默认 true，仅在显式启用时绑定 UDP。
	DisableQUICServer bool `mapstructure:"disable_quic_server,omitempty" json:"disable_quic_server,omitempty"`
}
