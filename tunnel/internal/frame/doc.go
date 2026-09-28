// Package frame implements the binary framing protocol for the tunnel data
// channel. All HTTP request/response bodies and WebSocket data travel as
// binary frames over a persistent connection between tunnel server and client.
//
// Binary frames are used instead of JSON+base64 because text encodings inflate
// payloads by roughly 33% and typically require full buffering before decode,
// which breaks true streaming for large files, video, and server-sent events.
// With binary frames, raw bytes move in bounded chunks and the receiver can
// process each chunk as it arrives without waiting for the full body.
//
// Concurrency model: one reader goroutine per connection, with higher-level
// packages responsible for dispatching decoded frames onto per-stream
// goroutines.
package frame

