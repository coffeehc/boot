package base

import (
	"context"
	"github.com/coffeehc/web"
)

type Service interface {
	Init(configPath string, server web.HttpServer, cxt context.Context) Error
	Run() Error
	Stop() Error
	//GetServiceInfo() ServiceInfo
	GetServiceDiscoveryRegister() (ServiceDiscoveryRegister, Error)
	GetConfig() interface{}
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

	GetScheme() string
}

type SimpleServiceInfo struct {
	ServiceName string `yaml:"service_name"`
	Version     string `yaml:"version"`
	Descriptor  string `yaml:"descriptor"`
	ApiDefine   string `yaml:"api_define"`
	Tag         string `yaml:"tag"`
	Scheme      string `yaml:"scheme"`
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

func (this *SimpleServiceInfo) GetScheme() string {
	return this.Scheme
}

func NewSimpleServiceInfo(serviceName, version, tag, scheme, descriptor, apiDefine string) ServiceInfo {
	return &SimpleServiceInfo{
		ServiceName: serviceName,
		Version:     version,
		Descriptor:  descriptor,
		ApiDefine:   apiDefine,
		Tag:         tag,
		Scheme:      scheme,
	}
}
