package registry_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourname/tunnel/internal/frame"
	"github.com/yourname/tunnel/internal/registry"
)

func clientWithDomain(domain string) *registry.Client {
	return &registry.Client{
		Domain: domain,
		Send:   func(frame.Header, []byte) error { return nil },
	}
}

func TestRegisterFindRoundTrip(t *testing.T) {
	reg := registry.New()
	c := clientWithDomain("cars")
	if err := reg.Register(c); err != nil {
		t.Fatalf("Register: %v", err)
	}
	got, ok := reg.Find("cars")
	if !ok || got != c {
		t.Fatalf("Find: ok=%v, got=%v, want c", ok, got)
	}
}

func TestRegisterDuplicateDomainReturnsErrDomainTaken(t *testing.T) {
	reg := registry.New()
	c1 := clientWithDomain("cars")
	c2 := clientWithDomain("cars")
	if err := reg.Register(c1); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := reg.Register(c2)
	if err != registry.ErrDomainTaken {
		t.Fatalf("second Register: got %v, want ErrDomainTaken", err)
	}
}

func TestUnregisterThenFindReturnsFalse(t *testing.T) {
	reg := registry.New()
	c := clientWithDomain("cars")
	if err := reg.Register(c); err != nil {
		t.Fatalf("Register: %v", err)
	}
	reg.Unregister(c.ID)
	_, ok := reg.Find("cars")
	if ok {
		t.Fatal("Find after Unregister: want false")
	}
}

func TestConcurrentStress(t *testing.T) {
	reg := registry.New()
	var idGen atomic.Int64
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				id := idGen.Add(1)
				domain := fmt.Sprintf("stress-%d", id)
				c := clientWithDomain(domain)
				if err := reg.Register(c); err != nil {
					// domain taken or capacity is fine
					continue
				}
				_, _ = reg.Find(c.Domain)
				reg.Unregister(c.ID)
			}
		}()
	}
	time.Sleep(1 * time.Second)
	close(done)
	wg.Wait()
}

func TestCollectStale(t *testing.T) {
	reg := registry.New()
	cutoff := time.Now().Add(-200 * time.Second).UnixNano()
	for i := 0; i < 100; i++ {
		c := clientWithDomain(fmt.Sprintf("stale-%d", i))
		if err := reg.Register(c); err != nil {
			t.Fatalf("Register %d: %v", i, err)
		}
		c.LastHeartbeat.Store(cutoff - int64(i))
	}
	stale := reg.CollectStale(199 * time.Second)
	if len(stale) != 100 {
		t.Fatalf("CollectStale(199s): got %d, want 100", len(stale))
	}
}

func TestCount(t *testing.T) {
	reg := registry.New()
	if n := reg.Count(); n != 0 {
		t.Fatalf("Count empty: got %d", n)
	}
	c1 := clientWithDomain("a")
	c2 := clientWithDomain("b")
	_ = reg.Register(c1)
	_ = reg.Register(c2)
	if n := reg.Count(); n != 2 {
		t.Fatalf("Count after 2 register: got %d", n)
	}
	reg.Unregister(c1.ID)
	if n := reg.Count(); n != 1 {
		t.Fatalf("Count after one unregister: got %d", n)
	}
	reg.Unregister(c2.ID)
	if n := reg.Count(); n != 0 {
		t.Fatalf("Count after both unregister: got %d", n)
	}
}

func TestListSnapshotExcludesUnregistered(t *testing.T) {
	reg := registry.New()
	c := clientWithDomain("list-one")
	_ = reg.Register(c)
	list1 := reg.List()
	if len(list1) != 1 {
		t.Fatalf("List after one register: got %d", len(list1))
	}
	reg.Unregister(c.ID)
	list2 := reg.List()
	if len(list2) != 0 {
		t.Fatalf("List after unregister: got %d", len(list2))
	}
}

func TestRegisterAtCapacityReturnsErrAtCapacity(t *testing.T) {
	reg := registry.NewWithMax(5)
	for i := 0; i < 5; i++ {
		c := clientWithDomain("cap-" + string(rune('a'+i)))
		if err := reg.Register(c); err != nil {
			t.Fatalf("Register %d: %v", i, err)
		}
	}
	c6 := clientWithDomain("cap-sixth")
	err := reg.Register(c6)
	if err != registry.ErrAtCapacity {
		t.Fatalf("6th Register: got %v, want ErrAtCapacity", err)
	}
}

func TestRegisterInvalidDomain(t *testing.T) {
	reg := registry.New()
	for _, domain := range []string{"Uppercase", "has space", "has/slash", ""} {
		c := clientWithDomain(domain)
		err := reg.Register(c)
		if err != registry.ErrInvalidDomain {
			t.Fatalf("Register %q: got %v, want ErrInvalidDomain", domain, err)
		}
	}
}

func TestGenerateDomainUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 10_000; i++ {
		d := registry.GenerateDomain()
		if _, ok := seen[d]; ok {
			t.Fatalf("duplicate domain at %d: %s", i, d)
		}
		seen[d] = struct{}{}
	}
}
