package common

import (
	"github.com/coffeehc/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

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
	//是否开发模式
	GetDevModule() bool
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

func LoadServiceInfoConfig(configFile string) (ServiceInfo, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	info := new(LoadingServiceInfo)
	err = yaml.Unmarshal(data, info)
	if err != nil {
		return nil, err
	}
	return info, nil
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

func (this *LoadingServiceInfo) GetServerPort() int {
	return this.ServerPort
}

func (this *LoadingServiceInfo) GetScheme() RpcScheme {
	switch strings.ToUpper(this.Scheme) {
	case "HTTP":
		return RpcScheme_Http
	case "HTTPS":
		return RpcScheme_https
	default:
		return RpcScheme_Http
	}
}

func (this *LoadingServiceInfo) GetTLSCert() (cartFile, keyFiler string) {
	return this.CartFile, this.KeyFile
}

func (this *LoadingServiceInfo) GetDevModule() bool {
	return this.DevModule
}
