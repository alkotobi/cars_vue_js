// Package wsproxy tunnels WebSocket connections through the binary frame protocol.
// A WebSocket upgrade from the browser is forwarded as FrameHTTPReqStart; the 101
// response comes as FrameHTTPRespStart; thereafter frames use FrameWSData/FrameWSClose.
//
// Concurrency model: one goroutine per active WebSocket session (browserToTunnel,
// tunnelToBrowser, idleWatcher), with clear shutdown via context cancellation.
// Session map is sharded by request ID to reduce lock contention.
package wsproxy

