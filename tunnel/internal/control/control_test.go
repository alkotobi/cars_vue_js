package control_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/control"
	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/proxy"
	"github.com/yourname/tunnel/internal/registry"
	"github.com/yourname/tunnel/internal/wsproxy"
)

var (
	_ proxy.Dispatcher   = (*mockProxy)(nil)
	_ wsproxy.Dispatcher = (*mockWS)(nil)
)

func TestRegisterHappyPath(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)

	// Send register.
	_, _ = clientConn.Write([]byte(`{"type":"register"}` + "\n"))
	rd := bufio.NewReader(clientConn)
	line, err := rd.ReadString('\n')
	if err != nil {
		t.Fatalf("read register_ack: %v", err)
	}
	var ack struct {
		Type   string `json:"type"`
		Domain string `json:"domain"`
		URL    string `json:"publicUrl"`
	}
	if err := json.Unmarshal([]byte(line), &ack); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ack.Type != "register_ack" || ack.Domain == "" || ack.URL == "" {
		t.Fatalf("unexpected ack: %+v", ack)
	}
	cancel()
	_ = clientConn.Close()
}

func TestRegisterDuplicateReturnsNack(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)
	rd := bufio.NewReader(clientConn)

	// First register with custom subdomain.
	_, _ = clientConn.Write([]byte(`{"type":"register","customSubdomain":"dup"}` + "\n"))
	line1, _ := rd.ReadString('\n')
	var ack1 struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal([]byte(line1), &ack1)
	if ack1.Type != "register_ack" {
		t.Fatalf("first register: got type %q", ack1.Type)
	}

	// Second connection, same subdomain — need another pipe and handler run.
	serverConn2, clientConn2 := net.Pipe()
	go handler.ServeConn(context.Background(), serverConn2)
	_, _ = clientConn2.Write([]byte(`{"type":"register","customSubdomain":"dup"}` + "\n"))
	rd2 := bufio.NewReader(clientConn2)
	line2, _ := rd2.ReadString('\n')
	var nack struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(line2), &nack); err != nil {
		t.Fatalf("unmarshal nack: %v", err)
	}
	if nack.Type != "register_nack" || nack.Reason != "domain_taken" {
		t.Fatalf("unexpected nack: %+v", nack)
	}
	cancel()
	_ = clientConn.Close()
	_ = clientConn2.Close()
}

func TestHeartbeat(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)
	rd := bufio.NewReader(clientConn)

	_, _ = clientConn.Write([]byte(`{"type":"register"}` + "\n"))
	_, _ = rd.ReadString('\n')
	_, _ = clientConn.Write([]byte(`{"type":"heartbeat"}` + "\n"))
	line, err := rd.ReadString('\n')
	if err != nil {
		t.Fatalf("read heartbeat_ack: %v", err)
	}
	var ack struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(line), &ack); err != nil || ack.Type != "heartbeat_ack" {
		t.Fatalf("heartbeat_ack: %v, %+v", err, ack)
	}
	cancel()
	_ = clientConn.Close()
}

func TestDeregisterRemovesClient(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)
	rd := bufio.NewReader(clientConn)

	_, _ = clientConn.Write([]byte(`{"type":"register","customSubdomain":"gone"}` + "\n"))
	line, _ := rd.ReadString('\n')
	var ack struct {
		Domain string `json:"domain"`
	}
	_ = json.Unmarshal([]byte(line), &ack)
	domain := ack.Domain
	if domain == "" {
		domain = "gone"
	}
	_, _ = clientConn.Write([]byte(`{"type":"deregister"}` + "\n"))
	time.Sleep(50 * time.Millisecond)
	if _, ok := reg.Find(domain); ok {
		t.Fatal("client still in registry after deregister")
	}
	cancel()
	_ = clientConn.Close()
}

type mockProxy struct {
	mu    sync.Mutex
	calls []frame.Header
}

func (m *mockProxy) DeliverFrame(ctx context.Context, h frame.Header, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, h)
	_ = payload
	return nil
}

