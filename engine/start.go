package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var daemonModel = pflag.Bool("daemon", false, "开启守护进程模式")

func readPid(serviceInfo configuration.ServiceInfo) int {
	pidFile := fmt.Sprintf("./%s.pid", serviceInfo.ServiceName)
	pidData, _ := os.ReadFile(pidFile)
	pid, _ := strconv.ParseInt(string(pidData), 10, 64)
	return int(pid)
}

func savePid(serviceInfo configuration.ServiceInfo, pid int) {
	pidFile := fmt.Sprintf("./%s.pid", serviceInfo.ServiceName)
	os.WriteFile(pidFile, []byte(strconv.FormatInt(int64(pid), 10)), 0644)
}

func buildStartCmd(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "启动服务",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Printf("守护模式:%t\n", *daemonModel)
			log.Debug("守护模式", zap.Bool("daemonModel", *daemonModel))
			if *daemonModel {
				return daemon(serviceInfo)
			}
			ctx, cancelFunc := context.WithCancel(ctx)
			configuration.InitConfiguration(ctx, serviceInfo)
			defer plugin.StopPlugins(ctx)
			var closeCallback ServiceCloseCallback = nil
			go func() {
				_closeCallback, err := start(ctx, cmd, args)
				if err != nil {
					log.Error("启动服务失败", zap.Error(err))
					cancelFunc()
				}
				closeCallback = _closeCallback
				defer func() {
					if e := recover(); e != nil {
						err := errors.ConverUnknowError(e)
						log.DPanic("程序捕获不能处理的异常", err.GetFieldsWithCause()...)
						cancelFunc()
					}
				}()
				plugin.StartPlugins(ctx)
				log.Info("服务启动完成")
			}()
			WaitServiceStop(ctx, cancelFunc, closeCallback)
			return nil
		},
	}
}

func daemon(serviceInfo configuration.ServiceInfo) error {
	pid := readPid(serviceInfo)
	if pid != 0 {
		p, _ := os.FindProcess(pid)
		if p != nil {
			p.Signal(syscall.SIGTERM)
		}
	}
	args := make([]string, 0, len(os.Args)-1)
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--daemon") {
			continue
		}
		args = append(args, arg)
	}
	var e error
	f, e := os.Open("/dev/null")
	if e != nil {
		return e
	}
	fd := f.Fd()
	args[0], e = filepath.Abs(args[0])
	if e != nil {
		return e
	}
	pwd, _ := os.Getwd()
	pid, e = syscall.ForkExec(args[0], args, &syscall.ProcAttr{
		Env: os.Environ(),
		// Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Files: []uintptr{fd, fd, fd},
		Dir:   pwd,
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	})
	if e != nil {
		fmt.Println("错误:%#v", e)
		return e
	}
	log.Info("创建新的进程", zap.Int("pid", pid))
	savePid(serviceInfo, pid)
	os.Exit(0)
	return nil
}
