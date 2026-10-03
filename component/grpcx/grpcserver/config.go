package grpcserver

// GRPCServerConfig 控制消息边界和每条连接的并发流数量。
// 启动前配置完成后应视为只读；零值使用 Boot 默认值。
type GRPCServerConfig struct {
	// MaxMsgSize 是单条收发消息的最大字节数；零值为 8 MiB，禁止负值。
	MaxMsgSize int `mapstructure:"max_msg_size,omitempty" json:"max_msg_size,omitempty"`
	// MaxConcurrentStreams 是每条 HTTP/2 连接的并发流上限；零值为 100。
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams,omitempty" json:"max_concurrent_streams,omitempty"`
}