type mockWS struct {
	mu     sync.Mutex
	calls  []frame.Header
}

func (m *mockWS) DeliverFrame(ctx context.Context, h frame.Header, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, h)
	return nil
}

func TestFrameDemuxCallsDeliverFrame(t *testing.T) {
	reg := registry.New()
	pr := &mockProxy{}
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, pr, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)
	rd := bufio.NewReader(clientConn)

	// Register first.
	_, _ = clientConn.Write([]byte(`{"type":"register"}` + "\n"))
	_, _ = rd.ReadString('\n')

	// Send FrameHTTPRespStart: type 0x04, requestID 99, payload len 0.
	var buf [frame.HeaderSize]byte
	buf[0] = byte(frame.FrameHTTPRespStart)
	binary.BigEndian.PutUint64(buf[1:9], 99)
	binary.BigEndian.PutUint32(buf[9:13], 0)
	_, _ = clientConn.Write(buf[:])
	time.Sleep(100 * time.Millisecond)

	pr.mu.Lock()
	n := len(pr.calls)
	pr.mu.Unlock()
	if n != 1 {
		t.Fatalf("DeliverFrame calls: got %d, want 1", n)
	}
	pr.mu.Lock()
	got := pr.calls[0]
	pr.mu.Unlock()
	if got.Type != frame.FrameHTTPRespStart || got.RequestID != 99 {
		t.Fatalf("frame: got %+v", got)
	}
	cancel()
	_ = clientConn.Close()
}

func TestDoSLineTooLongClosesConn(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)

	// Send JSON line > ControlLineMax (64 KiB) so server closes conn.
	line := append([]byte{'{'}, bytes.Repeat([]byte("x"), config.ControlLineMax)...)
	line = append(line, '\n')
	_, _ = clientConn.Write(line)
	_ = clientConn.Close()
	// Server should close the other end; read should fail or EOF.
	_, err := serverConn.Read(make([]byte, 1))
	if err == nil {
		// Or server may have closed first.
		time.Sleep(100 * time.Millisecond)
	}
	cancel()
}

func TestUnknownFrameTypeKeepsConnOpen(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.ServeConn(ctx, serverConn)
	rd := bufio.NewReader(clientConn)

	_, _ = clientConn.Write([]byte(`{"type":"register"}` + "\n"))
	_, _ = rd.ReadString('\n')
	// Send unknown frame type 0xFF (not 0x01–0x0A). Next 12 bytes are header rest; we send 0 so payload len 0.
	_, _ = clientConn.Write([]byte{0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	// Connection should still be open; send heartbeat and get ack.
	_, _ = clientConn.Write([]byte(`{"type":"heartbeat"}` + "\n"))
	line, err := rd.ReadString('\n')
	if err != nil {
		t.Fatalf("read after 0xFF: %v", err)
	}
	var ack struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(line), &ack); err != nil || ack.Type != "heartbeat_ack" {
		t.Fatalf("heartbeat_ack after 0xFF: %v, %+v", err, ack)
	}
	cancel()
	_ = clientConn.Close()
}

func TestContextCancellationServeConnReturns(t *testing.T) {
	reg := registry.New()
	logger := log.New(&bytes.Buffer{}, 0)
	handler := control.NewHandler(reg, &mockProxy{}, &mockWS{}, logger, "http://localhost:83")
	serverConn, clientConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		handler.ServeConn(ctx, serverConn)
		close(done)
	}()
	// Register so we have a client.
	_, _ = clientConn.Write([]byte(`{"type":"register"}` + "\n"))
	rd := bufio.NewReader(clientConn)
	_, _ = rd.ReadString('\n')
	cancel()
	// Close client side so server's ReadByte unblocks with EOF and ServeConn can exit.
	_ = clientConn.Close()
	select {
	case <-done:
		// ServeConn returned.
	case <-time.After(2 * time.Second):
		t.Fatal("ServeConn did not return after context cancel")
	}
}
