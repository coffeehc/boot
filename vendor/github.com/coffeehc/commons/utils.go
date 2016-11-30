package commons

import (
	"encoding/base64"
	"github.com/coffeehc/logger"
	"os/exec"
	"os"
	"path/filepath"
	"crypto/rand"
	"net"
	"syscall"
	"fmt"
	"os/signal"
)

func GetRand(size int) string {
	bs := make([]byte, size)
	_, err := rand.Read(bs)
	if err != nil {
		return GetRand(size)
	}
	return base64.RawStdEncoding.EncodeToString(bs)

}

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

/*
	wait,一般是可执行函数的最后用于阻止程序退出
*/
func WaitStop() {
	var sigChan = make(chan os.Signal, 3)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("接收到指令:%s,立即关闭程序", sig)
}
