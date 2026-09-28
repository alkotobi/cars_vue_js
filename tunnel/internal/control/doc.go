// Package control handles the tunnel server side of each client's control
// connection. It speaks a self-framing protocol: JSON control messages
// (newline-delimited) for registration and heartbeat, and binary frames for
// all proxied HTTP and WebSocket data.
//
// Concurrency model: one goroutine per client connection (started by the
// server's Accept loop). That goroutine owns a bufio.Reader and runs the
// demux loop; proxy request handling is spawned separately via frameDispatch.
package control
