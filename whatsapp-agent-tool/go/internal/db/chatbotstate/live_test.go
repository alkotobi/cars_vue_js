package chatbotstate

import (
	"os"
	"testing"
	"time"

	"whatsapp-agent-tool/internal/db"
)

// TestLeaseAndCursorsLive exercises the real shared state (tables
// chatbot_leases / chatbot_group_cursors, migration 009) against the `ai` DB.
// Opt-in: set CHATBOT_DB_LIVE=1 (it hits the live database and needs the
// migration applied). It uses a dedicated tenant and cleans up after itself.
func TestLeaseAndCursorsLive(t *testing.T) {
	if !liveEnabled() {
		t.Skip("set CHATBOT_DB_LIVE=1 to run the live chatbot-state smoke")
	}
	if _, err := db.Open(); err != nil {
		t.Fatal(err)
	}
	defer db.ClosePool()

	const tenant = "__chatbotstate_live_test__"
	clean := func() {
		_, _ = db.Execute("DELETE FROM chatbot_leases WHERE tenant = ?", tenant)
		_, _ = db.Execute("DELETE FROM chatbot_group_cursors WHERE tenant = ?", tenant)
	}
	clean()
	defer clean()

	// A short TTL so the failover can be observed without a long sleep.
	const ttl = 3 * time.Second

	// First claim wins; a second owner is refused while the lease is live.
	got, err := Acquire(tenant, "probeA", ttl)
	if err != nil || !got {
		t.Fatalf("Acquire A = %v, %v; want true", got, err)
	}
	got, err = Acquire(tenant, "probeB", ttl)
	if err != nil || got {
		t.Fatalf("Acquire B while A holds = %v, %v; want false", got, err)
	}

	// Renew keeps A holding; B cannot renew what it does not hold.
	if ok, err := Renew(tenant, "probeA", ttl); err != nil || !ok {
		t.Fatalf("Renew A = %v, %v; want true", ok, err)
	}
	if ok, err := Renew(tenant, "probeB", ttl); err != nil || ok {
		t.Fatalf("Renew B = %v, %v; want false", ok, err)
	}

	// Cursors are monotonic.
	if err := SaveCursor(tenant, 7, 500); err != nil {
		t.Fatal(err)
	}
	if err := SaveCursor(tenant, 7, 400); err != nil {
		t.Fatal(err)
	}
	cur, err := LoadCursors(tenant)
	if err != nil {
		t.Fatal(err)
	}
	if cur[7] != 500 {
		t.Fatalf("cursor = %d, want 500 (monotonic)", cur[7])
	}

	// Another owner is still refused while A's lease is live.
	if ok, err := Acquire(tenant, "probeC", ttl); err != nil || ok {
		t.Fatalf("Acquire C while A holds = %v, %v; want false", ok, err)
	}

	// A stops renewing: after the TTL its lease expires, C takes over, and the
	// stale owner A can no longer renew.
	time.Sleep(ttl + 500*time.Millisecond)
	if ok, err := Acquire(tenant, "probeC", 30*time.Second); err != nil || !ok {
		t.Fatalf("Acquire C after A expired = %v, %v; want true", ok, err)
	}
	if ok, err := Renew(tenant, "probeA", 30*time.Second); err != nil || ok {
		t.Fatalf("Renew stale A = %v, %v; want false", ok, err)
	}
	if err := Release(tenant, "probeC"); err != nil {
		t.Fatal(err)
	}
	if held, err := Holds(tenant, "probeC"); err != nil || held {
		t.Fatalf("Holds C after Release = %v, %v; want false", held, err)
	}
}

func liveEnabled() bool {
	return os.Getenv("CHATBOT_DB_LIVE") == "1"
}
