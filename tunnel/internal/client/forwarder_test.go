package client_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/client"
	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
)

func TestForwarderHandleRequest_Success(t *testing.T) {
	// Local server echoes path and returns 200 with body.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/foo" {
			t.Errorf("path: got %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello" {
			t.Errorf("body: got %q", body)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	var framesMu sync.Mutex
	var frames []struct {
		h frame.Header
		p []byte
	}
	send := func(h frame.Header, p []byte) error {
		pCopy := make([]byte, len(p))
		copy(pCopy, p)
		framesMu.Lock()
		frames = append(frames, struct {
			h frame.Header
			p []byte
		}{h, pCopy})
		framesMu.Unlock()
		return nil
	}

	fwd := client.NewForwarder(ts.URL, "", log.New(io.Discard, 0))
	fwd.SetSendFrame(send)

	payload, _ := frame.MarshalStart(frame.StartPayload{
		Method: "POST",
		Path:   "/api/foo",
		Headers: []frame.HeaderField{
			{Key: "Content-Type", Value: "application/octet-stream"},
		},
	})
	ctx := context.Background()
	fwd.HandleRequest(ctx, 1, payload)
	// Allow doAndStream goroutine to start and block on body read.
	time.Sleep(20 * time.Millisecond)
	fwd.DeliverBody(1, []byte("hello"))
	fwd.DeliverBodyEnd(1)

	// Wait for response frames (HandleRequest runs in goroutine).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		framesMu.Lock()
		n := len(frames)
		framesMu.Unlock()
		if n >= 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	framesMu.Lock()
	defer framesMu.Unlock()
	if len(frames) < 3 {
		t.Fatalf("expected at least 3 frames, got %d", len(frames))
	}
	if frames[0].h.Type != frame.FrameHTTPRespStart {
		t.Errorf("first frame: got type %v", frames[0].h.Type)
	}
	sp, _ := frame.UnmarshalStart(frames[0].p)
	if sp.Status != 200 {
		t.Errorf("status: got %d", sp.Status)
	}
	if frames[1].h.Type != frame.FrameHTTPRespBody {
		t.Errorf("second frame: got type %v", frames[1].h.Type)
	}
	if string(frames[1].p) != "ok" {
		t.Errorf("body: got %q", frames[1].p)
	}
	if frames[2].h.Type != frame.FrameHTTPRespEnd {
		t.Errorf("third frame: got type %v", frames[2].h.Type)
	}
}

func TestForwarderHandleRequest_PathStripSubdomain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bar" {
			t.Errorf("path: got %q (expected /api/bar)", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()

	gotStart := make(chan []byte, 1)
	send := func(h frame.Header, p []byte) error {
		if h.Type == frame.FrameHTTPRespStart {
			select {
			case gotStart <- p:
			default:
			}
		}
		return nil
	}
	fwd := client.NewForwarder(ts.URL, "mysub", log.New(io.Discard, 0))
	fwd.SetSendFrame(send)

	// Path already has subdomain prefix; forwarder should strip to /api/bar.
	payload, _ := frame.MarshalStart(frame.StartPayload{
		Method: "GET",
		Path:   "/mysub/api/bar",
	})
	fwd.HandleRequest(context.Background(), 2, payload)
	fwd.DeliverBodyEnd(2)

	select {
	case startPayload := <-gotStart:
		sp, _ := frame.UnmarshalStart(startPayload)
		if sp.Status != 204 {
			t.Errorf("status: got %d", sp.Status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no response received")
	}
}

func TestForwarderHandleRequest_InvalidPayload(t *testing.T) {
	statusCh := make(chan int, 1)
	send := func(h frame.Header, p []byte) error {
		if h.Type == frame.FrameHTTPRespStart {
			sp, _ := frame.UnmarshalStart(p)
			select {
			case statusCh <- sp.Status:
			default:
			}
		}
		return nil
	}
	fwd := client.NewForwarder("http://localhost:9999", "", log.New(io.Discard, 0))
	fwd.SetSendFrame(send)

	// Invalid payload (too short) → 502
	fwd.HandleRequest(context.Background(), 3, []byte{0x01})
	select {
	case status := <-statusCh:
		if status != 502 {
			t.Errorf("expected 502, got %d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no response received")
	}
}

func TestForwarderDeliverBody_NoActiveRequest(t *testing.T) {
	fwd := client.NewForwarder("http://localhost:9999", "", log.New(io.Discard, 0))
	// Must not panic.
	fwd.DeliverBody(999, []byte("x"))
	fwd.DeliverBodyEnd(999)
}

func TestForwarderDoubleSlashPrevention(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// localURL has no trailing slash; path is "foo". Expect /foo not //foo.
		if len(r.URL.Path) < 1 || r.URL.Path[0] != '/' {
			t.Errorf("path should start with single slash: %q", r.URL.Path)
		}
		if bytes.Contains([]byte(r.URL.Path), []byte("//")) {
			t.Errorf("path should not contain double slash: %q", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	send := func(h frame.Header, p []byte) error { return nil }
	fwd := client.NewForwarder(ts.URL, "", log.New(io.Discard, 0))
	fwd.SetSendFrame(send)
	payload, _ := frame.MarshalStart(frame.StartPayload{Method: "GET", Path: "/foo"})
	fwd.HandleRequest(context.Background(), 4, payload)
	fwd.DeliverBodyEnd(4)
}
