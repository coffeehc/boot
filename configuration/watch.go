package configuration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var errSourceChanged = errors.New("配置源读取期间发生变化")

// stableConfigRead 核对打开文件与当前路径的身份、大小和修改时间，避免跨 rename/写入拼接来源。
func stableConfigRead(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, errors.New("配置源不是普通文件")
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	after, err := f.Stat()
	if err != nil {
		return nil, err
	}
	current, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !os.SameFile(before, current) || !before.ModTime().Equal(after.ModTime()) || !after.ModTime().Equal(current.ModTime()) || before.Size() != after.Size() || after.Size() != current.Size() || int64(len(data)) != after.Size() {
		return nil, errSourceChanged
	}
	return data, nil
}

type configTimer interface {
	channel() <-chan time.Time
	reset(time.Duration)
	stop()
}

type realConfigTimer struct{ timer *time.Timer }

func (t realConfigTimer) channel() <-chan time.Time { return t.timer.C }
func (t realConfigTimer) stop() {
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
}
func (t realConfigTimer) reset(delay time.Duration) { t.stop(); t.timer.Reset(delay) }

type configOwner struct {
	mu       sync.Mutex // 唯一 Viper 的所有配置包读取、解码和文件替换均在此锁内。
	ctx      context.Context
	cancel   context.CancelFunc
	vp       *viper.Viper
	path     string
	delay    time.Duration
	bindings map[any]*configBinding
	watcher  *fsnotify.Watcher
	done     chan struct{}
	newTimer func(time.Duration) configTimer
}

func newConfigOwner(ctx context.Context, vp *viper.Viper, path string, delay time.Duration) *configOwner {
	ctx, cancel := context.WithCancel(ctx)
	return &configOwner{ctx: ctx, cancel: cancel, vp: vp, path: path, delay: delay, bindings: map[any]*configBinding{}, newTimer: func(delay time.Duration) configTimer { return realConfigTimer{time.NewTimer(delay)} }}
}

// ensureWatcher 只建立一个目录监听，保留文件被删除/原子替换后继续观察恢复的能力。
func (o *configOwner) ensureWatcher() error {
	if o.watcher != nil {
		return nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := w.Add(filepath.Dir(o.path)); err != nil {
		w.Close()
		return err
	}
	o.watcher, o.done = w, make(chan struct{})
	go func() { defer close(o.done); defer w.Close(); o.watch(w.Events, w.Errors) }()
	return nil
}

func (o *configOwner) stop() {
	o.cancel()
	o.mu.Lock()
	w, done := o.watcher, o.done
	o.mu.Unlock()
	if w != nil {
		_ = w.Close()
	}
	if done != nil {
		<-done
	}
}

func (o *configOwner) relevant(event fsnotify.Event) bool {
	return filepath.Clean(event.Name) == o.path && event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove|fsnotify.Chmod) != 0
}

func (o *configOwner) watch(events <-chan fsnotify.Event, failures <-chan error) {
	var timer configTimer
	var tick <-chan time.Time
	defer func() {
		if timer != nil {
			timer.stop()
		}
	}()
	schedule := func() {
		if timer == nil {
			timer = o.newTimer(o.delay)
		} else {
			timer.reset(o.delay)
		}
		tick = timer.channel()
	}
	for {
		select {
		case <-o.ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if o.relevant(event) {
				schedule()
			}
		case _, ok := <-failures:
			if !ok {
				failures = nil
				continue
			}
			// 溢出等监听异常后重新核对当前来源，不能用异常文本泄漏路径或配置。
			schedule()
		case <-tick:
			tick = nil
			// 已到达的事件优先重置尾沿，避免 timer/event 同时就绪时先读旧版本。
			dirty := false
		drain:
			for {
				select {
				case event, ok := <-events:
					if !ok {
						return
					}
					if o.relevant(event) {
						dirty = true
					}
				default:
					break drain
				}
			}
			if dirty {
				schedule()
				continue
			}
			if errors.Is(o.reload(), errSourceChanged) {
				schedule()
			}
		}
	}
}

type sourceFailure struct {
	stage  string
	digest [32]byte
}
type nodeFailure struct {
	message string
	value   any
}

func (o *configOwner) failAll(stage string, data []byte) {
	stamp := sourceFailure{stage, sha256.Sum256(data)}
	for _, b := range o.bindings {
		if b.notify == nil || (b.failed && reflect.DeepEqual(b.lastBad, stamp)) {
			continue
		}
		b.failed, b.lastBad = true, stamp
		b.enqueue(configNotice{err: nodeError(b.key, stage)})
	}
}

func (o *configOwner) reload() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.ctx.Err(); err != nil {
		return err
	}
	data, err := stableConfigRead(o.path)
	if errors.Is(err, errSourceChanged) {
		return err
	}
	if err != nil {
		o.failAll("读取配置源失败", nil)
		return nil
	}
	// Viper 1.21 ReadConfig 先解析临时 map，成功才替换文件 map；没有第二 Viper/解析器。
	if err := o.vp.ReadConfig(bytes.NewReader(data)); err != nil {
		o.failAll("解析配置源失败", data)
		return nil
	}
	for _, b := range o.bindings {
		if b.notify == nil {
			continue
		}
		candidate, err := o.decode(b)
		if err != nil {
			var value any
			if raw := o.vp.Get(b.key); raw != nil {
				value = copyConfig(reflect.ValueOf(raw)).Interface()
			}
			stamp := nodeFailure{err.Error(), value}
			if !b.failed || !reflect.DeepEqual(b.lastBad, stamp) {
				b.enqueue(configNotice{err: err})
			}
			b.failed, b.lastBad = true, stamp
			continue
		}
		if !b.failed && reflect.DeepEqual(candidate, b.lastGood) {
			continue
		}
		b.lastGood = copyConfig(reflect.ValueOf(candidate)).Interface()
		b.failed, b.lastBad = false, nil
		b.enqueue(configNotice{candidate: candidate})
	}
	return nil
}
