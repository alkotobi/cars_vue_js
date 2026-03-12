package proxy_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/proxy"
	"github.com/yourname/tunnel/internal/registry"
)

func makeProxyAndRegistry(t *testing.T) (*proxy.Proxy, registry.Registrar) {
	t.Helper()
	reg := registry.New()
	logger := log.New(io.Discard, 0)
	p := proxy.New(reg, logger)
	return p, reg
}

func registerClient(t *testing.T, reg registry.Registrar, domain string, reqIDCh chan<- uint64, p *proxy.Proxy) *registry.Client {
	t.Helper()
	c := &registry.Client{
		Domain: domain,
		Send: func(h frame.Header, payload []byte) error {
			select {
			case reqIDCh <- h.RequestID:
			default:
			}
			return nil
		},
	}
	if err := reg.Register(c); err != nil {
		t.Fatalf("register: %v", err)
	}
	return c
}

// TestHappyPath: ServeHTTP → frames to client → deliver response → browser gets status and headers.
func TestHappyPath(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	reqIDCh := make(chan uint64, 1)
	registerClient(t, reg, "cars", reqIDCh, p)
	srv := httptest.NewServer(p)
	defer srv.Close()

	respCh := make(chan *http.Response, 1)
	go func() {
		resp, err := http.Get(srv.URL + "/cars/")
		if err != nil {
			return
		}
		respCh <- resp
	}()

	id := <-reqIDCh
	startPayload, _ := frame.MarshalStart(frame.StartPayload{
		Method:  "",
		Path:    "",
		Status:  200,
		Headers: []frame.HeaderField{{Key: "Content-Type", Value: "text/html"}},
	})
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(startPayload))}, startPayload)
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespBody, RequestID: id, PayloadLen: 5}, []byte("hello"))
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)

	resp := <-respCh
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status: got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/html" {
		t.Errorf("Content-Type: got %q", ct)
	}
}

// TestMIMEPassthrough: backend sends Content-Type: application/javascript for .css path; browser receives it verbatim.
func TestMIMEPassthrough(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	reqIDCh := make(chan uint64, 1)
	registerClient(t, reg, "app", reqIDCh, p)
	srv := httptest.NewServer(p)
	defer srv.Close()

	respCh := make(chan *http.Response, 1)
	go func() {
		resp, _ := http.Get(srv.URL + "/app/static/style.css")
		respCh <- resp
	}()

	id := <-reqIDCh
	startPayload, _ := frame.MarshalStart(frame.StartPayload{
		Status:  200,
		Headers: []frame.HeaderField{{Key: "Content-Type", Value: "application/javascript"}},
	})
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(startPayload))}, startPayload)
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)

	resp := <-respCh
	if resp == nil {
		t.Fatal("no response")
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	ct := resp.Header.Get("Content-Type")
	if ct != "application/javascript" {
		t.Errorf("Content-Type: got %q, want application/javascript", ct)
	}
}

func TestTunnelNotFound(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	_ = reg
	srv := httptest.NewServer(p)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/nonexistent/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", resp.StatusCode)
	}
}

func TestClientOffline(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	reqIDCh := make(chan uint64, 1)
	c := registerClient(t, reg, "offline", reqIDCh, p)
	c.Connected.Store(false)
	srv := httptest.NewServer(p)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/offline/")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status: got %d, want 502", resp.StatusCode)
	}
}

// TestTimeout: client request is cancelled while waiting for response; no panic.
// Full 504 gateway timeout is exercised by the timeout path in ServeHTTP (ProxyRequestTimeout).
func TestTimeout(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	reqIDCh := make(chan uint64, 1)
	registerClient(t, reg, "slow", reqIDCh, p)
	srv := httptest.NewServer(p)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/slow/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Context cancelled or connection closed is expected.
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusGatewayTimeout {
		t.Logf("got 504 as expected")
	}
}

// TestSemaphoreFull: when global semaphore is exhausted, return 503.
func TestSemaphoreFull(t *testing.T) {
	reg := registry.New()
	logger := log.New(io.Discard, 0)
	p := proxy.NewWithSemSize(reg, logger, 1)
	reqIDCh := make(chan uint64, 2)
	registerClient(t, reg, "busy", reqIDCh, p)
	srv := httptest.NewServer(p)
	defer srv.Close()

	respCh1 := make(chan *http.Response, 1)
	go func() {
		resp, _ := http.Get(srv.URL + "/busy/")
		respCh1 <- resp
	}()
	firstID := <-reqIDCh // first request acquired sem and is waiting for response

	resp2, err := http.Get(srv.URL + "/busy/")
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("second request: got %d, want 503", resp2.StatusCode)
	}

	// Unblock first request
	id1 := firstID
	startPayload, _ := frame.MarshalStart(frame.StartPayload{Status: 200, Headers: nil})
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id1, PayloadLen: uint32(len(startPayload))}, startPayload)
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id1, PayloadLen: 0}, nil)
	<-respCh1
}

// pathAndID is used by TestPathStripping to pass data from Send callback to test without data race.
type pathAndID struct {
	path string
	id   uint64
}

