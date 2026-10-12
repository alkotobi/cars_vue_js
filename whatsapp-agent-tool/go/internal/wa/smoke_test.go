package wa

import (
	"fmt"
	"os"
	"testing"
	"time"

	"whatsapp-agent-tool/internal/chrome"
)

// TestSmokeLive attaches to the persistent Chrome and runs the read-only happy
// path (default profile only). It never sends, never marks read, never writes a
// chat. Run explicitly: WA_SMOKE=1 go test -run TestSmokeLive ./internal/wa -v
func TestSmokeLive(t *testing.T) {
	if os.Getenv("WA_SMOKE") != "1" {
		t.Skip("set WA_SMOKE=1 to run the read-only live smoke (attaches to the persistent Chrome)")
	}
	if chrome.Profile() != "whatsapp-profile" {
		t.Fatalf("smoke must run on the default profile only; got %q", chrome.Profile())
	}

	b := NewBrowser()
	if !b.attachToRunningChrome() {
		t.Skip("no CDP-debuggable Chrome on our port — nothing to attach to")
	}
	t.Cleanup(func() { _ = b.Close() })

	if err := b.useWhatsAppPage(); err != nil {
		t.Fatalf("useWhatsAppPage: %v", err)
	}

	state, _ := b.waitForAppReady(60 * time.Second)
	t.Logf("loginState=%s url=%s", state, b.pageURL())
	if state != "logged_in" {
		t.Fatalf("expected logged_in, got %s (do not proceed unless the default profile is linked)", state)
	}

	loggedIn, err := b.IsLoggedIn()
	if err != nil || !loggedIn {
		t.Fatalf("IsLoggedIn=%v err=%v", loggedIn, err)
	}

	info, err := b.SessionInfo()
	if err != nil {
		t.Fatalf("SessionInfo: %v", err)
	}
	t.Logf("session: profile=%s transport=%s pid=%v port=%d running=%v page=%v",
		info.Profile, info.Transport, info.ChromePid, info.CdpPort, info.ChromeRunning, info.PageUrl)
	if info.Transport != "cdp" {
		t.Fatalf("expected cdp transport, got %s", info.Transport)
	}

	screen, err := b.GetChats(15)
	if err != nil {
		t.Fatalf("GetChats: %v", err)
	}
	t.Logf("screen chats: %d", len(screen))
	for i, c := range screen {
		t.Logf("  [%d] %q preview=%q time=%q unread=%v (%s)", i, c.Name, c.Preview, c.Time, c.Unread, c.UnreadCount)
	}
	if len(screen) == 0 {
		t.Fatal("screen returned zero rendered rows")
	}
	fmt.Printf("SMOKE OK screen=%d\n", len(screen))
}
