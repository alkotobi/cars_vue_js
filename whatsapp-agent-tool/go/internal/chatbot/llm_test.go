package chatbot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"whatsapp-agent-tool/internal/config"
)

func TestExtractTextEvents(t *testing.T) {
	in := `
{"type":"step_start","step":"plan"}
{"type":"text","part":{"type":"text","text":"Salam khouya"},"step":1}
{"type":"text","part":{"type":"text","text":" voilà la réponse"},"step":1}
{"type":"step_finish","step":"plan"}
`
	got := extractTextEvents([]byte(in))
	want := "Salam khouya voilà la réponse"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExtractTextEventsEmpty(t *testing.T) {
	in := `{"type":"step_start","step":"plan"}
{"type":"step_finish","step":"plan"}`
	if got := extractTextEvents([]byte(in)); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestLooksLikeToolDump(t *testing.T) {
	dumps := []string{
		`whatsapp-tool stock "CHANGAN X5 PLUS"`,
		"whatsapp-tool stock \"CHANGAN X5 PLUS\"\nهذا الرد مولد بالذكاء الاصطناعي",
		"  ./whatsapp-tool quote --car \"X\"  ",
		"\nwhatsapp-tool faq search \"x\"\n",
	}
	for _, s := range dumps {
		if !looksLikeToolDump(s) {
			t.Errorf("want dump detected: %q", s)
		}
	}
	ok := []string{
		"",
		"أهلاً، السعر 4,389,652 دج.",
		"Sure, let me check that for you — whatsapp-tool has the data.",
		"لا، السعر لا يشمل TVN.",
	}
	for _, s := range ok {
		if looksLikeToolDump(s) {
			t.Errorf("false positive: %q", s)
		}
	}
}

func TestAPIErrorReason(t *testing.T) {
	freeTier := `{"type":"error","error":{"name":"APIError","data":{"message":"OpenCode's free tier can only be used from within OpenCode.","statusCode":403,"responseBody":"{\"type\":\"error\",\"error\":{\"type\":\"FreeTierError\"}}"}}}`
	rateLimit := `{"type":"error","error":{"name":"APIError","data":{"message":"Rate limit exceeded, please retry later","statusCode":429}}}`
	generic := `{"type":"error","error":{"data":{"message":"boom","statusCode":500}}}`
	textLine := `{"type":"text","part":{"type":"text","text":"hi"}}`

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"rate limit by code", rateLimit, "rate limit reached (HTTP 429) — the model is throttling us"},
		{"free tier 403", freeTier, "model access denied (HTTP 403 free tier)"},
		{"generic 500", generic, "model API error (HTTP 500): boom"},
		{"no error event", textLine, ""},
		{"empty", "", ""},
	}
	for _, c := range cases {
		if got := apiErrorReason([]byte(c.in)); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// TestEnsureConfigInstallsAgents proves the runtime workdir gets BOTH agents,
// with the editor's __DOCS_DIR__ placeholder replaced by a real path — a plain
// `agent/` dir or a missing editor agent would make `--agent <name>-editor`
// fail for any group granted can_give_edit_instruction.
func TestEnsureConfigInstallsAgents(t *testing.T) {
	old := config.C.Chatbot.WorkDir
	defer func() { config.C.Chatbot.WorkDir = old }()
	config.C.Chatbot.WorkDir = t.TempDir()

	work, err := ensureConfig()
	if err != nil {
		t.Fatalf("ensureConfig: %v", err)
	}
	agentDir := filepath.Join(work, ".opencode", "agent")
	for _, name := range []string{"chatbot.md", "chatbot-editor.md"} {
		if _, err := os.Stat(filepath.Join(agentDir, name)); err != nil {
			t.Fatalf("agent %s not installed: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(work, "opencode.json")); err != nil {
		t.Fatalf("opencode.json not installed: %v", err)
	}
	editor, err := os.ReadFile(filepath.Join(agentDir, "chatbot-editor.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(editor), "__DOCS_DIR__") {
		t.Fatal("editor agent left __DOCS_DIR__ unsubstituted")
	}
	if !strings.Contains(string(editor), filepath.ToSlash(config.Root())+"/**/*.md") {
		t.Fatalf("editor allowlist not anchored to repo docs (%s)", config.Root())
	}
}
