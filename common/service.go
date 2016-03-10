package common

type Service interface {
	Run() error
	Stop() error
	GetServiceInfo() ServiceInfo
	GetEndPoints() []EndPoint
}

type RpcScheme string

var (
	RpcScheme_Http  = RpcScheme("http")
	RpcScheme_https = RpcScheme("https")
)

type ServiceInfo interface {
	//获取 Api 定义的内容
	GetApiDefine() string
	//获取 Service 名称
	GetServiceName() string
	//获取服务版本号
	GetVersion() string
	//获取服务描述
	GetDescriptor() string
	//获取 Service tags
	GetServiceTags() []string
	//获取指定的服务器端口
	GetServerPort() int
	//获取 RPC 协议方式()
	GetScheme() RpcScheme
	//如果是 Https则实现该接口
	GetTLSCert() (cartFile, keyFiler string)
}
