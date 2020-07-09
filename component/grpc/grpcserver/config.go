package grpcserver

// Config grpcboot config
type GRPCServerConfig struct {
	MaxMsgSize           int    // `yaml:"max_msg_size"`
	MaxConcurrentStreams uint32 // `yaml:"max_concurrent_streams"`
}
