package proxy

import (
	"context"

	"github.com/yourname/tunnel/internal/frame"
)

// Dispatcher delivers response frames (HTTP or WebSocket) to the appropriate
// pending request or session, looked up by frame.RequestID.
type Dispatcher interface {
	// DeliverFrame delivers a frame to the request/session identified by the header.
	DeliverFrame(ctx context.Context, h frame.Header, payload []byte) error
}
