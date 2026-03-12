// Package ratelimit implements per-IP token-bucket rate limiting for tunnel
// control connections and HTTP proxy requests. Each bucket is independently
// tracked with no cross-goroutine locking on the hot path (65536 shards by
// FNV-1a hash of normalised IP). IPv6 uses /48 prefix, IPv4 uses /24.
//
// Concurrency model: Limiter is safe for concurrent use; each shard has its
// own mutex so different IPs rarely contend.
package ratelimit

