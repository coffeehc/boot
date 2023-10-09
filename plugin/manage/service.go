package manage

import (
	"context"
	"embed"
	"fmt"
	"github.com/coffeehc/boot/configuration"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/template/html/v2"
	"io/fs"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/base/utils"
	"github.com/coffeehc/boot/plugin"
	"github.com/coffeehc/httpx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

//go:embed  views
var views embed.FS

var _plugin *serviceImpl
var mutex = new(sync.Mutex)

//var WebEngine *gin.Engine

type serviceImpl struct {
	httpService httpx.Service
	endpoint    string
}

func (impl *serviceImpl) Start(_ context.Context) error {
	_plugin.registerManager()
	impl.httpService.Start(nil)
	log.Debug("启动ManageServer", zap.String("endpoint", GetManageEndpoint()))
	return nil
}
func (impl *serviceImpl) Stop(_ context.Context) error {
	err := impl.httpService.Shutdown()
	if err != nil {
		return errors.ConverError(err)
	}
	return nil
}

func EnablePlugin(_ context.Context) {
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
	for {
		lis, err1 := net.Listen("tcp4", manageEndpoint)
		if err1 != nil {
			log.Error("启动ManageServer失败,需要更换端口", zap.Error(err1))
			tcpAddr, err := net.ResolveTCPAddr("tcp", manageEndpoint)
			if err != nil {
				log.Error("启动ManageServer地址解析失败", zap.Error(err1))
				return
			}
			tcpAddr.Port = tcpAddr.Port + 1
			manageEndpoint = tcpAddr.String()
			continue
		}
		manageEndpoint = lis.Addr().String()
		err1 = lis.Close()
		if err1 != nil {
			log.Warn("管理Listen失败", zap.Error(err1))
			tcpAddr, err := net.ResolveTCPAddr("tcp", manageEndpoint)
			if err != nil {
				log.Error("启动ManageServer地址解析失败", zap.Error(err1))
				return
			}
			tcpAddr.Port = tcpAddr.Port + 1
			manageEndpoint = tcpAddr.String()
			continue
		}
		break
	}
	showManageEndpoint, err := utils.WarpServiceAddr(manageEndpoint)
	if err != nil {
		log.Panic("转换管理插件服务地址失败", zap.Error(err))
	}
	_plugin.endpoint = showManageEndpoint
	webRoot, err2 := fs.Sub(views, "views")
	if err2 != nil {
		log.Error("错误", zap.Error(err2))
		return
	}
	httpFileSystem := http.FS(webRoot)
	engine := html.NewFileSystem(httpFileSystem, ".gohtml")
	engine.Reload(false)      // Optional. Default: false
	engine.Debug(false)       // Optional. Default: false
	engine.Layout("embed")    // Optional. Default: "embed"
	engine.Delims("{{", "}}") // Optional. Default: engine delimiters
	_plugin.httpService = httpx.NewService(&httpx.Config{
		AppName:    "manage",
		ServerAddr: manageEndpoint,
		Views:      engine,
	})
	plugin.RegisterPlugin("manager", _plugin)
}

func GetManageEndpoint() string {
	return fmt.Sprintf("http://%s", _plugin.endpoint)
}

func (impl *serviceImpl) registerManager() {
	app := impl.httpService.GetEngine()
	app.Use(pprof.New())
	app.Get("/metrics", monitor.New())
	app.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendString("pong")
	})
	RegisterServiceRuntimeInfoEndpoint(app)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Format(map[string]interface{}{
			"goroutine_count": runtime.NumGoroutine(),
			"GOGC":            os.Getenv("GOGC"),
		})
	})
	app.Get("/gc/stats", func(c *fiber.Ctx) error {
		stat := &debug.GCStats{}
		debug.ReadGCStats(stat)
		data := &struct {
			debug.GCStats
			ServiceName string
			Version     string
		}{
			GCStats:     *stat,
			ServiceName: configuration.GetServiceInfo().ServiceName,
			Version:     configuration.GetServiceInfo().Version,
		}
		return c.Render("gcStats", data)
	})
	app.Get("/gc/stats/setmemlimit", func(ctx *fiber.Ctx) error {
		limit := ctx.QueryInt("limit", 0)
		if limit != 0 {
			debug.SetMemoryLimit(int64(limit))
		}
		return nil
	})
	app.Get("/gc/stats/setgogc", func(ctx *fiber.Ctx) error {
		gogc := ctx.QueryInt("gogc", 0)
		if gogc != 0 {
			debug.SetGCPercent(gogc)
		}
		return nil
	})
	app.Get("/shutdown", func(c *fiber.Ctx) error {
		if c.Query("key", "") != "coffee" {
			return nil
		}
		return syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})
	app.Get("/", func(c *fiber.Ctx) error {
		routesInfos := app.GetRoutes()
		//c := make([]string, 0)
		//c = append(c, "<html><body>")
		//for _, routeInfo := range routesInfos {
		//	c = append(c, fmt.Sprintf("<div><spen>%s</spen><a href='%s'>%s</a></div>", routeInfo.Method, routeInfo.Path, routeInfo.Path))
		//	// c = append(c, fmt.Sprintf("%s %s\n", routeInfo.Method,routeInfo.Path))
		//}
		//c = append(c, "</body></html>")
		//ctx.Set("Content-Type", "text/html")
		//return ctx.SendString(strings.Join(c, ""))
		data := &struct {
			Routers     []fiber.Route
			ServiceName string
			Version     string
		}{
			Routers:     routesInfos,
			ServiceName: configuration.GetServiceInfo().ServiceName,
			Version:     configuration.GetServiceInfo().Version,
		}
		return c.Render("index", data)
	})
}
