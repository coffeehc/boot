package manage

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/coffeehc/httpx"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v2"
)

// observedHTTPService retains the real httpx listener result for fixture cleanup.
type observedHTTPService struct {
	httpx.Service
	stopped <-chan error
}

func (service *observedHTTPService) Start(onShutdown func()) <-chan error {
	service.stopped = service.Service.Start(onShutdown)
	return service.stopped
}

func TestManagerStartsAndStopsWithUpgradedHTTPX(t *testing.T) {
	webRoot, err := fs.Sub(views, "views")
	if err != nil {
		t.Fatal(err)
	}
	viewEngine := html.NewFileSystem(http.FS(webRoot), ".gohtml")
	viewEngine.Layout("embed")
	service := &observedHTTPService{Service: httpx.NewService(&httpx.Config{
		AppName: "manage-upgrade-test", ServerAddr: "127.0.0.1:0", Views: viewEngine,
	})}
	if service.Service == nil {
		t.Fatal("httpx.NewService returned nil")
	}
	impl := &serviceImpl{httpService: service, endpoint: service.GetServerAddress()}
	previous := _plugin
	_plugin = impl
	t.Cleanup(func() { _plugin = previous })
	ready := make(chan struct{}, 1)
	service.GetEngine().Hooks().OnListen(func(fiber.ListenData) error {
		ready <- struct{}{}
		return nil
	})
	if err := impl.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := impl.Stop(t.Context()); err != nil {
			t.Error(err)
		}
		select {
		case err := <-service.stopped:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(5 * time.Second):
			t.Error("httpx listener did not finish after manager Stop")
		}
	})
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("manager listener did not become ready")
	}
	transport := &http.Transport{}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	for _, route := range []struct {
		path         string
		wantStatus   int
		bodyContains string
	}{
		{"/", http.StatusOK, "/ping"},
		{"/ping", http.StatusOK, "pong"},
		{"/info", http.StatusOK, "go_version"},
		{"/health", http.StatusOK, "goroutine_count"},
		{"/metrics", http.StatusOK, "# HELP"},
		{"/missing-upgrade-route", http.StatusNotFound, ""},
	} {
		t.Run(route.path, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://"+impl.endpoint+route.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			response, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil || response.StatusCode != route.wantStatus || !strings.Contains(string(body), route.bodyContains) {
				t.Fatalf("status=%d body=%s err=%v", response.StatusCode, body, err)
			}
		})
	}
}

func TestManagerRoutesWithFiberV3(t *testing.T) {
	app := fiber.New()
	RegisterManager(app)
	for _, path := range []string{"/ping", "/info", "/health", "/metrics"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Accept", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status=%d body=%s", resp.StatusCode, body)
			}
			switch path {
			case "/ping":
				if string(body) != "pong" {
					t.Fatalf("body=%s", body)
				}
			case "/info":
				var info serviceRuntimeInfo
				if err := json.Unmarshal(body, &info); err != nil || info.GoVersion != runtime.Version() {
					t.Fatalf("runtime info=%s error=%v", body, err)
				}
			case "/health":
				var info map[string]any
				if err := json.Unmarshal(body, &info); err != nil || info["goroutine_count"] == nil {
					t.Fatalf("health info=%s error=%v", body, err)
				}
			case "/metrics":
				if !strings.Contains(string(body), "# HELP") {
					t.Fatalf("metrics handler did not emit Prometheus metrics: %s", body)
				}
			}
		})
	}
}

func TestManagerParsesIntegerQueryWithFiberV3(t *testing.T) {
	previous := debug.SetGCPercent(100)
	defer debug.SetGCPercent(previous)
	app := fiber.New()
	RegisterManager(app)
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/gc/stats/setgogc?gogc=91", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if got := debug.SetGCPercent(100); got != 91 {
		t.Fatalf("GC percent=%d, want 91", got)
	}
}
