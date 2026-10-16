package chatbot

import (
	"os"
	"testing"
)

// TestGenerateManual exercises the real opencode CLI + local model end to end:
// workdir provisioning (config copy + model override) and one reply. Skip by
// default — run it on a machine with opencode + ollama up:
//
//	CHATBOT_LIVE=1 go test ./internal/chatbot/ -run TestGenerateManual -v
func TestGenerateManual(t *testing.T) {
	if os.Getenv("CHATBOT_LIVE") == "" {
		t.Skip("set CHATBOT_LIVE=1 to run the live opencode generation smoke")
	}
	reply, _, err := Generate("Reply with exactly the word: bonjour", Perms{})
	if err != nil {
		t.Fatal(err)
	}
	if len(reply) < 3 {
		t.Fatalf("unexpectedly short reply: %q", reply)
	}
	t.Logf("reply: %s", reply)
}
