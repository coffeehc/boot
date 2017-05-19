package base

import (
	"context"

	"github.com/coffeehc/httpx"
)

// Service 接口定义
type Service interface {
	Init(cxt context.Context, configPath string, server httpx.Server) Error
	Run() Error
	Stop() Error
	GetServiceDiscoveryRegister() (ServiceDiscoveryRegister, Error)
}

// ServiceInfo 接口定义
type ServiceInfo interface {
	//获取 Api 定义的内容
	GetAPIDefine() string
	//获取 Service 名称
	GetServiceName() string
	//获取服务版本号
	GetVersion() string
	//获取服务描述
	GetDescriptor() string
	//获取 Service tags
	GetServiceTag() string
	// GetScheme service 使用的协议
	GetScheme() string
}

//SimpleServiceInfo 简单的 ServiceInfo 配置
type SimpleServiceInfo struct {
	ServiceName string `yaml:"service_name" json:"service_name"`
	Version     string `yaml:"version" json:"version"`
	Descriptor  string `yaml:"descriptor" json:"descriptor"`
	APIDefine   string `yaml:"api_define" json:"api_define"`
	Tag         string `yaml:"tag" json:"tag"`
	Scheme      string `yaml:"scheme" json:"scheme"`
}

//GetAPIDefine implement ServiceInfo interface
func (ss *SimpleServiceInfo) GetAPIDefine() string {
	return ss.APIDefine
}

//GetServiceName implement ServiceInfo interface
func (ss *SimpleServiceInfo) GetServiceName() string {
	return ss.ServiceName
}

//GetVersion implement ServiceInfo interface
func (ss *SimpleServiceInfo) GetVersion() string {
	return ss.Version
}

//GetDescriptor implement ServiceInfo interface
func (ss *SimpleServiceInfo) GetDescriptor() string {
	return ss.Descriptor
}

//GetServiceTag implement ServiceInfo interface
func (ss *SimpleServiceInfo) GetServiceTag() string {
	return ss.Tag
}

//GetScheme implement ServiceInfo interface
func (ss *SimpleServiceInfo) GetScheme() string {
	return ss.Scheme
}

//NewSimpleServiceInfo create a simple ServiceInfo
func NewSimpleServiceInfo(serviceName, version, tag, scheme, descriptor, apiDefine string) ServiceInfo {
	return &SimpleServiceInfo{
		ServiceName: serviceName,
		Version:     version,
		Descriptor:  descriptor,
		APIDefine:   apiDefine,
		Tag:         tag,
		Scheme:      scheme,
	}
}
