package base

import (
	"io/ioutil"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/web"
)

type Service interface {
	Init(configPath string, server web.HttpServer) Error
	Run() Error
	Stop() Error
	GetServiceInfo() ServiceInfo
	GetEndPoints() []EndPoint
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

type LoadingServiceInfo struct {
	ApiDeifneFile string `yaml:"apiDeifneFile"`
	apiDefine     string
	ServiceName   string   `yaml:"serviceName"`
	Version       string   `yaml:"version"`
	Descriptor    string   `yaml:"descriptor"`
	Tags          []string `yaml:"tags"`
	ServerPort    int      `yaml:"serverPort"`
	Scheme        string   `yaml:"scheme"`
	CartFile      string   `yaml:"cartFile"`
	KeyFile       string   `yaml:"keyFile"`
	DevModule     bool     `yaml:"devModule"`
}

func (this *LoadingServiceInfo) GetApiDefine() string {
	if this.apiDefine == "" {
		data, err := ioutil.ReadFile(this.ApiDeifneFile)
		if err == nil {
			this.apiDefine = string(data)
		} else {
			logger.Error("read file error :%s", err)
			this.apiDefine = "no define"
		}
	}
	return this.apiDefine
}

func (this *LoadingServiceInfo) GetServiceName() string {
	return this.ServiceName
}

func (this *LoadingServiceInfo) GetVersion() string {
	return this.Version
}

func (this *LoadingServiceInfo) GetDescriptor() string {
	if this.Descriptor == "" {
		this.Descriptor = this.ServiceName
	}
	return this.Descriptor
}

func (this *LoadingServiceInfo) GetServiceTags() []string {
	return this.Tags
}
