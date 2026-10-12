package configuration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/viper"
)

type boundConfig struct {
	Limit  int            `mapstructure:"limit"`
	Names  []string       `mapstructure:"names"`
	Nested map[string]any `mapstructure:"nested"`
}

func (c *boundConfig) Validate() error {
	if c.Limit <= 0 {
		return errors.New("do-not-leak-secret-value")
	}
	return nil
}

func configTest(t *testing.T, contents string) (string, context.CancelFunc, *configOwner) {
	t.Helper()
	oldPath, oldDefault := *configFile, defaultRunModel
	ctx, cancel := context.WithCancel(context.Background())
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	*configFile = path
	SetRunModel(Model_test)
	if err := InitConfiguration(ctx, ServiceInfo{ServiceName: "isolated-binding-test"}, WithFileOnly(), WithReloadDebounce(15*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	o := configurationState.owner
	t.Cleanup(func() {
		cancel()
		configurationState.Lock()
		if configurationState.owner != nil {
			configurationState.owner.stop()
			configurationState.owner = nil
		}
		*configFile, defaultRunModel = oldPath, oldDefault
		viper.Reset()
		configurationState.Unlock()
	})
	return path, cancel, o
}

func waitNotice(t *testing.T, ch <-chan configNotice) configNotice {
	t.Helper()
	select {
	case n := <-ch:
		return n
	case <-time.After(2 * time.Second):
		t.Fatal("configuration notification timed out")
		return configNotice{}
	}
}

func rewrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestBindConfigFirstLoadValidationAndNilSubscription(t *testing.T) {
	path, _, o := configTest(t, "resource:\n  limit: 3\n  names: [first]\n  nested: {secret: original}\nbad: {limit: 0}\nwrong: {limit: do-not-leak-secret-value}\n")
	c := boundConfig{Limit: 91, Names: []string{"unchanged"}}
	if err := BindConfig("resource", &c, nil); err != nil || c.Limit != 3 || c.Names[0] != "first" {
		t.Fatal("first load failed", err, c)
	}
	if o.watcher != nil {
		t.Fatal("nil callback created a subscription watcher")
	}
	for _, target := range []any{boundConfig{}, (*boundConfig)(nil), new(int)} {
		if err := BindConfig("resource", target, nil); err == nil {
			t.Fatal("invalid target accepted")
		}
	}
	if err := BindConfig("resource", &c, nil); err == nil {
		t.Fatal("same object bound twice")
	}
	for _, key := range []string{"bad", "wrong", "missing"} {
		original := boundConfig{Limit: 91, Names: []string{"unchanged"}}
		before := copyConfig(reflect.ValueOf(original)).Interface()
		if err := BindConfig(key, &original, func(any, error) {}); err == nil || strings.Contains(err.Error(), "do-not-leak") || !strings.Contains(err.Error(), key) {
			t.Fatal("invalid node not attributed/redacted", key, err)
		}
		if !reflect.DeepEqual(before, original) {
			t.Fatal("failed initial decode/validation polluted original")
		}
		if _, exists := o.bindings[&original]; exists {
			t.Fatal("initial failure subscribed")
		}
	}
	rewrite(t, path, "resource: {limit: 9}\n")
	if err := o.reload(); err != nil {
		t.Fatal(err)
	}
	if c.Limit != 3 || c.Names[0] != "first" {
		t.Fatal("nil callback did not remain initial-only")
	}
}

func TestBindConfigKeyPrecisionSameValueCandidatesAndErrorRecovery(t *testing.T) {
	path, _, o := configTest(t, "left: {limit: 1, names: [one]}\nright: {limit: 2}\n")
	left, right := &boundConfig{}, &boundConfig{}
	a, b := make(chan configNotice, 20), make(chan configNotice, 20)
	if err := BindConfig("left", left, func(c any, e error) { a <- configNotice{c, e} }); err != nil {
		t.Fatal(err)
	}
	if err := BindConfig("right", right, func(c any, e error) { b <- configNotice{c, e} }); err != nil {
		t.Fatal(err)
	}
	reload := func(body string) {
		rewrite(t, path, body)
		if err := o.reload(); err != nil {
			t.Fatal(err)
		}
	}
	reload("left: {limit: 3, names: [three]}\nright: {limit: 2}\n")
	n := waitNotice(t, a)
	x, ok := n.candidate.(*boundConfig)
	if !ok || n.err != nil || x == left || x.Limit != 3 || left.Limit != 1 {
		t.Fatal("update was not an independent candidate", n)
	}
	x.Limit, x.Names[0] = 999, "service-mutated"
	reload("left: {limit: '3', names: [three]}\nright: {limit: 2}\nother: true\n")
	select {
	case <-a:
		t.Fatal("same effective value notified")
	default:
	}
	select {
	case <-b:
		t.Fatal("unrelated key notified")
	default:
	}
	reload("left: {limit: 0}\nright: {limit: 4}\n")
	n = waitNotice(t, a)
	if n.candidate != nil || n.err == nil || strings.Contains(n.err.Error(), "secret") || !strings.Contains(n.err.Error(), "left") {
		t.Fatal("invalid key notice not safe", n)
	}
	if waitNotice(t, b).candidate.(*boundConfig).Limit != 4 {
		t.Fatal("healthy key was incorrectly blocked")
	}
	reload("left: [ do-not-leak-secret-value\n")
	for _, ch := range []chan configNotice{a, b} {
		n = waitNotice(t, ch)
		if n.candidate != nil || n.err == nil || strings.Contains(n.err.Error(), "do-not-leak") {
			t.Fatal("global parse error leaked data", n)
		}
	}
	reload("left: {limit: 3, names: [three]}\nright: {limit: 4}\n")
	if waitNotice(t, a).candidate.(*boundConfig).Limit != 3 || waitNotice(t, b).candidate.(*boundConfig).Limit != 4 {
		t.Fatal("same-value recovery did not clear failure")
	}
	if left.Limit != 1 || right.Limit != 2 {
		t.Fatal("configuration owner applied reload to service objects")
	}
}

func TestBindConfigAtomicRenameDeletionAndRecovery(t *testing.T) {
	path, _, _ := configTest(t, "resource: {limit: 1}\n")
	notices := make(chan configNotice, 20)
	c := &boundConfig{}
	if err := BindConfig("resource", c, func(v any, e error) { notices <- configNotice{v, e} }); err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(filepath.Dir(path), "next.yml")
	rewrite(t, tmp, "resource: {limit: 2}\n")
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
	if waitNotice(t, notices).candidate.(*boundConfig).Limit != 2 {
		t.Fatal("atomic replacement not observed")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	n := waitNotice(t, notices)
	if n.candidate != nil || n.err == nil {
		t.Fatal("deletion did not notify error")
	}
	rewrite(t, path, "resource: {limit: 2}\n")
	if waitNotice(t, notices).candidate.(*boundConfig).Limit != 2 {
		t.Fatal("watcher failed to recover after deletion")
	}
	if c.Limit != 1 {
		t.Fatal("file watcher edited service configuration")
	}
}

func TestBindConfigSlowCallbackOrderedIndependentAndShutdown(t *testing.T) {
	path, cancel, o := configTest(t, "slow: {limit: 1}\nfast: {limit: 1}\nother: {limit: 1}\n")
	start, release := make(chan struct{}), make(chan struct{})
	slowSeen, fastSeen := make(chan configNotice, 10), make(chan configNotice, 10)
	var once sync.Once
	slow, fast := &boundConfig{}, &boundConfig{}
	if err := BindConfig("slow", slow, func(v any, e error) {
		once.Do(func() { close(start); <-release })
		var probe boundConfig
		// Reentry is outside every global/owner/binding lock.
		if err := BindConfig("other", &probe, nil); err != nil && o.ctx.Err() == nil {
			t.Error(err)
		}
		slowSeen <- configNotice{v, e}
	}); err != nil {
		t.Fatal(err)
	}
	if err := BindConfig("fast", fast, func(v any, e error) { fastSeen <- configNotice{v, e} }); err != nil {
		t.Fatal(err)
	}
	rewrite(t, path, "slow: {limit: 2}\nfast: {limit: 2}\nother: {limit: 1}\n")
	if err := o.reload(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-start:
	case <-time.After(time.Second):
		t.Fatal("slow callback did not start")
	}
	if waitNotice(t, fastSeen).candidate.(*boundConfig).Limit != 2 {
		t.Fatal("slow callback blocked another binding")
	}
	rewrite(t, path, "slow: {limit: 3}\nfast: {limit: 3}\nother: {limit: 1}\n")
	if err := o.reload(); err != nil {
		t.Fatal(err)
	}
	if waitNotice(t, fastSeen).candidate.(*boundConfig).Limit != 3 {
		t.Fatal("global reload blocked by callback")
	}
	close(release)
	for _, want := range []int{2, 3} {
		if waitNotice(t, slowSeen).candidate.(*boundConfig).Limit != want {
			t.Fatal("binding callbacks out of order")
		}
	}
	cancel()
	o.stop()
	rewrite(t, path, "slow: {limit: 4}\nfast: {limit: 4}\n")
	if err := o.reload(); !errors.Is(err, context.Canceled) {
		t.Fatal("closed owner reloaded", err)
	}
	if err := BindConfig("other", &boundConfig{}, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("closed owner accepted binding", err)
	}
}

func TestShutdownDoesNotWaitForSlowCallbackOrRunPendingNotices(t *testing.T) {
	path, cancel, o := configTest(t, "resource: {limit: 1}\n")
	entered, release, returned := make(chan struct{}), make(chan struct{}), make(chan struct{})
	c := &boundConfig{}
	var calls atomic.Int32
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	if err := BindConfig("resource", c, func(any, error) { calls.Add(1); close(entered); <-release; close(returned) }); err != nil {
		t.Fatal(err)
	}
	rewrite(t, path, "resource: {limit: 2}\n")
	if err := o.reload(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("callback not entered")
	}
	rewrite(t, path, "resource: {limit: 3}\n")
	if err := o.reload(); err != nil {
		t.Fatal(err)
	}
	cancel()
	stopped := make(chan struct{})
	go func() { o.stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("watcher shutdown waited for service callback")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("callback did not return")
	}
	select {
	case <-o.bindings[c].done:
	case <-time.After(time.Second):
		t.Fatal("binding worker did not exit")
	}
	if calls.Load() != 1 {
		t.Fatal("shutdown delivered a pending notice")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if c.Limit != 1 {
		t.Fatal("shutdown applied a pending candidate")
	}
}
