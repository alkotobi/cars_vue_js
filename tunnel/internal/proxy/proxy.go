package proxy

import (
	"context"
	"fmt"
	"io"
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

// WSUpgrader is the interface for handling WebSocket upgrade requests.
// When set on Proxy, Upgrade: websocket requests are delegated to it.
type WSUpgrader interface {
	HandleUpgrade(w http.ResponseWriter, r *http.Request, client *registry.Client, backendPath string)
}

// clientGate enforces per-client concurrency limit. mu guards count.
type clientGate struct {
	mu    sync.Mutex
	count int
	max   int
}

// pendingRequest holds the response channel and context for one proxied request.
// respCh is buffered so DeliverFrame rarely blocks; size allows ~4 MiB response (128 body frames) to queue.
const respChBuffer = 256

type pendingRequest struct {
	id     uint64
	respCh chan frameDelivery
	ctx    context.Context
}

type frameDelivery struct {
	h       frame.Header
	payload []byte
}

// shard holds a segment of the pending map to reduce lock contention.
type shard struct {
	mu      sync.Mutex
	entries map[uint64]*pendingRequest
}

// pendingMap is 256 shards keyed by requestID % 256.
type pendingMap struct {
	shards [numShards]shard
}

func (m *pendingMap) shard(id uint64) *shard {
	return &m.shards[id%numShards]
}

func (m *pendingMap) get(id uint64) *pendingRequest {
	s := m.shard(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries[id]
}

func (m *pendingMap) set(id uint64, pr *pendingRequest) {
	s := m.shard(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries == nil {
		s.entries = make(map[uint64]*pendingRequest)
	}
	s.entries[id] = pr
}

func (m *pendingMap) delete(id uint64) {
	s := m.shard(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, id)
}

// Proxy forwards HTTP requests through the tunnel and streams responses.
// pending is sharded by requestID; sem is the global semaphore.
type Proxy struct {
	cfg        config.Config
	registry   registry.Registrar
	log        *log.Logger
	ws         WSUpgrader
	pending    pendingMap
	sem        chan struct{}
	clientGates sync.Map // clientID (int64) -> *clientGate
	nextID     atomic.Uint64
}

// New returns a Proxy with the given config, registry, and logger.
func New(cfg config.Config, r registry.Registrar, l *log.Logger) *Proxy {
	return &Proxy{
		cfg:      cfg,
		registry: r,
		log:      l,
		sem:      make(chan struct{}, config.MaxPendingGlobal),
	}
}

// NewWithSemSize returns a Proxy with a custom global semaphore size (for tests).
func NewWithSemSize(cfg config.Config, r registry.Registrar, l *log.Logger, semSize int) *Proxy {
	return &Proxy{
		cfg:      cfg,
		registry: r,
		log:      l,
		sem:      make(chan struct{}, semSize),
	}
}

// SetWSUpgrader sets the WebSocket upgrader; when nil, Upgrade: websocket requests get 501.
func (p *Proxy) SetWSUpgrader(ws WSUpgrader) {
	p.ws = ws
}

// ServeHTTP implements http.Handler. It is the entry point for every browser request.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Step 1: Extract tunnel ID from Host header.
	tunnelID, ok := p.cfg.SubdomainFrom(r.Host)
	if !ok {
		if p.cfg.IsBaseDomain(r.Host) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"status":"ok","service":"tunnel","version":"%s"}`, config.Version)
			return
		}
		http.Error(w, "unknown host", http.StatusNotFound)
		return
	}

	// Step 2: WebSocket upgrade check.
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		if p.ws == nil {
			http.Error(w, "WebSocket not configured", http.StatusNotImplemented)
			return
		}
		client, ok := p.registry.Find(tunnelID)
		if !ok || !client.Connected.Load() {
			http.Error(w, "tunnel not found or offline", http.StatusNotFound)
			return
		}
		p.ws.HandleUpgrade(w, r, client, r.URL.Path)
		return
	}

	// Step 3: Find client.
	client, ok := p.registry.Find(tunnelID)
	if !ok {
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}
	if !client.Connected.Load() {
		http.Error(w, "client offline", http.StatusBadGateway)
		return
	}

	// Step 4: Acquire semaphores (backpressure).
	gateI, _ := p.clientGates.LoadOrStore(client.ID, &clientGate{max: config.MaxPendingPerClient})
	gate := gateI.(*clientGate)
	gate.mu.Lock()
	if gate.count >= gate.max {
		gate.mu.Unlock()
		http.Error(w, "too many requests for tunnel", http.StatusServiceUnavailable)
		return
	}
	gate.count++
	gate.mu.Unlock()
	defer func() {
		gate.mu.Lock()
		gate.count--
		gate.mu.Unlock()
	}()

	select {
	case p.sem <- struct{}{}:
		defer func() { <-p.sem }()
	default:
		http.Error(w, "server busy", http.StatusServiceUnavailable)
		return
	}

	// Generate request ID and create pending entry before sending so DeliverFrame can find it.
	id := p.nextID.Add(1)
	pr := &pendingRequest{
		id:     id,
		respCh: make(chan frameDelivery, respChBuffer),
		ctx:    r.Context(),
	}
	p.pending.set(id, pr)
	defer p.pending.delete(id)

	// Step 5: Build request frame and send FrameHTTPReqStart.
	sp := frame.StartPayload{
		Method: r.Method,
		Path:   r.URL.RequestURI(),
	}
	for k, vv := range r.Header {
		for _, v := range vv {
			sp.Headers = append(sp.Headers, frame.HeaderField{Key: k, Value: v})
		}
	}
	if r.RemoteAddr != "" {
		sp.Headers = append(sp.Headers, frame.HeaderField{Key: "X-Forwarded-For", Value: r.RemoteAddr})
	}
	if r.Host != "" {
		sp.Headers = append(sp.Headers, frame.HeaderField{Key: "X-Forwarded-Host", Value: r.Host})
	}
	if p.cfg.PublicScheme != "" {
		sp.Headers = append(sp.Headers, frame.HeaderField{Key: "X-Forwarded-Proto", Value: p.cfg.PublicScheme})
	}
	startPayload, err := frame.MarshalStart(sp)
	if err != nil {
		http.Error(w, "failed to build request", http.StatusInternalServerError)
		return
	}
	h := frame.Header{Type: frame.FrameHTTPReqStart, RequestID: id, PayloadLen: uint32(len(startPayload))}
	if err := client.Send(h, startPayload); err != nil {
		http.Error(w, "tunnel send failed", http.StatusBadGateway)
		return
	}

	// Step 6: Stream request body in CopyBufferSize chunks, then FrameHTTPReqEnd.
	buf := copyBufferPool.Get().([]byte)
	defer copyBufferPool.Put(buf)
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			h := frame.Header{Type: frame.FrameHTTPReqBody, RequestID: id, PayloadLen: uint32(n)}
			if sendErr := client.Send(h, buf[:n]); sendErr != nil {
				return
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
	}
	endH := frame.Header{Type: frame.FrameHTTPReqEnd, RequestID: id, PayloadLen: 0}
	if err := client.Send(endH, nil); err != nil {
		return
	}

	// Step 7 & 8 & 9: Wait for response frames; handle START, BODY, END.
	timeout := time.NewTimer(config.ProxyRequestTimeout)
	defer timeout.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout.C:
			http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
			return
		case d := <-pr.respCh:
			switch d.h.Type {
			case frame.FrameHTTPRespStart:
				sp, err := frame.UnmarshalStart(d.payload)
				if err != nil {
					return
				}
				for _, hf := range sp.Headers {
					w.Header().Set(hf.Key, hf.Value)
				}
				w.WriteHeader(sp.Status)
				if fl, ok := w.(http.Flusher); ok {
					fl.Flush()
				}
			case frame.FrameHTTPRespBody:
				if _, err := w.Write(d.payload); err != nil {
					return
				}
				if fl, ok := w.(http.Flusher); ok {
					fl.Flush()
				}
			case frame.FrameHTTPRespEnd:
				return
			}
		}
	}
}

var copyBufferPool = sync.Pool{
	New: func() any { return make([]byte, config.CopyBufferSize) },
}

// DeliverFrame implements Dispatcher. It is called from control.Handler with response frames.
func (p *Proxy) DeliverFrame(ctx context.Context, h frame.Header, payload []byte) error {
	pr := p.pending.get(h.RequestID)
	if pr == nil {
		return nil
	}
	// Copy payload so caller can reuse buffer.
	payloadCopy := append([]byte(nil), payload...)
	delivery := frameDelivery{h: h, payload: payloadCopy}
	select {
	case pr.respCh <- delivery:
	default:
		// Consumer slow or done; drop frame to avoid blocking DeliverFrame.
	}
	return nil
}
