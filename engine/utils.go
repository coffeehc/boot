package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"go.uber.org/zap"
)

func daemon(serviceInfo configuration.ServiceInfo) error {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("systemctl", "restart", serviceInfo.ServiceName)
		err := cmd.Start()
		if err == nil {
			cmd.Wait()
			os.Exit(0)
		}
	}
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
