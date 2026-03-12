package wsproxy_test

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/registry"
	"github.com/yourname/tunnel/internal/wsproxy"
)

// hijackRW implements http.ResponseWriter and http.Hijacker for tests.
type hijackRW struct {
	hdr  http.Header
	conn net.Conn
}

func (h *hijackRW) Header() http.Header {
	if h.hdr == nil {
		h.hdr = make(http.Header)
	}
	return h.hdr
}

func (h *hijackRW) WriteHeader(int) {}

func (h *hijackRW) Write([]byte) (int, error) { return 0, nil }

func (h *hijackRW) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw := bufio.NewReadWriter(bufio.NewReader(h.conn), bufio.NewWriter(h.conn))
	return h.conn, rw, nil
}

// TestDeliverFrameUnknownID verifies DeliverFrame with unknown RequestID is a no-op.
func TestDeliverFrameUnknownID(t *testing.T) {
	w := wsproxy.New(log.New(io.Discard, 0))
	err := w.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameWSData, RequestID: 999, PayloadLen: 2}, []byte("ab"))
	if err != nil {
		t.Errorf("DeliverFrame unknown id: got err %v", err)
	}
}

// TestHandleUpgradeThen101 verifies full upgrade flow: hijack, send ReqStart, deliver 101, browser receives 101.
func TestHandleUpgradeThen101(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sendCh := make(chan frame.Header, 1)
	client := &registry.Client{
		Domain: "ws",
		Send: func(h frame.Header, payload []byte) error {
			sendCh <- h
			return nil
		},
	}
	client.Connected.Store(true)

	logger := log.New(io.Discard, 0)
	w := wsproxy.New(logger)
	rw := &hijackRW{conn: serverConn}
	req, _ := http.NewRequest(http.MethodGet, "http://test/ws/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.RemoteAddr = "1.2.3.4:56"
	req = req.WithContext(context.Background())

	go w.HandleUpgrade(rw, req, client, "/")

	// Wait for ReqStart to be sent and get RequestID.
	var header frame.Header
	select {
	case header = <-sendCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for ReqStart")
	}
	if header.Type != frame.FrameHTTPReqStart {
		t.Fatalf("expected FrameHTTPReqStart, got %v", header.Type)
	}
	id := header.RequestID
	// Allow HandleUpgrade to register session and start goroutines before we deliver.
	time.Sleep(50 * time.Millisecond)

	// Start read in background so pipe Write in writeHTTP101 can complete.
	readDone := make(chan string, 1)
	go func() {
		_ = clientConn.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 512)
		n, err := clientConn.Read(buf)
		if err != nil {
			readDone <- "err:" + err.Error()
			return
		}
		readDone <- string(buf[:n])
	}()

	// Deliver 101 with minimal headers.
	payload, _ := frame.MarshalStart(frame.StartPayload{
		Status:  101,
		Headers: []frame.HeaderField{{Key: "Upgrade", Value: "websocket"}},
	})
	_ = w.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(payload))}, payload)

	got := <-readDone
	if len(got) >= 7 && got[:7] == "err:" {
		t.Fatalf("read from browser conn: %s", got[4:])
	}
	if len(got) < 19 || got[:19] != "HTTP/1.1 101 Switch" {
		t.Errorf("browser did not get 101: %q", got)
	}

	// Close session so goroutines exit.
	_ = w.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameWSClose, RequestID: id, PayloadLen: 0}, nil)
	time.Sleep(50 * time.Millisecond)
}

// TestDeliverFrameRejectNon101 verifies that status != 101 sends error to browser and closes session.
func TestDeliverFrameRejectNon101(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sendCh := make(chan frame.Header, 1)
	client := &registry.Client{
		Domain: "ws",
		Send: func(h frame.Header, payload []byte) error {
			sendCh <- h
			return nil
		},
	}
	client.Connected.Store(true)

	w := wsproxy.New(log.New(io.Discard, 0))
	rw := &hijackRW{conn: serverConn}
	req, _ := http.NewRequest(http.MethodGet, "http://test/", nil)
	req.Header.Set("Upgrade", "websocket")
	req = req.WithContext(context.Background())

	go w.HandleUpgrade(rw, req, client, "/")
	header := <-sendCh
	id := header.RequestID
	time.Sleep(50 * time.Millisecond)

	readDone := make(chan string, 1)
	go func() {
		_ = clientConn.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 256)
		n, err := clientConn.Read(buf)
		if err != nil {
			readDone <- "err:" + err.Error()
			return
		}
		readDone <- string(buf[:n])
	}()

	// Deliver 403; session should send 403 to browser and close.
	payload, _ := frame.MarshalStart(frame.StartPayload{Status: 403})
	_ = w.DeliverFrame(context.Background(), frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(payload))}, payload)

	got := <-readDone
	if len(got) >= 7 && got[:7] == "err:" {
		t.Fatalf("read: %s", got[4:])
	}
	if len(got) < 15 || got[:15] != "HTTP/1.1 403 Fo" {
		t.Errorf("expected 403 response, got %q", got)
	}
}
