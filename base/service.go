package base

type Service interface {
	Run() Error
	Stop() Error
	GetServiceInfo() ServiceInfo
	GetServiceDiscoveryRegister() ServiceDiscoveryRegister
}

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
}
