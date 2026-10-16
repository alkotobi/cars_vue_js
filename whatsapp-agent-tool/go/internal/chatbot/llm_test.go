package chatbot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"whatsapp-agent-tool/internal/config"
)

func TestParseEventsText(t *testing.T) {
	in := `
{"type":"step_start","step":"plan"}
{"type":"text","part":{"type":"text","text":"Salam khouya"},"step":1}
{"type":"text","part":{"type":"text","text":" voilà la réponse"},"step":1}
{"type":"step_finish","step":"plan"}
`
	got, usedWeb, _ := parseEvents([]byte(in))
	want := "Salam khouya voilà la réponse"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if usedWeb {
		t.Fatal("usedWeb should be false for a text-only stream")
	}
}

func TestParseEventsEmpty(t *testing.T) {
	in := `{"type":"step_start","step":"plan"}
{"type":"step_finish","step":"plan"}`
	if got, usedWeb, _ := parseEvents([]byte(in)); got != "" || usedWeb {
		t.Fatalf("want empty, got %q (usedWeb=%v)", got, usedWeb)
	}
}

// TestParseEventsDropsPlanning proves the answer is only the text from the final
// step: narration the model emits between tool calls must never reach the client
// (the "The user is asking… Let me look it up" leak), while a tool-less reply is
// returned whole.
func TestParseEventsDropsPlanning(t *testing.T) {
	plan := func(txt string) string {
		return `{"type":"step_start","part":{"type":"step-start"}}` + "\n" +
			`{"type":"text","part":{"type":"text","text":` + jsonStr(txt) + `}}` + "\n" +
			`{"type":"step_finish","part":{"type":"step-finish","reason":"tool-calls"}}`
	}
	answerStep := func(txt string) string {
		return `{"type":"step_start","part":{"type":"step-start"}}` + "\n" +
			`{"type":"text","part":{"type":"text","text":` + jsonStr(txt) + `}}` + "\n" +
			`{"type":"step_finish","part":{"type":"step-finish","reason":"stop"}}`
	}
	tool := `{"type":"tool_use","part":{"type":"tool","tool":"bash","state":{"input":{"command":"whatsapp-tool faq search \"x\""}}}}`

	in := plan("The user is asking to check credibility. Let me try the lookups.") + "\n" + tool + "\n" +
		plan("Let me run a few things.") + "\n" + tool + "\n" +
		answerStep("هذه المعلومة من الإنترنت، الشركة مسجلة في هونغ كونغ.")
	got, usedWeb, usedTool := parseEvents([]byte(in))
	want := "هذه المعلومة من الإنترنت، الشركة مسجلة في هونغ كونغ."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if !usedTool {
		t.Fatal("usedTool should be true")
	}
	if usedWeb {
		t.Fatal("plain faq bash call must not count as web use")
	}

	// A plain tool-less reply is returned whole (no step control at all).
	if got, _, _ := parseEvents([]byte(`{"type":"text","part":{"type":"text","text":"سلام خويا"}}`)); got != "سلام خويا" {
		t.Fatalf("tool-less reply: got %q", got)
	}
}

// TestParseEventsNoAnswerAfterTools proves a run that narrates a plan and then
// stops without a final answer yields no text (the caller then nudges/defers
// rather than posting the narration).
func TestParseEventsNoAnswerAfterTools(t *testing.T) {
	in := `{"type":"step_start","part":{"type":"step-start"}}` + "\n" +
		`{"type":"text","part":{"type":"text","text":"The user is asking X. Let me look it up."}}` + "\n" +
		`{"type":"tool_use","part":{"type":"tool","tool":"websearch","state":{"input":{}}}}` + "\n" +
		`{"type":"step_finish","part":{"type":"step-finish","reason":"tool-calls"}}`
	got, usedWeb, usedTool := parseEvents([]byte(in))
	if got != "" {
		t.Fatalf("want empty answer after tool-only step, got %q", got)
	}
	if !usedWeb || !usedTool {
		t.Fatalf("want web+tool flags set, got web=%v tool=%v", usedWeb, usedTool)
	}
	if !badAnswer(got, usedTool) {
		t.Fatal("an empty answer after a tool run must be a bad answer")
	}
}

