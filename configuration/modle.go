package configuration

type ServiceInfo struct {
	ServiceName string // `yaml:"service_name" json:"service_name"`
	Version     string // `yaml:"version" json:"version"`
	Descriptor  string // `yaml:"descriptor" json:"descriptor"`
	APIDefine   string // `yaml:"api_define" json:"api_define"`
	Scheme      string // `yaml:"scheme" json:"scheme"`
}

// 以下是本地配置，不可变更
type ServiceConfig struct {
	Model               string               // 模式
	RemoteConfigProvide *RemoteConfigProvide // 远程配置中心地址
}

type RemoteConfigProvide struct {
	Provider string
	Endpoint string
	Path     string
}
