package client

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/log"
)

func TestNewDistServerInvalidDir(t *testing.T) {
	logger := log.New(io.Discard, 0)
	if _, err := NewDistServer("/this/does/not/exist", logger); err == nil {
		t.Fatal("expected error for non-existent dir")
	}
}

func TestDistServerServesFilesAndSPA(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	jsPath := filepath.Join(dir, "main.js")

	if err := os.WriteFile(indexPath, []byte("<html>INDEX</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(jsPath, []byte("console.log('js');"), 0o644); err != nil {
		t.Fatalf("write js: %v", err)
	}

	logger := log.New(io.Discard, 0)
	ds, err := NewDistServer(dir, logger)
	if err != nil {
		t.Fatalf("NewDistServer: %v", err)
	}
	if err := ds.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = ds.Shutdown(ctx)
	}()

	baseURL := ds.URL()

	// Root should serve index.html.
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>INDEX</html>" {
		t.Errorf("index body: got %q", body)
	}

	// JS file should be served with correct MIME type and content.
	resp2, err := http.Get(baseURL + "/main.js")
	if err != nil {
		t.Fatalf("GET /main.js: %v", err)
	}
	defer resp2.Body.Close()
	if ct := resp2.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/javascript") && !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("main.js Content-Type: got %q", ct)
	}
	b2, _ := io.ReadAll(resp2.Body)
	if len(b2) == 0 {
		t.Error("expected non-empty js body")
	}

	// CSS file: ensure correct MIME type (avoids browser "strict MIME" errors).
	cssPath := filepath.Join(dir, "style.css")
	if err := os.WriteFile(cssPath, []byte("body{}"), 0o644); err != nil {
		t.Fatalf("write css: %v", err)
	}
	respCSS, err := http.Get(baseURL + "/style.css")
	if err != nil {
		t.Fatalf("GET /style.css: %v", err)
	}
	defer respCSS.Body.Close()
	if ct := respCSS.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("style.css Content-Type: got %q, want text/css", ct)
	}

	// SPA fallback: unknown route should still return index.html.
	resp3, err := http.Get(baseURL + "/nonexistent/route")
	if err != nil {
		t.Fatalf("GET /nonexistent/route: %v", err)
	}
	defer resp3.Body.Close()
	b3, _ := io.ReadAll(resp3.Body)
	if string(b3) != "<html>INDEX</html>" {
		t.Errorf("SPA fallback: got %q", b3)
	}
}

