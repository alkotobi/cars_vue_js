// Package client implements the tunnel client; client.go contains the control
// loop, connection, and backoff logic.

package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
)

// backoffState implements exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s, then
// cap at 60s (or config.ReconnectMax). Reset() on successful connection.
type backoffState struct {
	current time.Duration
	max     time.Duration
	step    time.Duration
}

func newBackoff(max time.Duration) *backoffState {
	if max <= 0 {
		max = 60 * time.Second
	}
	return &backoffState{
		current: time.Second,
		max:     max,
		step:    time.Second,
	}
}

// Next returns the next backoff duration and advances the state.
func (b *backoffState) Next() time.Duration {
	d := b.current
	if b.current < b.max {
		b.current *= 2
		if b.current > b.max {
			b.current = b.max
		}
	}
	return d
}

// Reset sets the backoff back to the initial duration (call after successful connect).
func (b *backoffState) Reset() {
	b.current = b.step
}

// Outbound JSON (client → server).
type msgRegister struct {
	Type            string `json:"type"`
	CustomSubdomain string `json:"customSubdomain"`
}
type msgHeartbeat struct {
	Type string `json:"type"`
}
type msgDeregister struct {
	Type string `json:"type"`
}

// Inbound JSON (server → client).
type msgRegisterAck struct {
	Type      string `json:"type"`
	Domain    string `json:"domain"`
	PublicURL string `json:"publicUrl"`
}
type msgHeartbeatAck struct {
	Type string `json:"type"`
}

// wsSession is a placeholder for an active WebSocket session (Agent 8 does not
// implement WebSocket forwarding on the client; sessions map is for future use).
type wsSession struct{}

// TunnelClient maintains a persistent control connection to the tunnel server,
// receives proxy request frames, and forwards them via the Forwarder. sem
// limits concurrent requests per client; reconnect is exponential backoff.
type TunnelClient struct {
	cfg       config.Config
	log       *log.Logger
	conn      net.Conn
	fwd       *Forwarder
	sessions  map[uint64]*wsSession
	sem       chan struct{}
	reconnect *backoffState
	writeMu   sync.Mutex
	connMu    sync.Mutex
}

// New returns a TunnelClient for the given config and logger.
func New(cfg config.Config, l *log.Logger) *TunnelClient {
	return &TunnelClient{
		cfg:       cfg,
		log:       l,
		fwd:       NewForwarder(cfg.LocalURL, cfg.Subdomain, l),
		sessions:  make(map[uint64]*wsSession),
		sem:       make(chan struct{}, config.MaxPendingPerClient),
		reconnect: newBackoff(cfg.ReconnectMax),
	}
}

// Run connects to the server and handles messages. It reconnects on failure
// with exponential backoff. Blocks until ctx is cancelled.
func (c *TunnelClient) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		err := c.connect(ctx)
		if err != nil {
			c.log.Warn("connection failed", "err", err)
			wait := c.reconnect.Next()
			c.log.Info("reconnecting", "in", wait)
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil
			}
			continue
		}
		c.reconnect.Reset()
	}
}

// connect dials the server, sends register, handles register_ack, starts
// heartbeat, and runs the demux loop. Returns when conn closes or ctx is cancelled.
func (c *TunnelClient) connect(ctx context.Context) error {
	dialer := net.Dialer{Timeout: config.DialTimeout}
	host := c.cfg.ControlHost
	if host == "" {
		host = c.cfg.ServerAddr
	}
	addr := fmt.Sprintf("%s:%d", host, c.cfg.ControlPort)
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	sendFrame := func(h frame.Header, payload []byte) error {
		c.writeMu.Lock()
		defer c.writeMu.Unlock()
		if err := frame.WriteHeader(conn, h); err != nil {
			return err
		}
		if len(payload) > 0 {
			_, err := conn.Write(payload)
			return err
		}
		return nil
	}
	c.fwd.SetSendFrame(sendFrame)

	if err := c.sendRegister(); err != nil {
		conn.Close()
		return err
	}
	ack, err := c.readRegisterAck(conn)
	if err != nil {
		conn.Close()
		return err
	}
	c.log.Info("✓ Tunnel active", "url", ack.PublicURL)

	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.sendJSON(conn, msgHeartbeat{Type: "heartbeat"}); err != nil {
					return
				}
			}
		}
	}()

	err = c.demux(ctx, conn)
	<-heartbeatDone

	if ctx.Err() != nil {
		c.gracefulShutdown(conn)
	}
	c.connMu.Lock()
	c.conn = nil
	c.connMu.Unlock()
	conn.Close()
	return err
}

