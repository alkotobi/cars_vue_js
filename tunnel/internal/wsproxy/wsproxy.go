// Package wsproxy tunnels WebSocket connections through the binary frame protocol.
// A WebSocket upgrade request from the browser is forwarded to the tunnel client
// as a normal FrameHTTPReqStart; the 101 response comes back as FrameHTTPRespStart;
// after that all WebSocket frames travel as FrameWSData.
//
// We do not use gorilla/websocket on the server side: we forward raw bytes and do
// not parse WebSocket frames. The tunnel client talks to the local app (e.g. Vite),
// which performs the real WebSocket handshake. We only hijack the browser connection
// and pipe bytes both ways. Idle sessions are closed after config.WSIdleTimeout via
// a watcher goroutine.
package wsproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
	"github.com/yourname/tunnel/internal/registry"
)

const numShards = 256

// toClientBuffer is the size of the per-session channel for frames from tunnel to browser.
const toClientBuffer = 64

// idleCheckInterval is how often the idle watcher runs.
const idleCheckInterval = 60 * time.Second

// sessionShard holds sessions for a subset of request IDs; mu guards entries.
type sessionShard struct {
	mu      sync.Mutex
	entries map[uint64]*wsSession
}

// sessionMap is 256 shards keyed by requestID % numShards to reduce lock contention.
type sessionMap struct {
	shards [numShards]sessionShard
}

func (m *sessionMap) shard(id uint64) *sessionShard {
	return &m.shards[id%numShards]
}

func (m *sessionMap) get(id uint64) *wsSession {
	s := m.shard(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries[id]
}

func (m *sessionMap) set(id uint64, sess *wsSession) {
	s := m.shard(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries == nil {
		s.entries = make(map[uint64]*wsSession)
	}
	s.entries[id] = sess
}

func (m *sessionMap) delete(id uint64) {
	s := m.shard(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, id)
}

// wsSession represents one WebSocket proxy session. clientConn is the hijacked
// browser connection; toClient is written by DeliverFrame and read by tunnelToBrowser.
// lastFrame is updated on each frame for idle timeout. ctx/cancel shut down both goroutines.
type wsSession struct {
	id           uint64
	clientConn   net.Conn
	tunnelClient *registry.Client
	created      time.Time
	lastFrame    atomic.Int64 // Unix nano; updated on each frame from tunnel
	toClient     chan []byte // buffered; frames to send to browser
	ctx          context.Context
	cancel       context.CancelFunc
}

// WSProxy handles WebSocket upgrade and frame forwarding. It implements proxy.WSUpgrader
// (HandleUpgrade) and Dispatcher (DeliverFrame). Sessions are sharded by request ID.
type WSProxy struct {
	log      *log.Logger
	sessions sessionMap
	nextID   atomic.Uint64
}

// New returns a WSProxy with the given logger.
func New(l *log.Logger) *WSProxy {
	return &WSProxy{log: l}
}

// HandleUpgrade implements proxy.WSUpgrader. It hijacks the connection, sends
// the upgrade as FrameHTTPReqStart to the tunnel client, and starts browser↔tunnel pumps.
func (w *WSProxy) HandleUpgrade(rw http.ResponseWriter, r *http.Request, client *registry.Client, backendPath string) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		w.log.Warn("response writer does not support hijack")
		http.Error(rw, "hijack not supported", http.StatusInternalServerError)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		w.log.Warn("hijack failed", "err", err)
		http.Error(rw, "hijack failed", http.StatusInternalServerError)
		return
	}
	// We own the connection; do not return until session ends so caller does not use conn.
	id := w.nextID.Add(1)

	// Build FrameHTTPReqStart with Upgrade and all request headers.
	headers := make([]frame.HeaderField, 0, len(r.Header)+2)
	for k, vv := range r.Header {
		for _, v := range vv {
			headers = append(headers, frame.HeaderField{Key: k, Value: v})
		}
	}
	headers = append(headers, frame.HeaderField{Key: "X-Forwarded-Host", Value: r.Host})
	if r.RemoteAddr != "" {
		headers = append(headers, frame.HeaderField{Key: "X-Forwarded-For", Value: r.RemoteAddr})
	}
	sp := frame.StartPayload{
		Method:  r.Method,
		Path:    backendPath,
		Headers: headers,
	}
	payload, err := frame.MarshalStart(sp)
	if err != nil {
		w.log.Warn("marshal ws req start", "err", err)
		conn.Close()
		return
	}
	h := frame.Header{Type: frame.FrameHTTPReqStart, RequestID: id, PayloadLen: uint32(len(payload))}
	if err := client.Send(h, payload); err != nil {
		w.log.Warn("send ws req start", "err", err)
		conn.Close()
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	sess := &wsSession{
		id:           id,
		clientConn:   conn,
		tunnelClient: client,
		created:      time.Now(),
		toClient:     make(chan []byte, toClientBuffer),
		ctx:          ctx,
		cancel:       cancel,
	}
	sess.lastFrame.Store(time.Now().UnixNano())
	w.sessions.set(id, sess)

	// browserToTunnel: read from browser, send FrameWSData; on EOF send FrameWSClose.
	go w.browserToTunnel(sess)
	// tunnelToBrowser: receive from toClient, write to browser; on close close conn.
	go w.tunnelToBrowser(sess)
	// Idle timeout: if no frame from tunnel for WSIdleTimeout, close session.
	go w.idleWatcher(sess)
}

// browserToTunnel reads raw bytes from the hijacked browser connection in
// CopyBufferSize chunks and sends each as FrameWSData. On EOF it sends FrameWSClose
// and cancels the session.
func (w *WSProxy) browserToTunnel(sess *wsSession) {
	defer sess.cancel()
	buf := copyBufferPool.Get().([]byte)
	defer copyBufferPool.Put(buf)
	for {
		select {
		case <-sess.ctx.Done():
			return
		default:
		}
		if err := sess.clientConn.SetReadDeadline(time.Now().Add(config.ClientIdleTimeout)); err != nil {
			return
		}
		n, err := sess.clientConn.Read(buf)
		if n > 0 {
			h := frame.Header{Type: frame.FrameWSData, RequestID: sess.id, PayloadLen: uint32(n)}
			if sendErr := sess.tunnelClient.Send(h, buf[:n]); sendErr != nil {
				return
			}
		}
		if err == io.EOF {
			closeH := frame.Header{Type: frame.FrameWSClose, RequestID: sess.id, PayloadLen: 0}
			_ = sess.tunnelClient.Send(closeH, nil)
			return
		}
		if err != nil {
			return
		}
	}
}

// idleWatcher runs every idleCheckInterval; if sess.lastFrame is older than
// config.WSIdleTimeout, it cancels the session.
func (w *WSProxy) idleWatcher(sess *wsSession) {
	ticker := time.NewTicker(idleCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-sess.ctx.Done():
			return
		case <-ticker.C:
			last := time.Unix(0, sess.lastFrame.Load())
			if time.Since(last) > config.WSIdleTimeout {
				sess.cancel()
				return
			}
		}
	}
}

