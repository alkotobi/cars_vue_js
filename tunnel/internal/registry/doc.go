// Package registry maintains the set of connected tunnel clients and provides
// O(1) lookup by subdomain. It is safe for concurrent use by multiple goroutines.
//
// Concurrency model: registry uses a single sync.RWMutex for byID and byDomain
// maps; read operations use RLock, write operations (Register, Unregister) use
// Lock. Client fields use sync/atomic for LastHeartbeat, Connected, and stats
// so proxy goroutines can update them without holding the registry mutex.
package registry