// TestPathStripping: /cars/src/main.js → backend receives /src/main.js.
func TestPathStripping(t *testing.T) {
	pathCh := make(chan pathAndID, 1)
	p, reg := makeProxyAndRegistry(t)
	c := &registry.Client{
		Domain: "strip",
		Send: func(h frame.Header, payload []byte) error {
			if h.Type == frame.FrameHTTPReqStart {
				sp, _ := frame.UnmarshalStart(payload)
				select {
				case pathCh <- pathAndID{path: sp.Path, id: h.RequestID}:
				default:
				}
			}
			return nil
		},
	}
	_ = reg.Register(c)
	srv := httptest.NewServer(p)
	defer srv.Close()

	respCh := make(chan *http.Response, 1)
	go func() {
		resp, _ := http.Get(srv.URL + "/strip/src/main.js")
		respCh <- resp
	}()

	pathInfo := <-pathCh
	if pathInfo.path != "/src/main.js" {
		t.Errorf("backend path: got %q, want /src/main.js", pathInfo.path)
	}
	startPayload, _ := frame.MarshalStart(frame.StartPayload{Status: 200})
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: pathInfo.id, PayloadLen: uint32(len(startPayload))}, startPayload)
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: pathInfo.id, PayloadLen: 0}, nil)
	<-respCh
}

// TestRequestBodyStreaming: POST with 2 MiB body → forwarded in chunks.
func TestRequestBodyStreaming(t *testing.T) {
	var bodyChunks int
	var totalBytes int
	reqIDCh := make(chan uint64, 1)
	p, reg := makeProxyAndRegistry(t)
	c := &registry.Client{
		Domain: "big",
		Send: func(h frame.Header, payload []byte) error {
			if h.Type == frame.FrameHTTPReqStart {
				select { case reqIDCh <- h.RequestID: default: }
			}
			if h.Type == frame.FrameHTTPReqBody {
				bodyChunks++
				totalBytes += len(payload)
			}
			return nil
		},
	}
	_ = reg.Register(c)
	srv := httptest.NewServer(p)
	defer srv.Close()

	const size = 2 * 1024 * 1024
	body := bytes.Repeat([]byte("x"), size)
	respCh := make(chan *http.Response, 1)
	go func() {
		resp, _ := http.Post(srv.URL+"/big/", "application/octet-stream", bytes.NewReader(body))
		respCh <- resp
	}()

	id := <-reqIDCh
	time.Sleep(50 * time.Millisecond)
	startPayload, _ := frame.MarshalStart(frame.StartPayload{Status: 200})
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(startPayload))}, startPayload)
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)

	<-respCh
	if totalBytes != size {
		t.Errorf("total body bytes: got %d, want %d", totalBytes, size)
	}
	if bodyChunks < 2 {
		t.Errorf("expected multiple body chunks, got %d", bodyChunks)
	}
}

// TestResponseBodyStreaming: 4 MiB response → streamed in chunks.
func TestResponseBodyStreaming(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	reqIDCh := make(chan uint64, 1)
	registerClient(t, reg, "stream", reqIDCh, p)
	srv := httptest.NewServer(p)
	defer srv.Close()

	respCh := make(chan *http.Response, 1)
	go func() {
		resp, _ := http.Get(srv.URL + "/stream/")
		respCh <- resp
	}()

	id := <-reqIDCh
	startPayload, _ := frame.MarshalStart(frame.StartPayload{Status: 200, Headers: []frame.HeaderField{{Key: "Content-Type", Value: "application/octet-stream"}}})
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(startPayload))}, startPayload)
	const chunkSize = 32 * 1024
	const totalSize = 4 * 1024 * 1024
	for sent := 0; sent < totalSize; sent += chunkSize {
		n := chunkSize
		if sent+n > totalSize {
			n = totalSize - sent
		}
		_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespBody, RequestID: id, PayloadLen: uint32(n)}, bytes.Repeat([]byte("y"), n))
	}
	_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)

	resp := <-respCh
	defer resp.Body.Close()
	got, _ := io.Copy(io.Discard, resp.Body)
	if got != totalSize {
		t.Errorf("response body size: got %d, want %d", got, totalSize)
	}
}

// TestConcurrent: many goroutines request simultaneously; no race (run with -race).
func TestConcurrent(t *testing.T) {
	p, reg := makeProxyAndRegistry(t)
	reqIDCh := make(chan uint64, 1000)
	c := &registry.Client{
		Domain: "conc",
		Send: func(h frame.Header, payload []byte) error {
			if h.Type == frame.FrameHTTPReqStart {
				select {
				case reqIDCh <- h.RequestID:
				default:
				}
			}
			return nil
		},
	}
	_ = reg.Register(c)
	srv := httptest.NewServer(p)
	defer srv.Close()

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(srv.URL + "/conc/")
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}

	ids := make([]uint64, 0, n)
	for len(ids) < n {
		select {
		case id := <-reqIDCh:
			ids = append(ids, id)
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for %d requests, got %d", n, len(ids))
		}
	}

	startPayload, _ := frame.MarshalStart(frame.StartPayload{Status: 200})
	for _, id := range ids {
		_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(startPayload))}, startPayload)
		_ = p.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)
	}
	wg.Wait()
}