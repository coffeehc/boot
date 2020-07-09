package rpc

type RpcConfig struct {
	MaxMsgSize           int    // `yaml:"max_msg_size"`
	MaxConcurrentStreams uint32 // `yaml:"max_concurrent_streams"`
	RPCServerAddr        string
}
