package registry

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yourname/tunnel/internal/config"
	"github.com/yourname/tunnel/internal/frame"
)

// Sentinel errors for the registry package.
var (
	ErrDomainTaken   = errors.New("registry: domain already taken")
	ErrAtCapacity    = errors.New("registry: at maximum client capacity")
	ErrClientGone    = errors.New("registry: client disconnected")
	ErrInvalidDomain = errors.New("registry: invalid domain characters")
)

// Client represents a connected tunnel client (one per control connection).
// Client fields updated by proxy goroutines (Requests, BytesIn, BytesOut) use
// atomics; LastHeartbeat and Connected are written by the control/heartbeat
// goroutine and read by the registry.
type Client struct {
	ID            int64
	Domain        string
	WorkerID      int
	ConnectedAt   time.Time
	LastHeartbeat atomic.Int64 // Unix nano; written by heartbeat goroutine
	Connected     atomic.Bool  // written on disconnect

	// Send queues a frame to be written to the client's control connection.
	// Returns ErrClientGone if the client has disconnected.
	Send func(h frame.Header, payload []byte) error

	// Stats — updated with atomic adds from proxy goroutines
	Requests atomic.Int64
	BytesIn  atomic.Int64
	BytesOut atomic.Int64
}

// Registrar is the interface for client storage.
type Registrar interface {
	Register(c *Client) error
	Unregister(id int64)
	Find(domain string) (*Client, bool)
	List() []*Client
	Count() int
	CollectStale(timeout time.Duration) []*Client
}

// registry is the concrete implementation. mu guards byID and byDomain;
// nextID is used only under mu when assigning new IDs.
type registry struct {
	mu       sync.RWMutex
	byID     map[int64]*Client
	byDomain map[string]*Client
	nextID   atomic.Int64
	maxClients int
}

// New returns a Registrar with capacity config.MaxRegisteredClients.
func New() Registrar {
	return newRegistry(config.MaxRegisteredClients)
}

// NewWithMax returns a Registrar with the given maximum client count.
// It is intended for tests (e.g. AtCapacity); production code uses New().
func NewWithMax(max int) Registrar {
	return newRegistry(max)
}

func newRegistry(max int) *registry {
	return &registry{
		byID:       make(map[int64]*Client),
		byDomain:   make(map[string]*Client),
		maxClients: max,
	}
}

// validateDomain returns nil only if domain is 1–MaxDomainLen chars and
// contains only [a-z0-9-].
func validateDomain(domain string) error {
	if domain == "" || len(domain) > config.MaxDomainLen {
		return ErrInvalidDomain
	}
	for _, r := range domain {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return ErrInvalidDomain
		}
	}
	return nil
}

// Register adds the client to the registry. It assigns c.ID if zero, validates
// c.Domain, and returns ErrDomainTaken or ErrAtCapacity when applicable.
func (r *registry) Register(c *Client) error {
	if err := validateDomain(c.Domain); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.byDomain) >= r.maxClients {
		return ErrAtCapacity
	}
	if _, exists := r.byDomain[c.Domain]; exists {
		return ErrDomainTaken
	}
	if c.ID == 0 {
		c.ID = r.nextID.Add(1)
	}
	c.ConnectedAt = time.Now()
	c.LastHeartbeat.Store(time.Now().UnixNano())
	c.Connected.Store(true)
	r.byID[c.ID] = c
	r.byDomain[c.Domain] = c
	return nil
}

// Unregister removes the client by id and marks it disconnected. No-op if not found.
func (r *registry) Unregister(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return
	}
	delete(r.byID, id)
	delete(r.byDomain, c.Domain)
	c.Connected.Store(false)
}

// Find returns the client for the given domain, or (nil, false) if not found.
// Callers must check client.Connected.Load() to distinguish online vs offline (e.g. proxy returns 502 when offline).
func (r *registry) Find(domain string) (*Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byDomain[domain]
	if !ok {
		return nil, false
	}
	return c, true
}

// List returns a snapshot of all registered clients (copy of values).
func (r *registry) List() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Client, 0, len(r.byID))
	for _, c := range r.byID {
		out = append(out, c)
	}
	return out
}

// Count returns the number of registered clients.
func (r *registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byID)
}

// CollectStale returns clients whose LastHeartbeat is older than timeout.
// It does not remove them; the caller should Unregister after cleanup.
func (r *registry) CollectStale(timeout time.Duration) []*Client {
	cutoff := time.Now().Add(-timeout).UnixNano()
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Client
	for _, c := range r.byID {
		if c.LastHeartbeat.Load() < cutoff {
			out = append(out, c)
		}
	}
	return out
}

// GenerateDomain returns a unique-enough subdomain for dev use: "tunnel-<timestamp>-<rand>".
func GenerateDomain() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return fmt.Sprintf("tunnel-%d-%x", time.Now().UnixNano(), buf)
}
