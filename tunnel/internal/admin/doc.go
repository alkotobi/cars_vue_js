// Package admin provides an HTTP server on a separate admin port exposing
// health, stats, and pprof endpoints. It must NEVER be on the same port as
// the proxy server to avoid routing collisions with tunnel IDs. Binds to
// 127.0.0.1 by default; do not expose the admin port to the internet.
//
// Concurrency model: one goroutine runs http.Server.Serve; handlers are
// invoked per request by the server and read from the registry (thread-safe).
package admin

