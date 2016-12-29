package consultool

import (
	"time"

	"io/ioutil"

	"github.com/coffeehc/logger"
	"gopkg.in/yaml.v2"
)

//LoadConsulConfig 加载 consul 配置
func loadConsulConfig(configPath string) *ConsulConfig {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Error("加载配置文件失败,使用默认配置")
		return &ConsulConfig{}
	}
	i := &struct {
		Consul *ConsulConfig `yaml:"consul"`
	}{}
	err = yaml.Unmarshal(data, i)
	if err != nil {
		logger.Error("解析Consul配置失败,使用默认配置,%s", err)
		return &ConsulConfig{}
	}
	return i.Consul

}

//ConsulConfig consul连接配置
type ConsulConfig struct {
	Address    string         `yaml:"address"`
	Scheme     string         `yaml:"scheme"`
	DataCenter string         `yaml:"daraCenter"`
	WaitTime   time.Duration  `yaml:"waitTime"`
	Token      string         `yaml:"token"`
	BasicAuth  *HTTPBasicAuth `yaml:"basic_auth"`
}

//HTTPBasicAuth consul的 BaseAuth
type HTTPBasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

//GetAddress 获取 consul 的连接地址 默认为127.0.0.1:8500
func (cc *ConsulConfig) GetAddress() string {
	if cc.Address == "" {
		return "127.0.0.1:8500"
	}
	return cc.Address
}

//GetScheme 获取 consul的连接协议,默认为: http
func (cc *ConsulConfig) GetScheme() string {
	if cc.Scheme == "" {
		return "http"
	}
	return cc.Scheme
}

//GetDataCenter 获取配置的数据中心,默认为: dc
func (cc *ConsulConfig) GetDataCenter() string {
	if cc.DataCenter == "" {
		return "dc"
	}
	return cc.DataCenter
}

//GetWaitTime 获取等待时间
func (cc *ConsulConfig) GetWaitTime() time.Duration {
	return cc.WaitTime
}

//GetToken 获取 Token
func (cc *ConsulConfig) GetToken() string {
	return cc.Token
}
