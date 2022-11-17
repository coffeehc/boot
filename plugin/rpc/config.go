package rpc

type RpcConfig struct {
	MaxMsgSize           int    `mapstructure:"max_msg_size,omitempty" json:"max_msg_size,omitempty"`
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams,omitempty" json:"max_concurrent_streams,omitempty"`
	RPCServerAddr        string `mapstructure:"rpc_server_addr,omitempty" json:"rpc_server_addr,omitempty"`
	DisableTCPServer     bool   `mapstructure:"disable_tcp_server,omitempty" json:"disable_tcp_server,omitempty"`
	DisableQUICServer    bool   `mapstructure:"disable_quic_server,omitempty" json:"disable_quic_server,omitempty"`
}
