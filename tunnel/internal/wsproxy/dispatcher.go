package wsproxy

import (
	"context"

	"github.com/yourname/tunnel/internal/frame"
)

// Dispatcher delivers WebSocket frames to the appropriate pending session,
// looked up by frame.RequestID.
type Dispatcher interface {
	// DeliverFrame delivers a frame to the session identified by the header.
	DeliverFrame(ctx context.Context, h frame.Header, payload []byte) error
}
