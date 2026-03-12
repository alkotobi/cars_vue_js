package control

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/proxy"
	"github.com/yourname/tunnel/internal/registry"
	"github.com/yourname/tunnel/internal/wsproxy"
)

// payloadPool holds buffers for reading frame payloads; returned after dispatch.
var payloadPool = sync.Pool{
	New: func() any { return make([]byte, 0, 64*1024) },
}

// Inbound JSON (client → server).
type msgRegister struct {
	Type             string `json:"type"`
	CustomSubdomain  string `json:"customSubdomain"`
}
type msgHeartbeat struct {
	Type string `json:"type"`
}
type msgDeregister struct {
	Type string `json:"type"`
}

// Outbound JSON (server → client).
type msgRegisterAck struct {
	Type      string `json:"type"`
	Domain    string `json:"domain"`
	PublicURL string `json:"publicUrl"`
}
type msgRegisterNack struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}
type msgHeartbeatAck struct {
	Type string `json:"type"`
}
type msgPing struct {
	Type string `json:"type"`
}

// Handler holds dependencies for handling one client control connection.
type Handler struct {
	registry      registry.Registrar
	proxy         proxy.Dispatcher
	wsproxy       wsproxy.Dispatcher
	log           *log.Logger
	publicBaseURL string
}

// NewHandler returns a Handler for the given dependencies. publicBaseURL is
// the base URL for register_ack (e.g. "http://163.245.222.142:83").
func NewHandler(r registry.Registrar, p proxy.Dispatcher, w wsproxy.Dispatcher, l *log.Logger, publicBaseURL string) *Handler {
	return &Handler{
		registry:      r,
		proxy:         p,
		wsproxy:       w,
		log:           l,
		publicBaseURL: publicBaseURL,
	}
}

// ServeConn handles one tunnel client connection. It blocks until the
// connection closes. Call from a goroutine: go handler.ServeConn(ctx, conn).
func (h *Handler) ServeConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	br := bufio.NewReaderSize(conn, 32*1024)
	var client *registry.Client
	var lastPong atomic.Int64
	lastPong.Store(time.Now().UnixNano())

	// Ping sender: every 30s send JSON ping.
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := sendJSON(conn, msgPing{Type: "ping"}); err != nil {
					return
				}
			}
		}
	}()

	// Stale watcher: if no pong within HeartbeatTimeout, close conn.
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-config.HeartbeatTimeout).UnixNano()
				if lastPong.Load() < cutoff {
					conn.Close()
					return
				}
			}
		}
	}()

	// Demux loop: JSON lines (leading '{') or binary frames (0x01–0x0A).
	for {
		select {
		case <-ctx.Done():
			goto done
		default:
		}
		b, err := br.ReadByte()
		if err != nil {
			goto done
		}
		if b == '{' {
			_ = br.UnreadByte()
			line, err := br.ReadString('\n')
			if err != nil {
				goto done
			}
			if len(line) > config.ControlLineMax {
				h.log.Warn("control line too long, closing conn", "len", len(line))
				goto done
			}
			if err := h.handleControlMessage(conn, line, &client, &lastPong); err != nil {
				goto done
			}
			continue
		}
		if b >= 0x01 && b <= 0x0A {
			_ = br.UnreadByte()
			header, err := frame.ReadHeader(br)
			if err != nil {
				goto done
			}
			payload := payloadPool.Get().([]byte)
			if cap(payload) < int(header.PayloadLen) {
				payload = make([]byte, header.PayloadLen)
			} else {
				payload = payload[:header.PayloadLen]
			}
			if header.PayloadLen > 0 {
				if _, err := io.ReadFull(br, payload); err != nil {
					payload = payload[:0]
					payloadPool.Put(payload)
					goto done
				}
			}
			h.frameDispatch(ctx, client, header, payload)
			payload = payload[:0]
			payloadPool.Put(payload)
			continue
		}
		// Unknown frame type: log and continue for forward compatibility.
		h.log.Warn("unknown frame type, ignoring", "byte", b)
		continue
	}
done:
	if client != nil {
		h.registry.Unregister(client.ID)
	}
	cancel()
	<-pingDone
	<-watchDone
}

func (h *Handler) handleControlMessage(conn net.Conn, line string, client **registry.Client, lastPong *atomic.Int64) error {
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return err
	}
	switch raw.Type {
	case "register":
		var m msgRegister
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return err
		}
		domain := m.CustomSubdomain
		if domain == "" {
			domain = registry.GenerateDomain()
		}
		c := &registry.Client{
			Domain: domain,
			Send: func(header frame.Header, payload []byte) error {
				if (*client) != nil && !(*client).Connected.Load() {
					return registry.ErrClientGone
				}
				return sendFrame(conn, header, payload)
			},
		}
		if err := h.registry.Register(c); err != nil {
			var reason string
			switch err {
			case registry.ErrDomainTaken:
				reason = "domain_taken"
			case registry.ErrInvalidDomain:
				reason = "invalid_domain"
			case registry.ErrAtCapacity:
				reason = "at_capacity"
			default:
				reason = "invalid_domain"
			}
			return sendJSON(conn, msgRegisterNack{Type: "register_nack", Reason: reason})
		}
		*client = c
		publicURL := h.publicBaseURL + "/" + c.Domain + "/"
		return sendJSON(conn, msgRegisterAck{Type: "register_ack", Domain: c.Domain, PublicURL: publicURL})
	case "heartbeat":
		lastPong.Store(time.Now().UnixNano())
		return sendJSON(conn, msgHeartbeatAck{Type: "heartbeat_ack"})
	case "deregister":
		if *client != nil {
			h.registry.Unregister((*client).ID)
			*client = nil
		}
		return nil
	default:
		return nil
	}
}

func (h *Handler) frameDispatch(ctx context.Context, client *registry.Client, header frame.Header, payload []byte) {
	switch header.Type {
	case frame.FrameHTTPRespStart, frame.FrameHTTPRespBody, frame.FrameHTTPRespEnd:
		// FrameHTTPRespStart may be 101 for WebSocket; wsproxy no-ops if session not found.
		_ = h.wsproxy.DeliverFrame(ctx, header, payload)
		_ = h.proxy.DeliverFrame(ctx, header, payload)
	case frame.FrameWSData, frame.FrameWSClose:
		_ = h.wsproxy.DeliverFrame(ctx, header, payload)
	case frame.FramePong:
		if client != nil {
			client.LastHeartbeat.Store(time.Now().UnixNano())
		}
	default:
		h.log.Warn("unknown frame type", "type", header.Type)
	}
}

func sendJSON(conn net.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(data, '\n'))
	return err
}

func sendFrame(conn net.Conn, h frame.Header, payload []byte) error {
	if err := frame.WriteHeader(conn, h); err != nil {
		return err
	}
	if len(payload) > 0 {
		_, err := conn.Write(payload)
		return err
	}
	return nil
}
