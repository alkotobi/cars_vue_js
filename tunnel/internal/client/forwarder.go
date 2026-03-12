// Package client implements the tunnel client; forwarder.go provides HTTP forwarding
// of proxy request frames to the local application.

package client

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/log"
)

// sendFrameFunc sends one binary frame to the tunnel server. Set by the client
// after the control connection is established. Calls are serialized by the client.
type sendFrameFunc func(h frame.Header, payload []byte) error

// activeReq holds state for one in-flight request: the pipe write end feeds the
// request body to http.Client; done is closed when the response has been fully sent.
type activeReq struct {
	bodyWriter io.WriteCloser
	done       chan struct{}
}

// Forwarder proxies request frames to the local application and streams
// responses back as binary frames. mu guards active; sendFrame is set once by the client.
type Forwarder struct {
	localURL  string
	subdomain string
	client    *http.Client
	log       *log.Logger
	mu        sync.Mutex
	active    map[uint64]*activeReq
	sendFrame sendFrameFunc
}

// NewForwarder returns a Forwarder for the given local base URL and subdomain.
// The client must call SetSendFrame before any requests are dispatched.
func NewForwarder(localURL, subdomain string, l *log.Logger) *Forwarder {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:  false,
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &Forwarder{
		localURL:  strings.TrimSuffix(localURL, "/"),
		subdomain: subdomain,
		client:    &http.Client{Transport: transport, Timeout: config.ProxyRequestTimeout},
		log:       l,
		active:    make(map[uint64]*activeReq),
	}
}

// SetSendFrame sets the function used to send frames to the server. Must be
// called by the client before dispatching any requests.
func (c *Forwarder) SetSendFrame(fn sendFrameFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sendFrame = fn
}

func (c *Forwarder) getSend() sendFrameFunc {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sendFrame
}

// HandleRequest is called in a new goroutine from the client demux when
// FrameHTTPReqStart is received. It unmarshals the payload, builds a local
// http.Request, and streams the response back as frames.
func (c *Forwarder) HandleRequest(ctx context.Context, id uint64, payload []byte) {
	send := c.getSend()
	if send == nil {
		return
	}
	sp, err := frame.UnmarshalStart(payload)
	if err != nil {
		c.sendError(id, send, 502, "invalid start payload")
		return
	}
	path := sp.Path
	if c.subdomain != "" && len(path) > len(c.subdomain)+2 {
		prefix := "/" + c.subdomain + "/"
		if strings.HasPrefix(path, prefix) {
			path = path[len(prefix):]
		}
	}
	path = strings.TrimPrefix(path, "/")
	fullURL := c.localURL + "/" + path

	pr, pw := io.Pipe()
	c.mu.Lock()
	c.active[id] = &activeReq{bodyWriter: pw, done: make(chan struct{})}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, sp.Method, fullURL, pr)
	if err != nil {
		c.removeActive(id)
		pw.Close()
		c.sendError(id, send, 502, err.Error())
		return
	}
	for _, hf := range sp.Headers {
		req.Header.Set(hf.Key, hf.Value)
	}

	go c.doAndStream(ctx, id, req, send, pw)
}

// doAndStream runs http.Client.Do and streams the response as frames.
func (c *Forwarder) doAndStream(ctx context.Context, id uint64, req *http.Request, send sendFrameFunc, pw *io.PipeWriter) {
	defer c.removeActive(id)
	defer pw.Close()
	resp, err := c.client.Do(req)
	if err != nil {
		c.sendError(id, send, 502, err.Error())
		return
	}
	defer resp.Body.Close()

	rsp := frame.StartPayload{Status: resp.StatusCode}
	for k, vs := range resp.Header {
		for _, v := range vs {
			rsp.Headers = append(rsp.Headers, frame.HeaderField{Key: k, Value: v})
		}
	}
	payload, err := frame.MarshalStart(rsp)
	if err != nil {
		c.sendError(id, send, 502, "marshal response")
		return
	}
	if err := send(frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(payload))}, payload); err != nil {
		return
	}

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if sendErr := send(frame.Header{Type: frame.FrameHTTPRespBody, RequestID: id, PayloadLen: uint32(n)}, buf[:n]); sendErr != nil {
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
	_ = send(frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)

	c.mu.Lock()
	if ar, ok := c.active[id]; ok {
		close(ar.done)
	}
	c.mu.Unlock()
}

func (c *Forwarder) sendError(id uint64, send sendFrameFunc, status int, msg string) {
	sp := frame.StartPayload{Status: status, Headers: []frame.HeaderField{{Key: "X-Error", Value: msg}}}
	payload, _ := frame.MarshalStart(sp)
	_ = send(frame.Header{Type: frame.FrameHTTPRespStart, RequestID: id, PayloadLen: uint32(len(payload))}, payload)
	_ = send(frame.Header{Type: frame.FrameHTTPRespEnd, RequestID: id, PayloadLen: 0}, nil)
}

func (c *Forwarder) removeActive(id uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.active, id)
}

// DeliverBody writes the payload to the active request's body pipe. Called
// from the client demux when FrameHTTPReqBody is received.
func (c *Forwarder) DeliverBody(id uint64, payload []byte) {
	c.mu.Lock()
	ar := c.active[id]
	c.mu.Unlock()
	if ar != nil && ar.bodyWriter != nil {
		_, _ = ar.bodyWriter.Write(payload)
	}
}

// DeliverBodyEnd closes the active request's body writer (signals EOF to
// http.Client). Called when FrameHTTPReqEnd is received.
func (c *Forwarder) DeliverBodyEnd(id uint64) {
	c.mu.Lock()
	ar := c.active[id]
	c.mu.Unlock()
	if ar != nil && ar.bodyWriter != nil {
		_ = ar.bodyWriter.Close()
	}
}

var bufPool = sync.Pool{
	New: func() any { return make([]byte, config.CopyBufferSize) },
}