// tunnelToBrowser reads payloads from sess.toClient and writes them to the
// browser connection. When toClient is closed, it closes the browser connection.
func (w *WSProxy) tunnelToBrowser(sess *wsSession) {
	defer func() {
		w.sessions.delete(sess.id)
		_ = sess.clientConn.Close()
	}()
	for {
		select {
		case <-sess.ctx.Done():
			return
		case payload, ok := <-sess.toClient:
			if !ok {
				return
			}
			if err := sess.clientConn.SetWriteDeadline(time.Now().Add(config.ProxyRequestTimeout)); err != nil {
				return
			}
			if _, err := sess.clientConn.Write(payload); err != nil {
				return
			}
		}
	}
}

// DeliverFrame implements Dispatcher. It is called from the control handler for
// FrameHTTPRespStart (101 upgrade response), FrameWSData, and FrameWSClose.
func (w *WSProxy) DeliverFrame(ctx context.Context, h frame.Header, payload []byte) error {
	sess := w.sessions.get(h.RequestID)
	if sess == nil {
		return nil
	}
	switch h.Type {
	case frame.FrameHTTPRespStart:
		sp, err := frame.UnmarshalStart(payload)
		if err != nil {
			w.sendErrorAndClose(sess, "invalid response", http.StatusBadGateway)
			return err
		}
		if sp.Status != 101 {
			w.sendErrorAndClose(sess, fmt.Sprintf("upgrade rejected: %d", sp.Status), sp.Status)
			return nil
		}
		// 101 — write raw HTTP response to browser so WebSocket is established.
		if err := w.writeHTTP101(sess, sp); err != nil {
			w.log.Warn("write 101", "err", err)
			sess.cancel()
			return err
		}
		sess.lastFrame.Store(time.Now().UnixNano())
	case frame.FrameWSData:
		sess.lastFrame.Store(time.Now().UnixNano())
		payloadCopy := append([]byte(nil), payload...)
		select {
		case sess.toClient <- payloadCopy:
		default:
			// Consumer slow; drop frame to avoid blocking DeliverFrame.
		}
	case frame.FrameWSClose:
		close(sess.toClient)
		sess.cancel()
	}
	return nil
}

// writeHTTP101 writes the HTTP/1.1 101 Switching Protocols response and headers to the browser.
func (w *WSProxy) writeHTTP101(sess *wsSession, sp frame.StartPayload) error {
	var b strings.Builder
	b.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	for _, hf := range sp.Headers {
		b.WriteString(hf.Key)
		b.WriteString(": ")
		b.WriteString(hf.Value)
		b.WriteString("\r\n")
	}
	b.WriteString("\r\n")
	_, err := sess.clientConn.Write([]byte(b.String()))
	return err
}

// sendErrorAndClose writes a minimal HTTP error response to the browser and closes the session.
func (w *WSProxy) sendErrorAndClose(sess *wsSession, msg string, status int) {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, http.StatusText(status))
	if status < 200 || status >= 300 {
		body := fmt.Sprintf("%d %s", status, msg)
		statusLine += fmt.Sprintf("Content-Length: %d\r\n", len(body))
		statusLine += "\r\n" + body
	} else {
		statusLine += "\r\n"
	}
	_, _ = io.WriteString(sess.clientConn, statusLine)
	sess.cancel()
}

var copyBufferPool = sync.Pool{
	New: func() any { return make([]byte, config.CopyBufferSize) },
}
