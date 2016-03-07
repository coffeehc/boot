package common

type Service interface {
	Run() error
	Stop() error
	GetServiceInfo() ServiceInfo
	GetEndPoints() []EndPoint
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
