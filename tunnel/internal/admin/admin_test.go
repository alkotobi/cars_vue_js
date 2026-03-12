package admin_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/admin"
	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/registry"
)

func TestHealth(t *testing.T) {
	cfg := config.ServerDefaults()
	cfg.AdminPort = 18083
	reg := registry.New()
	logger := log.New(io.Discard, 0)
	srv := admin.New(cfg, reg, logger)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	resp, err := http.Get("http://127.0.0.1:18083/api/health")
	if err != nil {
		t.Fatalf("GET /api/health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d", resp.StatusCode)
	}
	var out struct {
		Status     string `json:"status"`
		Version    string `json:"version"`
		UptimeSec  int64  `json:"uptime_sec"`
		Goroutines int    `json:"goroutines"`
		Clients    int    `json:"clients"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Status != "ok" {
		t.Errorf("status: got %q", out.Status)
	}
	if out.Version != config.Version {
		t.Errorf("version: got %q", out.Version)
	}
	if out.Clients != 0 {
		t.Errorf("clients: got %d", out.Clients)
	}
}

func TestStats(t *testing.T) {
	cfg := config.ServerDefaults()
	cfg.AdminPort = 18084
	reg := registry.New()
	_ = reg.Register(&registry.Client{Domain: "test"})
	logger := log.New(io.Discard, 0)
	srv := admin.New(cfg, reg, logger)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	resp, err := http.Get("http://127.0.0.1:18084/api/stats")
	if err != nil {
		t.Fatalf("GET /api/stats: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d", resp.StatusCode)
	}
	var out struct {
		RegisteredClients     int  `json:"registered_clients"`
		MaxRegistered         int  `json:"max_registered"`
		MaxActiveProxyPerNode int  `json:"max_active_proxy_per_node"`
		ScaleNote             string `json:"scale_note"`
		Memory                struct {
			AllocMB  float64 `json:"alloc_mb"`
			SysMB    float64 `json:"sys_mb"`
			GCCycles uint32  `json:"gc_cycles"`
		} `json:"memory"`
		Clients []struct {
			Domain string `json:"domain"`
		} `json:"clients"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.RegisteredClients != 1 {
		t.Errorf("registered_clients: got %d", out.RegisteredClients)
	}
	if out.ScaleNote == "" {
		t.Error("scale_note empty")
	}
	if len(out.Clients) != 1 || out.Clients[0].Domain != "test" {
		t.Errorf("clients: got %+v", out.Clients)
	}
}

func TestClientsList(t *testing.T) {
	cfg := config.ServerDefaults()
	cfg.AdminPort = 18085
	reg := registry.New()
	logger := log.New(io.Discard, 0)
	srv := admin.New(cfg, reg, logger)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	resp, err := http.Get("http://127.0.0.1:18085/api/clients")
	if err != nil {
		t.Fatalf("GET /api/clients: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d", resp.StatusCode)
	}
	var out []struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("clients: got %+v", out)
	}
}

func TestDisconnect(t *testing.T) {
	cfg := config.ServerDefaults()
	cfg.AdminPort = 18086
	reg := registry.New()
	c := &registry.Client{Domain: "disco"}
	_ = reg.Register(c)
	logger := log.New(io.Discard, 0)
	srv := admin.New(cfg, reg, logger)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:18086/api/clients/disco/disconnect", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST disconnect: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: got %d", resp.StatusCode)
	}
	_, ok := reg.Find("disco")
	if ok {
		t.Error("client should be unregistered after disconnect")
	}
}

func TestDisconnectNotFound(t *testing.T) {
	cfg := config.ServerDefaults()
	cfg.AdminPort = 18087
	reg := registry.New()
	logger := log.New(io.Discard, 0)
	srv := admin.New(cfg, reg, logger)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:18087/api/clients/nonexistent/disconnect", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", resp.StatusCode)
	}
}

func TestPprofDisabled(t *testing.T) {
	cfg := config.ServerDefaults()
	cfg.AdminPort = 18088
	cfg.PProfEnabled = false
	reg := registry.New()
	logger := log.New(io.Discard, 0)
	srv := admin.New(cfg, reg, logger)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	resp, err := http.Get("http://127.0.0.1:18088/debug/pprof/")
	if err != nil {
		t.Fatalf("GET pprof: %v", err)
	}
	defer resp.Body.Close()
	// When pprof is disabled, the route is not registered so we get 404.
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("pprof should be disabled: got status %d", resp.StatusCode)
	}
}
