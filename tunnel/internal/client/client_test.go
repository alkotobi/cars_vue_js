package client_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/client"
	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/log"
)

// TestRun_ContextCancelReturns verifies Run exits when context is cancelled.
func TestRun_ContextCancelReturns(t *testing.T) {
	cfg := config.ClientDefaults()
	cfg.ServerAddr = "127.0.0.1:19999" // no server listening
	logger := log.New(io.Discard, 0)
	c := client.New(cfg, logger)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = c.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

// TestNew_NoPanic ensures New returns a non-nil client without panicking.
func TestNew_NoPanic(t *testing.T) {
	cfg := config.ClientDefaults()
	cfg.LocalURL = "http://127.0.0.1:5173"
	cfg.Subdomain = "cars"
	logger := log.New(io.Discard, 0)
	c := client.New(cfg, logger)
	if c == nil {
		t.Fatal("New returned nil")
	}
}

