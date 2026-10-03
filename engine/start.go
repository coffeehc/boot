package engine

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/coffeehc/boot/plugin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"strconv"
	"syscall"
)

func ReadPid(serviceInfo configuration.ServiceInfo) int {
	pidFile := GetPidFilePath(serviceInfo.ServiceName)
	pidData, _ := os.ReadFile(pidFile)
	pid, _ := strconv.ParseInt(string(pidData), 10, 64)
	return int(pid)
}

func buildReStartCmd(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "重启程序",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			pid := ReadPid(serviceInfo)
			//process, err := os.FindProcess(pid)
			//if err != nil {
			//	return err
			//}
			//log.Info("关闭服务", zap.Int("pid", pid))
			Kill(pid)
			//log.Info("args", zap.Strings("args", os.Args))
			appPath, err := os.Executable()
			if err != nil {
				return err
			}
			for i, arg := range os.Args {
				if arg == "daemonStart" {
					os.Args[i] = "start"
				}
			}
			_, err = Background(appPath, os.Args, "") //GetLogFile(serviceInfo.ServiceName))
			if err != nil {
				return err
			}
			return nil
		},
	}
}

const Env_DaemonMode = "__DaemonMode"

func buildDaemonStartCmd(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) *cobra.Command {
	return &cobra.Command{
		Use:   "daemonStart",
		Short: "启动守护进程",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) error {
			//log.Info("args", zap.Strings("args", os.Args))
			appPath, err := os.Executable()
			if err != nil {
				return err
			}
			for i, arg := range os.Args {
				if arg == "daemonStart" {
					os.Args[i] = "start"
				}
			}
			os.Setenv(ENVDaemonIndex, "0")
			os.Setenv(Env_DaemonMode, "true")
			_, err = Background(appPath, os.Args, GetLogFile(serviceInfo.ServiceName))
			return err
		},
	}
}

func buildStartCmd(ctx context.Context, serviceInfo configuration.ServiceInfo, start ServiceStart) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "启动服务",
		Long:  serviceInfo.Descriptor,
		RunE: func(cmd *cobra.Command, args []string) (runErr error) {
			parent := cmd.Context()
			if parent == nil {
				parent = ctx
			}
			runCtx, cancelRun, releaseRun, err := beginRunContext(parent)
			if err != nil {
				return err
			}
			defer releaseRun()
			defer func() {
				if e := recover(); e != nil {
					converted := errors.ConverUnknowError(e)
					if converted == nil {
						runErr = stderrors.Join(runErr, fmt.Errorf("程序捕获不能处理的异常: %v", e))
						log.Error("程序捕获不能处理的异常", zap.Error(runErr))
					} else {
						runErr = stderrors.Join(runErr, converted)
						log.Error("程序捕获不能处理的异常", converted.GetFieldsWithCause()...)
					}
					cancelRun()
				}
			}()
			if err = runCtx.Err(); err != nil {
				return err
			}
			if os.Getenv(Env_DaemonMode) == "true" {
				log.Debug("守护进程模式运行", zap.String("ServiceName", serviceInfo.ServiceName))
				pidFileLocker, err := OpenPidFileLocker(GetPidFilePath(serviceInfo.ServiceName), os.ModePerm)
				if err != nil {
					log.Error("打开进程文件失败", zap.Error(err))
					return err
				}
				pid, _ := pidFileLocker.ReadPid()
				if pid > 0 {
					log.Debug("获取了已经在运行的程序", zap.Int("pid", pid))
					process, _ := os.FindProcess(pid)
					err = process.Signal(syscall.Signal(0x0))
					if err == nil {
						log.Debug("mobileNodePid已经在运行", zap.Any("pid", process.Pid))
						return nil
					}
				}
				err = pidFileLocker.WritePid()
				if err != nil {
					return err
				}
				err = pidFileLocker.Lock()
				if err != nil {
					return err
				}
				defer func() {
					pidFileLocker.Unlock()
					pidFileLocker.Remove()
				}()
				_, err = os.Stat(GetPidFilePath(serviceInfo.ServiceName))
				if err != nil {
					log.Error("进程文件创建失败")
					return err
				}
				log.Debug("进程文件已经创建好了")
			}
			var closeCallback ServiceCloseCallback
			// 在 PID 文件清理之后登记 defer，保证业务和插件停止完成后才释放守护进程身份。
			defer func() {
				cancelRun()
				shutdownCtx, cancelShutdown := context.WithTimeout(context.WithoutCancel(runCtx), shutdownTimeout)
				defer cancelShutdown()
				if closeCallback != nil {
					runErr = stderrors.Join(runErr, closeService(closeCallback))
				}
				runErr = stderrors.Join(runErr, plugin.StopPluginsWithError(shutdownCtx))
			}()
			configuration.InitConfiguration(runCtx, serviceInfo)
			closeCallback, err = start(runCtx, cmd, args)
			if err != nil {
				log.Error("启动服务失败", zap.Error(err))
				return err
			}
			if err = plugin.StartPluginsWithCancel(runCtx, cancelRun); err != nil {
				return err
			}
			log.Debug("插件全部启动完成")
			<-runCtx.Done()
			return nil
		},
	}
}

// closeService 隔离业务关闭回调的 panic，使插件清理继续执行并向命令保留关闭失败。
func closeService(callback ServiceCloseCallback) (closeErr error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if cause, ok := recovered.(error); ok {
				closeErr = fmt.Errorf("业务关闭回调 panic: %w", cause)
			} else {
				closeErr = fmt.Errorf("业务关闭回调 panic: %v", recovered)
			}
			log.Error("业务关闭回调失败", zap.Error(closeErr))
		}
	}()
	callback()
	return nil
}
