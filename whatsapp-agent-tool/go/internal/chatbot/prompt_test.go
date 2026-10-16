package chatbot

import (
	"strings"
	"testing"

	"whatsapp-agent-tool/internal/carshttp"
)

func TestBuildPrompt(t *testing.T) {
	history := []carshttp.GroupMessage{
		{ID: 1, Time: "2026-10-13 09:00:00", SenderName: "yacine", SenderRole: "Admin", Message: "bonjour"},
	}
	incoming := []carshttp.GroupMessage{
		{ID: 2, Time: "2026-10-13 09:01:00", SenderName: "Ahmed", SenderRole: "Client", Message: "wesh كيفاش"},
	}
	p := BuildPrompt("Selection: LIVAN", "ai", Perms{}, history, incoming)

	for _, want := range []string{
		`"ai"`,
		`"Selection: LIVAN"`,
		"<yacine> Admin", // history line
		"<Ahmed> Client", // incoming line
		"Algerian Derja/Arabic",
		"Arabic numerals",
		"whatsapp-tool freight", // shipping questions answered without a car
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q\n%s", want, p)
		}
	}
	if strings.Contains(p, "هذا الرد مولد") {
		t.Error("prompt must not contain the AI disclosure line (it is appended later by us)")
	}
}

func TestFinalizeReply(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		max      int
		want     string
		wantFail bool
	}{
		{name: "plain", raw: "  Salam khouya 🙂  ", max: 400, want: "Salam khouya 🙂\n" + AiDisclosureLine},
		{name: "fenced", raw: "```\nVoilà le prix\n```", max: 400, want: "Voilà le prix\n" + AiDisclosureLine},
		{name: "blank", raw: "   \n```\n```\n  ", max: 400, wantFail: true},
		{name: "truncated", raw: strings.Repeat("a", 50) + " " + strings.Repeat("b", 50), max: 30, want: strings.Repeat("a", 30) + "\n" + AiDisclosureLine},
		{name: "inline backticks", raw: "`ok` bon", max: 400, want: "ok bon\n" + AiDisclosureLine},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FinalizeReply(tc.raw, tc.max)
			if tc.wantFail {
				if got != "" {
					t.Fatalf("want empty reply, got %q", got)
				}
				return
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if !strings.HasSuffix(got, "\n"+AiDisclosureLine) {
				t.Fatalf("reply must end with the AI disclosure line: %q", got)
			}
		})
	}
}

func TestFinalizeReplyCapsRunesNotBytes(t *testing.T) {
	raw := strings.Repeat("مرحبا ", 50)
	got := FinalizeReply(raw, 30)
	if strings.Contains(got, "\ufffd") {
		t.Fatalf("reply contains a broken rune: %q", got)
	}
}
