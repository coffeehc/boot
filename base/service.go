package base

import "github.com/coffeehc/web"

type Service interface {
	Init(configPath string, server web.HttpServer) Error
	Run() Error
	Stop() Error
	GetServiceInfo() ServiceInfo
	GetServiceDiscoveryRegister() (ServiceDiscoveryRegister, Error)
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
	GetServiceTag() string
}

type SimpleServiceInfo struct {
	ServiceName string
	Version     string
	Descriptor  string
	ApiDefine   string
	Tag         string
}

func (this *SimpleServiceInfo) GetApiDefine() string {
	return this.ApiDefine
}
func (this *SimpleServiceInfo) GetServiceName() string {
	return this.ServiceName
}
func (this *SimpleServiceInfo) GetVersion() string {
	return this.Version
}
func (this *SimpleServiceInfo) GetDescriptor() string {
	return this.Descriptor
}
func (this *SimpleServiceInfo) GetServiceTag() string {
	return this.Tag
}

func BuildSimpleServiceInfo(serviceName string, version string, tag string, descriptor string, apiDefine string) ServiceInfo {
	return &SimpleServiceInfo{
		ServiceName: serviceName,
		Version:     version,
		Descriptor:  descriptor,
		ApiDefine:   apiDefine,
		Tag:         tag,
	}
}
