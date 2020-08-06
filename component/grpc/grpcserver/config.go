package grpcserver

// Config grpcboot config
type GRPCServerConfig struct {
	MaxMsgSize           int    `mapstructure:"max_msg_size,omitempty" json:"max_msg_size,omitempty"`
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams,omitempty" json:"max_concurrent_streams,omitempty"`
}
