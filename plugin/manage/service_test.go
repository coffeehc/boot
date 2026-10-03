package manage

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

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
