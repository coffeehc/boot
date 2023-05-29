package engine

import (
	"fmt"
	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
	"io"
	"os"
)

func ReadPidFile(name string) (pid int, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDONLY, 0640); err != nil {
		log.Error("打开pid文件失败", zap.Error(err))
		return
	}
	defer file.Close()
	lock := &PidFileLocker{file}
	pid, err = lock.ReadPid()
	return
}

func NewPidFileLocker(file *os.File) *PidFileLocker {
	return &PidFileLocker{file}
}

func CreatePidFileLocker(name string, perm os.FileMode) (lock *PidFileLocker, err error) {
	if lock, err = OpenPidFileLocker(name, perm); err != nil {
		return
	}
	if err = lock.Lock(); err != nil {
		lock.Remove()
		return
	}
	if err = lock.WritePid(); err != nil {
		lock.Remove()
	}
	return
}

func OpenPidFileLocker(name string, perm os.FileMode) (lock *PidFileLocker, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm); err == nil {
		lock = &PidFileLocker{file}
	}
	return
}

type PidFileLocker struct {
	*os.File
}

func (file *PidFileLocker) Lock() error {
	return lockFile(file.Fd())
}

func (file *PidFileLocker) Unlock() error {
	return unlockFile(file.Fd())
}

func (file *PidFileLocker) WritePid() (err error) {
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}
	var fileLen int
	if fileLen, err = fmt.Fprint(file, os.Getpid()); err != nil {
		return
	}
	if err = file.Truncate(int64(fileLen)); err != nil {
		return
	}
	err = file.Sync()
	return
}

func (file *PidFileLocker) ReadPid() (pid int, err error) {
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}
	_, err = fmt.Fscan(file, &pid)
	return
}

func (file *PidFileLocker) Remove() error {
	defer file.Close()

	if err := file.Unlock(); err != nil {
		return err
	}

	return os.Remove(file.Name())
}
