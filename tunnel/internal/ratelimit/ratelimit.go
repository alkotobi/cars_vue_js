// Package ratelimit implements per-IP token-bucket rate limiting for tunnel
// control connections and HTTP proxy requests. Each bucket is independently
// tracked with no cross-goroutine locking on the hot path (sharded by FNV-1a
// hash of the normalised IP).
//
// Recommended limits (document in call sites):
//   - Control connections: rate=10/s, capacity=10 (10 new tunnels/s per /24 or /48)
//   - HTTP requests: rate=200/s, capacity=200 (200 req/s per /24 or /48)
package ratelimit

import (
	"hash/fnv"
	"net/netip"
	"sync"
	"time"
)

const numShards = 65536

// cleanupStaleAfter is how long a bucket can be idle before Cleanup removes it.
const cleanupStaleAfter = 5 * time.Minute

// bucket holds token count and last refill time for one normalised address.
// capacity and rate are on the Limiter; only per-bucket state lives here.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// shard protects a map of buckets; mu guards buckets.
type shard struct {
	mu      sync.Mutex
	buckets map[netip.Addr]*bucket
}

// Limiter is a sharded token-bucket rate limiter keyed by normalised IP.
// capacity is the max tokens per bucket; rate is tokens added per second.
// staleAfterOverride is for tests only (0 = use cleanupStaleAfter).
type Limiter struct {
	shards             [numShards]shard
	capacity           float64
	rate               float64
	staleAfterOverride time.Duration
}

// NewLimiter creates a Limiter allowing rate tokens per second with burst capacity.
func NewLimiter(rate, capacity float64) *Limiter {
	return &Limiter{capacity: capacity, rate: rate}
}

// shardIndex returns the shard index for the normalised address using FNV-1a.
func (l *Limiter) shardIndex(addr netip.Addr) uint64 {
	h := fnv.New64a()
	h.Write(addr.AsSlice())
	return h.Sum64() % numShards
}

func (l *Limiter) shard(addr netip.Addr) *shard {
	return &l.shards[l.shardIndex(addr)]
}

// normaliseAddr returns a canonical key for the address: IPv6 uses /48 prefix
// (zero last 80 bits), IPv4 uses /24 prefix (zero last 8 bits).
func normaliseAddr(addr netip.Addr) netip.Addr {
	if addr.Is6() {
		p, err := addr.Prefix(48)
		if err != nil {
			return addr
		}
		return p.Addr()
	}
	p, err := addr.Prefix(24)
	if err != nil {
		return addr
	}
	return p.Addr()
}

// Allow returns true if the request from addr should be allowed (one token
// consumed). Thread-safe; O(1) amortised. Refills tokens based on elapsed
// time since last refill, then allows if tokens >= 1.
func (l *Limiter) Allow(addr netip.Addr) bool {
	key := normaliseAddr(addr)
	s := l.shard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.buckets == nil {
		s.buckets = make(map[netip.Addr]*bucket)
	}
	b, ok := s.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, lastSeen: time.Now()}
		s.buckets[key] = b
	}
	now := time.Now()
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastSeen = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// SetCleanupStaleForTest sets the stale duration used by Cleanup (for tests only).
// Call with 0 to reset to the default 5 minutes.
func (l *Limiter) SetCleanupStaleForTest(d time.Duration) {
	l.staleAfterOverride = d
}

// Cleanup removes buckets not seen for longer than cleanupStaleAfter. Call
// periodically (e.g. from a timer) to avoid unbounded map growth.
func (l *Limiter) Cleanup() {
	after := cleanupStaleAfter
	if l.staleAfterOverride != 0 {
		after = l.staleAfterOverride
	}
	cutoff := time.Now().Add(-after)
	for i := range l.shards {
		s := &l.shards[i]
		s.mu.Lock()
		for addr, b := range s.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(s.buckets, addr)
			}
		}
		s.mu.Unlock()
	}
}
