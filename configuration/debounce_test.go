package configuration

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

type controlledClock struct {
	mu     sync.Mutex
	now    time.Duration
	timer  *controlledTimer
	resets chan time.Duration
}
type controlledTimer struct {
	clock  *controlledClock
	due    time.Duration
	active bool
	ch     chan time.Time
}

func (c *controlledClock) newTimer(d time.Duration) configTimer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timer = &controlledTimer{clock: c, due: c.now + d, active: true, ch: make(chan time.Time, 1)}
	c.resets <- d
	return c.timer
}
func (t *controlledTimer) channel() <-chan time.Time { return t.ch }
func (t *controlledTimer) stop()                     { t.clock.mu.Lock(); t.active = false; t.clock.mu.Unlock() }
func (t *controlledTimer) reset(d time.Duration) {
	t.clock.mu.Lock()
	t.due, t.active = t.clock.now+d, true
	t.clock.mu.Unlock()
	t.clock.resets <- d
}
func (c *controlledClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now += d
	if c.timer != nil && c.timer.active && c.now >= c.timer.due {
		c.timer.active = false
		c.timer.ch <- time.Unix(0, int64(c.now))
	}
}

func TestDebounceUsesControlledTrailingClockAndLatestSnapshot(t *testing.T) {
	path, _, o := configTest(t, "resource: {limit: 1}\n")
	clock := &controlledClock{resets: make(chan time.Duration, 20)}
	o.delay, o.newTimer = 5*time.Second, clock.newTimer
	notices := make(chan configNotice, 10)
	initial := &boundConfig{Limit: 1}
	b := &configBinding{key: "resource", typ: reflect.TypeOf(initial), notify: func(c any, e error) { notices <- configNotice{c, e} }, lastGood: &boundConfig{Limit: 1}, wake: make(chan struct{}, 1)}
	o.bindings[initial] = b
	go b.run(o.ctx)
	events, failures := make(chan fsnotify.Event, 10), make(chan error)
	done := make(chan struct{})
	go func() { defer close(done); o.watch(events, failures) }()
	t.Cleanup(func() { o.cancel(); <-done })
	waitReset := func() {
		t.Helper()
		select {
		case d := <-clock.resets:
			if d != 5*time.Second {
				t.Fatal("wrong debounce", d)
			}
		case <-time.After(time.Second):
			t.Fatal("event not consumed")
		}
	}
	waitReset() // Consume the initial reconciliation timer before delivering file events.
	rewrite(t, path, "resource: {limit: 2}\n")
	events <- fsnotify.Event{Name: path, Op: fsnotify.Write}
	waitReset()
	clock.advance(4 * time.Second)
	select {
	case <-notices:
		t.Fatal("read before quiet window")
	default:
	}
	rewrite(t, path, "resource: {limit: 3}\n")
	events <- fsnotify.Event{Name: path, Op: fsnotify.Write}
	waitReset()
	clock.advance(4 * time.Second)
	select {
	case <-notices:
		t.Fatal("new event failed to reset trailing window")
	default:
	}
	clock.advance(time.Second)
	if n := waitNotice(t, notices); n.err != nil || n.candidate.(*boundConfig).Limit != 3 {
		t.Fatal("debounce did not read final consistent version", n)
	}
	if initial.Limit != 1 {
		t.Fatal("debouncer applied candidate")
	}
}

func TestConcurrentBindingsAndReloadUseOneControlledViper(t *testing.T) {
	path, _, o := configTest(t, "resource: {limit: 1, nested: {value: before}}\n")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := &boundConfig{}
			if err := BindConfig("resource", c, func(v any, e error) {
				if e == nil {
					v.(*boundConfig).Nested["value"] = "callback-owned"
				}
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	for i := 0; i < 10; i++ {
		rewrite(t, path, "resource: {limit: 2, nested: {value: after}}\n")
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := o.reload(); err != nil && !errors.Is(err, errSourceChanged) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if err := o.reload(); err != nil {
		t.Fatal(err)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.vp.GetString("resource.nested.value") != "after" {
		t.Fatal("candidate alias edited the Viper source")
	}
}
