// The single-writer lease keeper for the chat AI. While a poll cycle runs it
// holds the tenant's lease open: a background goroutine renews expires_at on a
// heartbeat (TTL/3) and flips Lost() the moment the lease is gone, so the cycle
// can stop before posting a reply another machine is about to post too. stop()
// releases the lease so a standby takes over without waiting out the TTL.
package chatbot

import (
	"log"
	"sync/atomic"
	"time"

	"whatsapp-agent-tool/internal/db/chatbotstate"
)

// leaseKeeper renews a held lease for the duration of one poll cycle. It is safe
// for a single goroutine to observe Lost() while the heartbeat runs.
type leaseKeeper struct {
	tenant string
	owner  string
	ttl    time.Duration
	log    *log.Logger

	stopCh chan struct{}
	done   chan struct{}
	lost   atomic.Bool
}

// startLeaseKeeper begins renewing the (already acquired) lease and returns a
// keeper the caller must stop() when the cycle ends.
func startLeaseKeeper(tenant, owner string, ttl time.Duration, logger *log.Logger) *leaseKeeper {
	k := &leaseKeeper{
		tenant: tenant,
		owner:  owner,
		ttl:    ttl,
		log:    logger,
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
	go k.loop()
	return k
}

func (k *leaseKeeper) loop() {
	defer close(k.done)
	interval := k.ttl / 3
	if interval <= 0 {
		interval = time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-k.stopCh:
			return
		case <-t.C:
			ok, err := chatbotstate.Renew(k.tenant, k.owner, k.ttl)
			if err != nil {
				k.log.Printf("lease: renew failed: %v", err)
				k.lost.Store(true)
				return
			}
			if !ok {
				k.log.Printf("lease: lost — taken over by another daemon or expired")
				k.lost.Store(true)
				return
			}
		}
	}
}

// Lost reports whether the lease was lost mid-cycle. Once true it stays true.
func (k *leaseKeeper) Lost() bool { return k.lost.Load() }

// stop ends the heartbeat and releases the lease so a standby can take over.
func (k *leaseKeeper) stop() {
	close(k.stopCh)
	<-k.done
	if err := chatbotstate.Release(k.tenant, k.owner); err != nil {
		k.log.Printf("lease: release failed: %v", err)
	}
}
