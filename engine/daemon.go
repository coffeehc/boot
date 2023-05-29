package engine

import (
	"fmt"
	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const ENVDaemonIndex = "__DAEMON_IDX"

// 运行时调用background的次数
var runIdx int = 0

//// 守护进程
//type DaemonContext struct {
//	LogFile string //日志文件, 记录守护进程和子进程的标准输出和错误输出. 若为空则不记录
//	//MaxError    int    //连续启动失败或异常退出的最大次数, 超过此数, 守护进程退出, 不再重启子进程
//	MinExitTime int64 //子进程正常退出的最小时间(秒). 小于此时间则认为是异常退出
//}

// 把本身程序转化为后台运行(启动一个子进程, 然后自己退出)
// logFile 若不为空,子程序的标准输出和错误输出将记入此文件
// isExit  启动子加进程后是否直接退出主程序, 若为false, 主程序返回*os.Process, 子程序返回 nil. 需自行判断处理
func Background(appPath string, args []string, logFile string) (*exec.Cmd, error) {
	//判断子进程还是父进程
	runIdx++
	envIdx, err := strconv.Atoi(os.Getenv(ENVDaemonIndex))
	if err != nil {
		envIdx = 0
	}
	if runIdx <= envIdx { //子进程, 退出
		log.Error("当前子进程变量错误", zap.Int("runIdx", runIdx), zap.Int("envIdx", envIdx))
		return nil, errors.SystemError("当前子进程变量错误")
	}

	//设置子进程环境变量
	env := os.Environ()
	//log.Debug("envs", zap.Strings("env", env))
	needAppend := true
	runInxEnv := fmt.Sprintf("%s=%d", ENVDaemonIndex, runIdx)
	for i, e := range env {
		if strings.HasPrefix(e, ENVDaemonIndex) {
			env[i] = runInxEnv
			needAppend = false
		}
	}
	if needAppend {
		env = append(env, runInxEnv)
	}
	//启动子进程
	cmd, err := startProc(appPath, args, env, logFile)
	if err != nil {
		//log.Println(os.Getpid(), "启动子进程失败:", err)
		log.Error("启动进程失败", zap.Error(err))
		return nil, err
	}
	return cmd, nil
}

func startProc(appPath string, args, env []string, logFile string) (*exec.Cmd, error) {
	//log.Debug("启动", zap.String("path", appPath), zap.Strings("args", args), zap.Strings("env", env))
	cmd := &exec.Cmd{
		Path:        appPath,
		Args:        args,
		Env:         env,
		SysProcAttr: NewSysProcAttr(),
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if logFile != "" {
		stdout, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Error("打开日志文件错误", zap.Error(err))
			return nil, err
		}
		defer stdout.Close()
		cmd.Stderr = stdout
		cmd.Stdout = stdout
	}
	err := cmd.Start()
	if err != nil {
		log.Error("启动错误", zap.Error(err))
		return nil, err
	}
	return cmd, nil
}