// jsonStr encodes s as a JSON string literal (helper for building event lines).
func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// TestParseEventsDetectsWebTools proves the mandatory source disclosure is
// triggered by the real opencode tool-event shapes: a `websearch`/`webfetch`
// tool_use event and any bash call to `whatsapp-tool web fetch` set usedWeb,
// while other bash calls (e.g. a stock lookup) do not.
func TestParseEventsDetectsWebTools(t *testing.T) {
	bash := `{"type":"tool_use","part":{"type":"tool","tool":"bash","callID":"c1","state":{"status":"completed","input":{"command":"whatsapp-tool stock \"CHANGAN X5 PLUS\""}}}}`
	webFetch := `{"type":"tool_use","part":{"type":"tool","tool":"bash","callID":"c4","state":{"status":"completed","input":{"command":"whatsapp-tool web fetch \"https://en.wikipedia.org/wiki/X\""}}}}`
	search := `{"type":"tool_use","part":{"type":"tool","tool":"websearch","callID":"c2","state":{"status":"completed"}}}`
	fetch := `{"type":"tool_use","part":{"type":"tool","tool":"webfetch","callID":"c3","state":{"status":"completed"}}}`
	text := `{"type":"text","part":{"type":"text","text":"voilà"}}`

	if _, usedWeb, _ := parseEvents([]byte(bash + "\n" + text)); usedWeb {
		t.Fatal("a non-web bash tool call must not count as web use")
	}
	if _, usedWeb, _ := parseEvents([]byte(search + "\n" + text)); !usedWeb {
		t.Fatal("websearch must set usedWeb")
	}
	if _, usedWeb, _ := parseEvents([]byte(fetch + "\n" + text)); !usedWeb {
		t.Fatal("webfetch must set usedWeb")
	}
	if _, usedWeb, _ := parseEvents([]byte(webFetch + "\n" + text)); !usedWeb {
		t.Fatal("a `whatsapp-tool web fetch` bash call must set usedWeb")
	}
}

func TestIsWebFetchCommand(t *testing.T) {
	yes := []string{
		`whatsapp-tool web fetch "https://en.wikipedia.org/wiki/X"`,
		`/opt/bin/whatsapp-tool web fetch https://gsxt.gov.cn`,
		`whatsapp-tool  web   fetch https://qcc.com`,
	}
	for _, c := range yes {
		if !isWebFetchCommand(c) {
			t.Errorf("want web fetch: %q", c)
		}
	}
	no := []string{
		"",
		`whatsapp-tool stock "X"`,
		`whatsapp-tool faq search "web fetch"`,
		`whatsapp-tool web list`,
	}
	for _, c := range no {
		if isWebFetchCommand(c) {
			t.Errorf("false positive: %q", c)
		}
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

// TestAgentWebPolicy pins the web policy in the agent templates: the base
// agent must allow websearch, DISABLE opencode's own webfetch (1.x cannot scope
// it to a domain list), and route page reads through the allowlisted
// `whatsapp-tool web fetch` (enforced in Go, internal/webfetch). The editor
// agent must have NO web access at all — its docs edits must never follow
// fetched content.
func TestAgentWebPolicy(t *testing.T) {
	base, err := os.ReadFile(filepath.Join(config.Root(), "chatbot", "agent", "chatbot.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"websearch: allow",
		"webfetch: deny",
		"whatsapp-tool web fetch",
	} {
		if !strings.Contains(string(base), want) {
			t.Errorf("base agent missing web policy line %q", want)
		}
	}

	editor, err := os.ReadFile(filepath.Join(config.Root(), "chatbot", "agent", "chatbot-editor.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The editor agent runs for groups with the edit capability (e.g. "ai group"),
	// which is where real client questions land — it must carry the same restricted
	// web policy, or a credibility/spec question auto-rejects websearch and the
	// reply is dropped with no text.
	for _, want := range []string{
		"websearch: allow",
		"webfetch: deny",
		"whatsapp-tool web fetch",
	} {
		if !strings.Contains(string(editor), want) {
			t.Errorf("editor agent missing web policy line %q", want)
		}
	}
	// ...while still being the only agent that may edit docs.
	if !strings.Contains(string(editor), "edit:") {
		t.Error("editor agent must keep its edit permission")
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
