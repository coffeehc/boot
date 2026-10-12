package consul_rc

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coffeehc/boot/configuration"
	"github.com/hashicorp/consul/api"
)

// TestService exercises the real Consul client against a local HTTP fixture.
// It needs neither a running Consul agent nor process-global plugin state.
func TestService(t *testing.T) {
	for _, test := range []struct {
		name    string
		status  int
		timeout bool
	}{
		{name: "register_and_deregister", status: http.StatusOK},
		{name: "registration_error_is_reported", status: http.StatusServiceUnavailable},
		{name: "registration_timeout_is_reported", timeout: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Prevent developer ACL credentials from being read or sent to the fixture.
			t.Setenv(api.HTTPTokenEnvName, "")
			t.Setenv(api.HTTPTokenFileEnvName, "")
			var mu sync.Mutex
			var registrations []api.AgentServiceRegistration
			var deregistered []string
			consul := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("method = %s, want PUT", r.Method)
					http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
					return
				}
				switch {
				case r.URL.Path == "/v1/agent/service/register":
					if got := r.URL.Query().Get("replace-existing-checks"); got != "true" {
						t.Errorf("replace-existing-checks = %q, want true", got)
					}
					var registration api.AgentServiceRegistration
					if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
						t.Errorf("decode registration: %v", err)
						http.Error(w, "invalid registration", http.StatusBadRequest)
						return
					}
					mu.Lock()
					registrations = append(registrations, registration)
					mu.Unlock()
					if test.timeout {
						select {
						case <-r.Context().Done():
						case <-time.After(5 * time.Second):
							t.Error("registration request was not canceled")
						}
						return
					}
					w.WriteHeader(test.status)
				case strings.HasPrefix(r.URL.Path, "/v1/agent/service/deregister/"):
					mu.Lock()
					deregistered = append(deregistered, strings.TrimPrefix(r.URL.Path, "/v1/agent/service/deregister/"))
					mu.Unlock()
					w.WriteHeader(http.StatusOK)
				default:
					t.Errorf("unexpected Consul request: %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(consul.Close)
			client, err := api.NewClient(&api.Config{
				Address:    consul.URL,
				Scheme:     "http",
				HttpClient: consul.Client(),
			})
			if err != nil {
				t.Fatalf("create Consul client: %v", err)
			}
			requestTimeout := 5 * time.Second
			if test.timeout {
				requestTimeout = 200 * time.Millisecond
			}
			address := consul.Listener.Addr().String()
			service, err := newService(context.Background(), &Config{
				Client:         client,
				ServiceAddress: address,
				ServiceID:      "boot-test-instance",
				Tags:           []string{"test"},
				RequestTimeout: requestTimeout,
			})
			if err != nil {
				t.Fatalf("create registration service: %v", err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := service.Stop(ctx); err != nil {
					t.Errorf("cleanup registration service: %v", err)
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := service.Start(ctx); err != nil {
				t.Fatalf("start registration service: %v", err)
			}
			serviceInfo := configuration.ServiceInfo{
				ServiceName: "boot-test",
				Version:     "0.0.1",
				Descriptor:  "Consul registration test",
				APIDefine:   "test-api",
				Metadata:    map[string]string{"owner": "test"},
			}
			err = service.Register(ctx, serviceInfo)
			switch {
			case test.timeout:
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("Register error = %v, want deadline exceeded", err)
				}
			case test.status != http.StatusOK:
				if err == nil || !strings.Contains(err.Error(), strconv.Itoa(test.status)) {
					t.Fatalf("Register error = %v, want HTTP %d", err, test.status)
				}
			default:
				if err != nil {
					t.Fatalf("register service: %v", err)
				}
			}
			if !maps.Equal(serviceInfo.Metadata, map[string]string{"owner": "test"}) {
				t.Errorf("Register mutated caller metadata: %v", serviceInfo.Metadata)
			}
			// Failed/unknown registration outcomes also require deregistration.
			if err := service.Stop(ctx); err != nil {
				t.Fatalf("stop registration service: %v", err)
			}
			if err := service.Stop(ctx); err != nil {
				t.Fatalf("stop registration service twice: %v", err)
			}
			if err := service.Register(ctx, serviceInfo); err == nil {
				t.Error("Register after Stop succeeded")
			}
			mu.Lock()
			defer mu.Unlock()
			if len(registrations) != 1 {
				t.Fatalf("registration requests = %d, want 1", len(registrations))
			}
			registration := registrations[0]
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				t.Fatalf("split test address: %v", err)
			}
			if registration.ID != "boot-test-instance" || registration.Name != serviceInfo.ServiceName || registration.Address != host || strconv.Itoa(registration.Port) != port {
				t.Errorf("unexpected registration identity/address: %+v", registration)
			}
			if !slices.Equal(registration.Tags, []string{"test"}) {
				t.Errorf("registration tags = %v, want [test]", registration.Tags)
			}
			wantMetadata := map[string]string{
				"owner": "test", "Version": serviceInfo.Version, "Descriptor": serviceInfo.Descriptor,
				"APIDefine": serviceInfo.APIDefine, "Address": address,
			}
			if !maps.Equal(registration.Meta, wantMetadata) {
				t.Errorf("registration metadata = %v, want %v", registration.Meta, wantMetadata)
			}
			if check := registration.Check; check == nil || check.GRPC != address || check.GRPCUseTLS || check.CheckID != "boot-test-instance_grpcHealth" {
				t.Errorf("unexpected gRPC health check: %+v", check)
			}
			if !slices.Equal(deregistered, []string{"boot-test-instance"}) {
				t.Errorf("deregistered IDs = %v, want exactly [boot-test-instance]", deregistered)
			}
		})
	}
}
