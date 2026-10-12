package configuration

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/coffeehc/base/log"
)

// ChangeNotify 接收同类型的独立配置候选；成功为 (candidate, nil)，失败为 (nil, err)。
// 配置包不会替服务应用候选；回调完成也不表示资源已经切换。
// 同一绑定的回调顺序执行，不同绑定互不等待。回调须及时返回并配合服务根 context 停机。
type ChangeNotify func(candidate any, err error)

// BindConfig 在服务 Start 开头同步加载 key 节点到非 nil 结构体指针。
// 使用当前 Boot Viper 的解码规则；候选实现 Validate() error 时先执行其业务校验。
// 解码或校验失败不覆盖原对象，也不建立订阅。默认值应在 Viper 中注册。
// nil 回调只加载首次值；非 nil 回调只接收后续独立候选，从不再次修改原对象。
// 同一初始化轮次内，一个对象指针只能绑定一次；不同对象可绑定相同节点。
func BindConfig(key string, config any, onChange ChangeNotify) error {
	key = strings.ToLower(strings.TrimSpace(key))
	v := reflect.ValueOf(config)
	if key == "" || !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("配置绑定需要非空节点和非 nil 结构体指针")
	}
	configurationState.RLock()
	o := configurationState.owner
	configurationState.RUnlock()
	if o == nil {
		return fmt.Errorf("配置源尚未初始化")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.ctx.Err(); err != nil {
		return err
	}
	if _, exists := o.bindings[config]; exists {
		return fmt.Errorf("配置对象已经绑定")
	}
	b := &configBinding{key: key, typ: v.Type(), notify: onChange, wake: make(chan struct{}, 1), done: make(chan struct{})}
	candidate, err := o.decode(b)
	if err != nil {
		return err
	}
	if onChange != nil {
		if err := o.ensureWatcher(); err != nil {
			return nodeError(key, "监听配置源失败")
		}
	}
	if err := o.ctx.Err(); err != nil {
		return err
	}
	b.lastGood = copyConfig(reflect.ValueOf(candidate)).Interface()
	v.Elem().Set(reflect.ValueOf(candidate).Elem())
	o.bindings[config] = b
	if onChange != nil {
		go b.run(o.ctx)
	}
	return nil
}

func nodeError(key, stage string) error {
	// 不拼接底层解析错误、校验错误或 panic；这些内容可能包含配置秘密。
	return fmt.Errorf("配置节点 %q：%s", key, stage)
}

func (o *configOwner) decode(b *configBinding) (candidate any, err error) {
	defer func() {
		if recover() != nil {
			candidate, err = nil, nodeError(b.key, "解码或校验异常")
		}
	}()
	if !o.vp.IsSet(b.key) {
		return nil, nodeError(b.key, "节点不存在")
	}
	x := reflect.New(b.typ.Elem()).Interface()
	if err := o.vp.UnmarshalKey(b.key, x); err != nil {
		return nil, nodeError(b.key, "解码失败")
	}
	// 包括 any 字段中的 map/slice，避免候选引用 Viper 持有的值。
	x = copyConfig(reflect.ValueOf(x)).Interface()
	if validator, ok := x.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return nil, nodeError(b.key, "校验失败")
		}
	}
	return x, nil
}

type configNotice struct {
	candidate any
	err       error
}

type configBinding struct {
	key      string
	typ      reflect.Type
	notify   ChangeNotify
	lastGood any // owner.mu 保护，绝不交给回调修改。
	lastBad  any
	failed   bool
	mu       sync.Mutex
	queue    []configNotice
	wake     chan struct{}
	done     chan struct{}
}

func (b *configBinding) enqueue(notice configNotice) {
	b.mu.Lock()
	b.queue = append(b.queue, notice)
	b.mu.Unlock()
	select {
	case b.wake <- struct{}{}:
	default:
	}
}

func (b *configBinding) run(ctx context.Context) {
	if b.done != nil {
		defer close(b.done)
	}
	defer func() { b.mu.Lock(); b.queue = nil; b.mu.Unlock() }()
	for {
		select {
		case <-ctx.Done():
			return
		case <-b.wake:
			for {
				if ctx.Err() != nil {
					return
				}
				b.mu.Lock()
				if len(b.queue) == 0 {
					b.mu.Unlock()
					break
				}
				notice := b.queue[0]
				b.queue[0] = configNotice{}
				b.queue = b.queue[1:]
				b.mu.Unlock()
				if ctx.Err() != nil {
					return
				}
				func() {
					defer func() {
						if recover() != nil {
							log.Error("配置变更回调异常；请检查服务回调")
						}
					}()
					b.notify(notice.candidate, notice.err)
				}()
			}
		}
	}
}

// copyConfig 复制配置值，不引入另一套配置解码器；未导出字段不属于 Viper 节点。
func copyConfig(v reflect.Value) reflect.Value {
	type visit struct {
		typ    reflect.Type
		ptr    uintptr
		length int
	}
	seen := map[visit]reflect.Value{}
	var copyValue func(reflect.Value) reflect.Value
	copyValue = func(v reflect.Value) reflect.Value {
		if !v.IsValid() {
			return v
		}
		x := reflect.New(v.Type()).Elem()
		switch v.Kind() {
		case reflect.Pointer, reflect.Map, reflect.Slice:
			if v.IsNil() {
				return x
			}
			var ptr uintptr
			if v.Kind() == reflect.Map {
				ptr = uintptr(v.UnsafePointer())
			} else {
				ptr = v.Pointer()
			}
			length := 0
			if v.Kind() == reflect.Slice {
				length = v.Len()
			}
			id := visit{v.Type(), ptr, length}
			if previous, ok := seen[id]; ok {
				return previous
			}
			switch v.Kind() {
			case reflect.Pointer:
				x.Set(reflect.New(v.Type().Elem()))
				seen[id] = x
				x.Elem().Set(copyValue(v.Elem()))
			case reflect.Map:
				x.Set(reflect.MakeMapWithSize(v.Type(), v.Len()))
				seen[id] = x
				for it := v.MapRange(); it.Next(); {
					x.SetMapIndex(it.Key(), copyValue(it.Value()))
				}
			case reflect.Slice:
				x.Set(reflect.MakeSlice(v.Type(), v.Len(), v.Len()))
				seen[id] = x
				for i := 0; i < v.Len(); i++ {
					x.Index(i).Set(copyValue(v.Index(i)))
				}
			}
		case reflect.Interface:
			if !v.IsNil() {
				x.Set(copyValue(v.Elem()))
			}
		case reflect.Struct:
			x.Set(v)
			for i := 0; i < v.NumField(); i++ {
				if x.Field(i).CanSet() && v.Field(i).CanInterface() {
					x.Field(i).Set(copyValue(v.Field(i)))
				}
			}
		case reflect.Array:
			for i := 0; i < v.Len(); i++ {
				x.Index(i).Set(copyValue(v.Index(i)))
			}
		default:
			x.Set(v)
		}
		return x
	}
	return copyValue(v)
}
