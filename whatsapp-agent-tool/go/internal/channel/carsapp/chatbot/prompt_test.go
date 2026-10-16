package chatbot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"whatsapp-agent-tool/internal/config"
)

func TestBuildPrompt(t *testing.T) {
	history := []Message{
		{ID: 1, Time: "2026-10-13 09:00:00", SenderName: "yacine", SenderRole: "Admin", Text: "bonjour"},
	}
	incoming := []Message{
		{ID: 2, Time: "2026-10-13 09:01:00", SenderName: "Ahmed", SenderRole: "Client", Text: "wesh كيفاش"},
	}
	p := BuildPrompt("Selection: LIVAN", "ai", Perms{}, history, incoming)

	for _, want := range []string{
		`"ai"`,
		`"Selection: LIVAN"`,
		"<yacine> Admin", // history line
		"<Ahmed> Client", // incoming line
		"exact language of the client's newest message",
		"Arabic numerals",
		"whatsapp-tool freight",    // shipping questions answered without a car
		"Web lookups — RESTRICTED", // web allowed only for the two topics
		"credibility",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q\n%s", want, p)
		}
	}
	if strings.Contains(p, "هذا الرد مولد") {
		t.Error("prompt must not contain the AI disclosure line (it is appended later by us)")
	}
}

// TestSharedRulesSourceOfTruth is the drift guard for the shared client-reply
// rules: it asserts the one rules file exists, carries the load-bearing rules
// (incl. the stock date flags), and is exactly what BuildPrompt injects. If
// someone re-hardcodes a rule into prompt.go or an agent file, this fails.
func TestSharedRulesSourceOfTruth(t *testing.T) {
	path := filepath.Join(config.Root(), "chatbot", "rules", "client-reply.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("shared rules file missing: %v", err)
	}
	text := string(raw)
	for _, want := range []string{
		"whatsapp-tool stock",
		"--this-month", // the stock date flag that started this
		"--status sold",
		"Web lookups — RESTRICTED",
		"exact language of the client's newest message",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("shared rules missing %q", want)
		}
	}
	if got := sharedRules(); got != strings.TrimRight(text, "\n") {
		t.Fatalf("sharedRules() must return the file verbatim; got %d bytes, file %d", len(got), len(text))
	}
}

func TestFinalizeReply(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		max      int
		usedWeb  bool
		want     string
		wantFail bool
	}{
		{name: "plain", raw: "  Salam khouya 🙂  ", max: 400, want: "Salam khouya 🙂\n" + AiDisclosureLine},
		{name: "fenced", raw: "```\nVoilà le prix\n```", max: 400, want: "Voilà le prix\n" + AiDisclosureLine},
		{name: "blank", raw: "   \n```\n```\n  ", max: 400, wantFail: true},
		{name: "truncated", raw: strings.Repeat("a", 50) + " " + strings.Repeat("b", 50), max: 30, want: strings.Repeat("a", 30) + "\n" + AiDisclosureLine},
		{name: "inline backticks", raw: "`ok` bon", max: 400, want: "ok bon\n" + AiDisclosureLine},
		{name: "web sourced", raw: "حسب المعلومات المتاحة", max: 400, usedWeb: true, want: "حسب المعلومات المتاحة\n" + WebSourceLine + "\n" + AiDisclosureLine},
		{name: "web sourced blank", raw: "   ", max: 400, usedWeb: true, wantFail: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FinalizeReply(tc.raw, tc.max, tc.usedWeb)
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
			if tc.usedWeb && !strings.Contains(got, WebSourceLine) {
				t.Fatalf("web-sourced reply must carry the source line: %q", got)
			}
			if !tc.usedWeb && strings.Contains(got, WebSourceLine) {
				t.Fatalf("non-web reply must not carry the source line: %q", got)
			}
		})
	}
}

func TestFinalizeReplyCapsRunesNotBytes(t *testing.T) {
	raw := strings.Repeat("مرحبا ", 50)
	got := FinalizeReply(raw, 30, false)
	if strings.Contains(got, "\ufffd") {
		t.Fatalf("reply contains a broken rune: %q", got)
	}
}

// TestStripReasoningPrefix reproduces the real leak: big-pickle emitted its
// chain-of-thought as the final text part, then the client-facing answer. The
// planning prefix must be dropped while the answer (and a legitimate trailing
// "Let me know…") is kept.
func TestStripReasoningPrefix(t *testing.T) {
	contaminated := "The OpenCorporates fetch didn't find it. But the search results gave me useful info: X is a Hong Kong company.\n\n" +
		"Let me provide a reply in Algerian Derja/Arabic. The client wrote in English actually.\n\n" +
		"Let me compile honest info. I consulted the web.\n\n" +
		"Should I be careful? The name matches the working directory.\n\n" +
		"Let me write it.\n\n" +
		"هذه المعلومة من الإنترنت وغير مؤكدة. الشركة مسجلة في هونغ كونغ.\n\n" +
		"خلاصة القول: تبدو شركة حقيقية. Let me know if you need more."
	got := stripReasoningPrefix(contaminated)
	want := "هذه المعلومة من الإنترنت وغير مؤكدة. الشركة مسجلة في هونغ كونغ.\n\n" +
		"خلاصة القول: تبدو شركة حقيقية. Let me know if you need more."
	if got != want {
		t.Fatalf("strip:\n got %q\nwant %q", got, want)
	}
	if looksLikeReasoning(got) {
		t.Fatalf("clean answer must not look like reasoning: %q", got)
	}

	// A normal reply is untouched.
	clean := "أهلاً! السعر 3,269,810 دج واصل ميناء الجزائر."
	if stripReasoningPrefix(clean) != clean {
		t.Fatalf("clean reply was modified: %q", stripReasoningPrefix(clean))
	}
	// A single "I'll…" opener is not enough to trigger a strip.
	oneCue := "I'll check that for you.\n\nThe price is 3,269,810 DZD."
	if stripReasoningPrefix(oneCue) != oneCue {
		t.Fatalf("single-cue reply should be untouched: %q", stripReasoningPrefix(oneCue))
	}
}

// TestLooksLikeReasoning proves a reply that is entirely the model's
// deliberation (no client-facing answer) is flagged as unusable.
func TestLooksLikeReasoning(t *testing.T) {
	reasoning := "The user is asking to check credibility. Let me first try the lookups.\n\n" +
		"Let me search the web for it.\n\n" +
		"Let me run a few things."
	if !looksLikeReasoning(reasoning) {
		t.Fatal("pure deliberation must be flagged as reasoning")
	}
	if looksLikeReasoning("أهلاً بك، كيفاش نقدر نعاونك؟") {
		t.Fatal("a normal reply must not be flagged")
	}
}
