package base

import (
	"net"

	"github.com/coffeehc/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

var (
	localIp = net.IPv4(127, 0, 0, 1)
)

func GetLocalIp() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error("无法获取网络接口信息,%s", err)
		return localIp
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return localIp
}

func GetDefaultConfigPath(configPath string) string {
	if configPath == "" {
		configPath = path.Join(GetAppDir(), "config.yml")
		_, err := os.Open(configPath)
		if err != nil {
			logger.Error("%s 不存在", configPath)
			dir, err := os.Getwd()
			if err != nil {
				logger.Error("获取不到工作目录")
				return ""
			}
			configPath = path.Join(dir, "config.yml")
			_, err = os.Open(configPath)
			if err != nil {
				logger.Error("%s 不存在", configPath)
				return ""
			}
		}
	}
	return configPath
}

func LoadConfig(configPath string, config interface{}) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, config)
}

func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return path
}

/*
	获取App执行文件目录
*/
func GetAppDir() string {
	return filepath.Dir(GetAppPath())
}
