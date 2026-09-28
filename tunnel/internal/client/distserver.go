package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yourname/tunnel/internal/log"
)

// DistServer serves a static dist/ directory over HTTP on 127.0.0.1 with an
// SPA fallback to index.html. It is used by the tunnel client when the user
// passes a local dist directory instead of a running HTTP server.
type DistServer struct {
	dir      string
	listener net.Listener
	srv      *http.Server
	port     int
	log      *log.Logger
}

// NewDistServer creates an HTTP file server for dir. It binds to a random
// available port on 127.0.0.1 but does not start serving until Start is called.
func NewDistServer(dir string, l *log.Logger) (*DistServer, error) {
	if dir == "" {
		return nil, fmt.Errorf("distserver: dir is required")
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("distserver: stat %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("distserver: %q is not a directory", dir)
	}
	if l == nil {
		l = log.New(os.Stdout, 0)
	}

	fs := http.FileServer(http.Dir(dir))
	mux := http.NewServeMux()
	// SPA-aware handler: serve static files when present, otherwise index.html.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" || path == "/" {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fullPath := filepath.Join(dir, filepath.Clean(path))
		if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		// Not a file: SPA route → index.html.
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("distserver: listen: %w", err)
	}
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		_ = ln.Close()
		return nil, fmt.Errorf("distserver: unexpected addr type %T", ln.Addr())
	}

	return &DistServer{
		dir:      dir,
		listener: ln,
		srv:      &http.Server{Handler: mux},
		port:     tcpAddr.Port,
		log:      l,
	}, nil
}

// Start begins serving HTTP in a background goroutine.
func (d *DistServer) Start() error {
	if d == nil || d.listener == nil || d.srv == nil {
		return fmt.Errorf("distserver: not initialised")
	}
	go func() {
		if err := d.srv.Serve(d.listener); err != nil && err != http.ErrServerClosed {
			d.log.Error("distserver serve", "err", err)
		}
	}()
	return nil
}

// URL returns the base URL for the dist server, e.g. "http://127.0.0.1:12345".
func (d *DistServer) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", d.port)
}

// Shutdown stops the server gracefully.
func (d *DistServer) Shutdown(ctx context.Context) error {
	if d == nil || d.srv == nil {
		return nil
	}
	return d.srv.Shutdown(ctx)
}
