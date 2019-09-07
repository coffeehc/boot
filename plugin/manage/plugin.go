package manage

import (
	"context"
	"fmt"
	"net"
	"sync"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/base/utils"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/httpx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var _plugin *pluginImpl
var mutex = new(sync.Mutex)

type pluginImpl struct {
	httpService httpx.Service
	endpoint    string
}

func (impl *pluginImpl) Start(ctx context.Context) errors.Error {
	impl.httpService.Start(nil)
	return nil
}
func (impl *pluginImpl) Stop(ctx context.Context) errors.Error {
	impl.httpService.Shutdown()
	return nil
}

func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if _plugin != nil {
		return
	}
	_plugin = &pluginImpl{}
	if !viper.IsSet("manage.serverAddr") {
		viper.SetDefault("manage.serverAddr", "0.0.0.0:0")
	}
	manageEndpoint := viper.GetString("manage.serverAddr")
	manageEndpoint, err := utils.WarpServiceAddr(manageEndpoint)
	if err != nil {
		log.Fatal("转换管理插件服务地址失败", zap.Error(err))
	}
	lis, err1 := net.Listen("tcp4", manageEndpoint)
	if err1 != nil {
		log.Fatal("启动ManageServer失败", zap.Error(err1))
	}
	manageEndpoint = lis.Addr().String()
	lis.Close()
	_plugin.endpoint = manageEndpoint
	log.Debug("启动ManageServer", zap.String("endpoint", GetManageEndpoint()))
	_plugin.httpService = httpx.NewService("manage", &httpx.Config{
		ServerAddr: manageEndpoint,
	}, log.GetLogger())
	_plugin.registerManager()
	plugin.RegisterPlugin("manager", _plugin)
}

func GetManageEndpoint() string {
	return fmt.Sprintf("http://%s", _plugin.endpoint)
}

func (impl *pluginImpl) registerManager() {
	router := impl.httpService.GetGinEngine()
	// router.Use(gin.BasicAuth(gin.Accounts{
	// 	"root": "abc###123",
	// }))
	impl.registerServiceRuntimeInfoEndpoint(router)
	impl.registerHealthEndpoint(router)
	impl.registerMetricsEndpoint(router)
}
