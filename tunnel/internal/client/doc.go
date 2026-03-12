// Package client implements the tunnel client. It maintains a persistent
// control connection to the tunnel server, receives proxy request frames,
// forwards them to the local application using Go's net/http client, and
// streams responses back as binary frames. No libcurl — pure Go.
//
// Concurrency model: one control goroutine per connection (demux loop), one
// goroutine per active forwarded request (Forwarder.HandleRequest), plus
// heartbeat and backoff. Pipe-based body streaming connects frame delivery
// to http.Client. http.Transport is reused for keep-alive. Exponential
// backoff (1s, 2s, 4s, … cap at ReconnectMax) on connect failure.
package client

