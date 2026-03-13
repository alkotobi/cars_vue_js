// Package admin provides an HTTP server on a separate admin port exposing
// health, stats, and pprof endpoints. It must NEVER be on the same port as
// the proxy server to avoid routing collisions with tunnel IDs.
//
// Security: the admin server binds to 127.0.0.1 by default so it is not
// exposed to the network. In production, put this behind nginx with authentication.
// The admin port must NOT be exposed to the internet.

package admin

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/registry"
)

// Server is the admin HTTP server. It runs on a separate port (cfg.AdminPort),
// bound to 127.0.0.1. startTime is used for uptime; srv is set in Start().
type Server struct {
	cfg       config.Config
	registry  registry.Registrar
	srv       *http.Server
	listener  net.Listener
	serveDone chan struct{}
	startTime time.Time
	log       *log.Logger
}

// New returns an admin Server. Call Start() to begin listening (non-blocking).
func New(cfg config.Config, r registry.Registrar, l *log.Logger) *Server {
	return &Server{
		cfg:       cfg,
		registry:  r,
		startTime: time.Now(),
		log:       l,
	}
}

// Start starts the admin HTTP server on 127.0.0.1:AdminPort. It is non-blocking;
// a goroutine runs the server. Returns an error if listening fails.
func (s *Server) Start() error {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(s.cfg.AdminPort))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.serveDone = make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/stats", s.handleStats)
	mux.HandleFunc("GET /api/clients", s.handleClientsList)
	mux.HandleFunc("POST /api/clients/{domain}/disconnect", s.handleDisconnect)

	if s.cfg.PProfEnabled {
		// pprof endpoints are opt-in to avoid accidental exposure.
		mux.HandleFunc("GET /debug/pprof/", s.handlePprofIndex)
		mux.HandleFunc("GET /debug/pprof/cmdline", s.handlePprofCmdline)
		mux.HandleFunc("GET /debug/pprof/profile", s.handlePprofProfile)
		mux.HandleFunc("GET /debug/pprof/symbol", s.handlePprofSymbol)
		mux.HandleFunc("GET /debug/pprof/trace", s.handlePprofTrace)
	}

	s.srv = &http.Server{Handler: mux}
	go func() {
		_ = s.srv.Serve(s.listener)
		close(s.serveDone)
	}()
	return nil
}

// Shutdown gracefully shuts down the admin server. It closes the listener so
// Serve returns, then waits for the serve goroutine to exit.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	if s.serveDone != nil {
		select {
		case <-s.serveDone:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// healthResponse is the JSON for GET /api/health.
type healthResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	UptimeSec   int64  `json:"uptime_sec"`
	Goroutines  int    `json:"goroutines"`
	Workers     int    `json:"workers"`
	Clients     int    `json:"clients"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	workers := s.cfg.Workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	resp := healthResponse{
		Status:     "ok",
		Version:    config.Version,
		UptimeSec:  int64(time.Since(s.startTime).Seconds()),
		Goroutines: runtime.NumGoroutine(),
		Workers:    workers,
		Clients:    s.registry.Count(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// statsResponse is the full JSON for GET /api/stats.
type statsResponse struct {
	Timestamp            int64          `json:"timestamp"`
	Version              string         `json:"version"`
	RegisteredClients    int            `json:"registered_clients"`
	MaxRegistered        int            `json:"max_registered"`
	MaxActiveProxyPerNode int            `json:"max_active_proxy_per_node"`
	ScaleNote            string         `json:"scale_note"`
	Memory               statsMemory    `json:"memory"`
	Clients              []statsClient  `json:"clients"`
}

type statsMemory struct {
	AllocMB float64 `json:"alloc_mb"`
	SysMB   float64 `json:"sys_mb"`
	GCCycles uint32 `json:"gc_cycles"`
}

type statsClient struct {
	ID           int64  `json:"id"`
	Domain       string `json:"domain"`
	PublicURL    string `json:"public_url"`
	ConnectedSec int64  `json:"connected_sec"`
	Requests     int64  `json:"requests"`
	BytesIn      int64  `json:"bytes_in"`
	BytesOut     int64  `json:"bytes_out"`
}

func (s *Server) handleStats(w http.ResponseWriter, _ *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	clients := s.registry.List()
	clientStats := make([]statsClient, 0, len(clients))
	for _, c := range clients {
		connectedSec := int64(0)
		if !c.ConnectedAt.IsZero() {
			connectedSec = int64(time.Since(c.ConnectedAt).Seconds())
		}
		clientStats = append(clientStats, statsClient{
			ID:           c.ID,
			Domain:       c.Domain,
			PublicURL:    s.cfg.PublicURL(c.Domain),
			ConnectedSec: connectedSec,
			Requests:     c.Requests.Load(),
			BytesIn:      c.BytesIn.Load(),
			BytesOut:     c.BytesOut.Load(),
		})
	}

	resp := statsResponse{
		Timestamp:             time.Now().Unix(),
		Version:               config.Version,
		RegisteredClients:     len(clients),
		MaxRegistered:         config.MaxRegisteredClients,
		MaxActiveProxyPerNode: config.MaxActiveProxyReqs,
		ScaleNote:             "shard horizontally beyond 50k concurrent proxy requests",
		Memory: statsMemory{
			AllocMB:  float64(mem.Alloc) / (1024 * 1024),
			SysMB:    float64(mem.Sys) / (1024 * 1024),
			GCCycles: mem.NumGC,
		},
		Clients: clientStats,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleClientsList(w http.ResponseWriter, _ *http.Request) {
	clients := s.registry.List()
	out := make([]statsClient, 0, len(clients))
	for _, c := range clients {
		connectedSec := int64(0)
		if !c.ConnectedAt.IsZero() {
			connectedSec = int64(time.Since(c.ConnectedAt).Seconds())
		}
		out = append(out, statsClient{
			ID:           c.ID,
			Domain:       c.Domain,
			PublicURL:    s.cfg.PublicURL(c.Domain),
			ConnectedSec: connectedSec,
			Requests:     c.Requests.Load(),
			BytesIn:      c.BytesIn.Load(),
			BytesOut:     c.BytesOut.Load(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	if domain == "" {
		http.Error(w, "missing domain", http.StatusBadRequest)
		return
	}
	client, ok := s.registry.Find(domain)
	if !ok {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	s.registry.Unregister(client.ID)
	w.WriteHeader(http.StatusNoContent)
}

// pprof handlers (only registered when PProfEnabled).
func (s *Server) handlePprofIndex(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}

func (s *Server) handlePprofCmdline(w http.ResponseWriter, r *http.Request) {
	pprof.Cmdline(w, r)
}

func (s *Server) handlePprofProfile(w http.ResponseWriter, r *http.Request) {
	pprof.Profile(w, r)
}

func (s *Server) handlePprofSymbol(w http.ResponseWriter, r *http.Request) {
	pprof.Symbol(w, r)
}

func (s *Server) handlePprofTrace(w http.ResponseWriter, r *http.Request) {
	pprof.Trace(w, r)
}
