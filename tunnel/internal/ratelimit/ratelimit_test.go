package ratelimit_test

import (
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/ratelimit"
)

func TestSingleIPUnderLimit(t *testing.T) {
	l := ratelimit.NewLimiter(10, 10)
	addr := netip.MustParseAddr("192.168.1.1")
	for i := 0; i < 10; i++ {
		if !l.Allow(addr) {
			t.Errorf("request %d: expected allowed", i+1)
		}
	}
}

func TestSingleIPBurst(t *testing.T) {
	l := ratelimit.NewLimiter(200, 200)
	addr := netip.MustParseAddr("10.0.0.1")
	for i := 0; i < 200; i++ {
		if !l.Allow(addr) {
			t.Errorf("request %d: expected allowed", i+1)
		}
	}
	// 201st and beyond denied
	for i := 0; i < 10; i++ {
		if l.Allow(addr) {
			t.Errorf("request 201+ %d: expected denied", i+1)
		}
	}
}

func TestTokenRefill(t *testing.T) {
	l := ratelimit.NewLimiter(10, 1) // 10/s, capacity 1
	addr := netip.MustParseAddr("172.16.0.1")
	if !l.Allow(addr) {
		t.Fatal("first request should be allowed")
	}
	if l.Allow(addr) {
		t.Fatal("second request immediate should be denied")
	}
	time.Sleep(150 * time.Millisecond) // refill ~1.5 tokens (cap 1 → 1 token)
	if !l.Allow(addr) {
		t.Error("after 100ms refill, request should be allowed")
	}
}

func TestConcurrent(t *testing.T) {
	l := ratelimit.NewLimiter(1000, 1000)
	addr := netip.MustParseAddr("192.168.0.42")
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = l.Allow(addr)
			}
		}()
	}
	wg.Wait()
	// No race (run with -race); total 64*20 = 1280, capacity 1000 so at least 280 denied
}

func TestDifferentIPsIndependent(t *testing.T) {
	l := ratelimit.NewLimiter(10, 2)
	a1 := netip.MustParseAddr("192.168.1.1")
	a2 := netip.MustParseAddr("192.168.2.1") // different /24
	if !l.Allow(a1) || !l.Allow(a1) {
		t.Fatal("a1 first two should be allowed")
	}
	if l.Allow(a1) {
		t.Fatal("a1 third should be denied")
	}
	// a2 has its own bucket (different /24)
	if !l.Allow(a2) || !l.Allow(a2) {
		t.Fatal("a2 first two should be allowed")
	}
	if l.Allow(a2) {
		t.Fatal("a2 third should be denied")
	}
}

func TestIPv6Slash48Grouping(t *testing.T) {
	l := ratelimit.NewLimiter(10, 2)
	// Same /48, different host bits
	a1 := netip.MustParseAddr("2001:db8::1")
	a2 := netip.MustParseAddr("2001:db8::2")
	if !l.Allow(a1) || !l.Allow(a1) {
		t.Fatal("a1 first two should be allowed")
	}
	if l.Allow(a1) {
		t.Fatal("a1 third should be denied")
	}
	// a2 shares bucket with a1 (same /48)
	if l.Allow(a2) {
		t.Fatal("a2 should be denied (bucket shared with a1, exhausted)")
	}
}

func TestCleanupDoesNotRemoveRecent(t *testing.T) {
	l := ratelimit.NewLimiter(10, 10)
	addr := netip.MustParseAddr("192.168.99.1")
	_ = l.Allow(addr)
	l.Cleanup()
	if !l.Allow(addr) {
		t.Error("Allow after Cleanup (recent bucket) should succeed")
	}
}

func TestCleanupRemovesStale(t *testing.T) {
	l := ratelimit.NewLimiter(10, 1)
	addr := netip.MustParseAddr("192.168.99.2")
	_ = l.Allow(addr) // bucket now has 0 tokens
	l.SetCleanupStaleForTest(1 * time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	l.Cleanup() // bucket not seen for 2ms > 1ms → removed
	l.SetCleanupStaleForTest(0)
	// Next Allow gets a fresh bucket (full tokens) → allowed
	if !l.Allow(addr) {
		t.Error("Allow after Cleanup of stale bucket should get fresh bucket and succeed")
	}
}

func BenchmarkAllow(b *testing.B) {
	l := ratelimit.NewLimiter(200, 200)
	addr := netip.MustParseAddr("192.168.1.1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Allow(addr)
	}
}
