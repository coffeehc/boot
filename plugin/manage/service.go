package manage

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"git.xiagaogao.com/coffee/base/utils"
	"github.com/gin-gonic/gin"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/httpx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var _plugin *serviceImpl
var mutex = new(sync.Mutex)

type serviceImpl struct {
	httpService httpx.Service
	endpoint    string
}

func (impl *serviceImpl) Start(ctx context.Context) errors.Error {
	impl.httpService.Start(nil)
	log.Debug("启动ManageServer", zap.String("endpoint", GetManageEndpoint()))
	return nil
}
func (impl *serviceImpl) Stop(ctx context.Context) errors.Error {
	err := impl.httpService.Shutdown()
	if err != nil {
		return errors.ConverError(err)
	}
	return nil
}

func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if _plugin != nil {
		return
	}
	_plugin = &serviceImpl{}
	if !viper.IsSet("manage.serverAddr") {
		viper.SetDefault("manage.serverAddr", "0.0.0.0:8889")
	}
	manageEndpoint := viper.GetString("manage.serverAddr")

	lis, err1 := net.Listen("tcp4", manageEndpoint)
	if err1 != nil {
		log.Panic("启动ManageServer失败", zap.Error(err1))
	}
	manageEndpoint = lis.Addr().String()
	err1 = lis.Close()
	if err1 != nil {
		log.Warn("管理Listen失败")
	}
	showManageEndpoint, err := utils.WarpServiceAddr(manageEndpoint)
	if err != nil {
		log.Panic("转换管理插件服务地址失败", zap.Error(err))
	}
	_plugin.endpoint = showManageEndpoint
	_plugin.httpService = httpx.NewService("manage", &httpx.Config{
		ServerAddr: manageEndpoint,
	}, log.GetLogger())
	_plugin.registerManager()
	plugin.RegisterPlugin("manager", _plugin)
}

func GetManageEndpoint() string {
	return fmt.Sprintf("http://%s", _plugin.endpoint)
}

func (impl *serviceImpl) registerManager() {
	router := impl.httpService.GetGinEngine()
	// router.Use(gin.BasicAuth(gin.Accounts{
	// 	"root": "abc###123",
	// }))
	impl.registerServiceRuntimeInfoEndpoint(router)
	impl.registerHealthEndpoint(router)
	impl.registerMetricsEndpoint(router)
	router.GET("/", func(i *gin.Context) {
		routesInfos := router.Routes()
		c := make([]string, 0)
		c = append(c, "<html><body>")
		for _, routeInfo := range routesInfos {
			c = append(c, fmt.Sprintf("<div><spen>%s</spen><a href='%s'>%s</a></div>", routeInfo.Method, routeInfo.Path, routeInfo.Path))
			// c = append(c, fmt.Sprintf("%s %s\n", routeInfo.Method,routeInfo.Path))
		}
		c = append(c, "</body></html>")
		i.Data(http.StatusOK, "text/html; charset=utf-8", []byte(strings.Join(c, "")))
		// i.String(http.StatusOK,strings.Join(c,""))
	})
}
