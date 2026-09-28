// Package proxy contains the HTTP proxy implementation on the server side,
// including backpressure via semaphores and buffered io.Copy calls.
//
// Concurrency model: one goroutine per active proxy request, with shared
// limits enforced by per-client and global semaphores.
package proxy

