package configuration

type ServiceInfo struct {
	ServiceName string            `mapstructure:"service_name,omitempty" json:"service_name,omitempty"`
	Version     string            `mapstructure:"version,omitempty" json:"version,omitempty"`
	Descriptor  string            `mapstructure:"descriptor,omitempty" json:"descriptor,omitempty"`
	APIDefine   string            `mapstructure:"api_define,omitempty" json:"api_define,omitempty"`
	Metadata    map[string]string `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`
	TargetUrl   string            `mapstructure:"target_url,omitempty" json:"target_url,omitempty"`
}

// 以下是本地配置，不可变更

type RemoteConfigProvide struct {
	Provider string `mapstructure:"provider,omitempty" json:"provider,omitempty"`
	Endpoint string `mapstructure:"endpoint,omitempty" json:"endpoint,omitempty"`
	Path     string `mapstructure:"path,omitempty" json:"path,omitempty"`
}