func (c *TunnelClient) sendRegister() error {
	msg := msgRegister{Type: "register", CustomSubdomain: c.cfg.Subdomain}
	return c.sendJSON(c.conn, msg)
}

func (c *TunnelClient) sendJSON(conn net.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(data, '\n'))
	return err
}

func (c *TunnelClient) readRegisterAck(conn net.Conn) (msgRegisterAck, error) {
	var ack msgRegisterAck
	br := bufio.NewReaderSize(conn, 4096)
	line, err := br.ReadString('\n')
	if err != nil {
		return ack, err
	}
	line = strings.TrimSuffix(line, "\n")
	if err := json.Unmarshal([]byte(line), &ack); err != nil {
		return ack, err
	}
	if ack.Type != "register_ack" {
		return ack, fmt.Errorf("expected register_ack, got %q", ack.Type)
	}
	return ack, nil
}

// demux runs the main read loop: JSON lines (leading '{') or binary frames.
func (c *TunnelClient) demux(ctx context.Context, conn net.Conn) error {
	br := bufio.NewReaderSize(conn, 32*1024)
	payloadBuf := make([]byte, 0, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		b, err := br.ReadByte()
		if err != nil {
			return err
		}
		if b == '{' {
			_ = br.UnreadByte()
			line, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			line = strings.TrimSuffix(line, "\n")
			if len(line) > config.ControlLineMax {
				c.log.Warn("control line too long")
				return io.EOF
			}
			c.handleControlLine(line)
			continue
		}
		if b >= 0x01 && b <= 0x0A {
			_ = br.UnreadByte()
			header, err := frame.ReadHeader(br)
			if err != nil {
				return err
			}
			if cap(payloadBuf) < int(header.PayloadLen) {
				payloadBuf = make([]byte, header.PayloadLen)
			} else {
				payloadBuf = payloadBuf[:header.PayloadLen]
			}
			if header.PayloadLen > 0 {
				if _, err := io.ReadFull(br, payloadBuf); err != nil {
					return err
				}
			}
			c.dispatchFrame(ctx, header, payloadBuf)
			continue
		}
		c.log.Warn("unknown frame type, ignoring", "byte", b)
	}
}

func (c *TunnelClient) handleControlLine(line string) {
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return
	}
	switch raw.Type {
	case "register_ack", "heartbeat_ack":
		// Already handled or no-op
	default:
		// Ignore
	}
}

// dispatchFrame routes a binary frame to the forwarder or handles control frames.
func (c *TunnelClient) dispatchFrame(ctx context.Context, h frame.Header, payload []byte) {
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)
	switch h.Type {
	case frame.FrameHTTPReqStart:
		select {
		case c.sem <- struct{}{}:
			go func() {
				defer func() { <-c.sem }()
				c.fwd.HandleRequest(ctx, h.RequestID, payloadCopy)
			}()
		default:
			c.log.Warn("too many pending requests, dropping")
		}
	case frame.FrameHTTPReqBody:
		c.fwd.DeliverBody(h.RequestID, payloadCopy)
	case frame.FrameHTTPReqEnd:
		c.fwd.DeliverBodyEnd(h.RequestID)
	case frame.FramePing:
		_ = c.writeFrame(frame.Header{Type: frame.FramePong, RequestID: h.RequestID, PayloadLen: 0}, nil)
	default:
		c.log.Warn("unknown frame type, ignoring", "type", h.Type)
	}
}

func (c *TunnelClient) writeFrame(h frame.Header, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()
	if conn == nil {
		return io.ErrClosedPipe
	}
	if err := frame.WriteHeader(conn, h); err != nil {
		return err
	}
	if len(payload) > 0 {
		_, err := conn.Write(payload)
		return err
	}
	return nil
}

// gracefulShutdown sends deregister and waits up to 500ms before closing.
func (c *TunnelClient) gracefulShutdown(conn net.Conn) {
	_ = c.sendJSON(conn, msgDeregister{Type: "deregister"})
	<-time.After(500 * time.Millisecond)
}
