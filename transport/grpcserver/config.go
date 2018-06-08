package grpcserver

//Config grpcboot config
type GRPCConfig struct {
	MaxMsgSize           int    `yaml:"max_msg_size"`
	MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams"`
}

func (config *GRPCConfig) initGRPCConfig() {
	if config.MaxConcurrentStreams == 0 {
		config.MaxConcurrentStreams = 100000
	}
	if config.MaxMsgSize == 0 {
		config.MaxMsgSize = 1024 * 1024 * 4
	}
}
