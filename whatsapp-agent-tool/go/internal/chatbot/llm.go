// The opencode CLI as the reply generator. To keep the model honest we run it
// in an isolated project dir (<workdir>) whose config disables the WhatsApp MCP
// and whose agent allows only bash restricted by permission pattern to
// read-only `whatsapp-tool …` lookups (pricing/FAQ/stock), the built-in
// websearch tool, and `whatsapp-tool web fetch` — an allowlisted page reader
// (internal/webfetch) for company credibility + car specs only. opencode's own
// webfetch is disabled: 1.x cannot scope it to a domain list. Text in, text
// out — the model can quote the company DB or a vetted public page but cannot
// touch anything else. The files are mirrored from the repo's chatbot/ template
// at each start so the prompt stays editable in-repo while the runtime dir stays
// out of it.
package chatbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/groupauth"
)

// llmTimeout bounds one reply generation (local 27B models are slow).
const llmTimeout = 240 * time.Second

// WorkDir returns the isolated opencode project dir the bot runs in.
func WorkDir() (string, error) {
	if d := config.C.Chatbot.WorkDir; d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("chat: resolve home dir: %w", err)
	}
	return filepath.Join(home, ".whatsapp-tool", "chatbot"), nil
}

// ensureConfig copies the repo chatbot/ template (opencode.json + agent file)
// into the runtime workdir and applies the CHAT_MODEL override if set. The
// agent must land in .opencode/agent/ (hidden dir, opencode's on-disk agent
// location) — a plain agent/ folder is silently ignored and opencode then
// falls back to its DEFAULT agent, which is tool-enabled. The workdir is
// deliberately OUTSIDE this repo: opencode loads the nearest project config
// and AGENTS.md walking upward, so a dir inside whatsapp-agent-tool would boot
// the tool's own MCP config and obey its AGENTS.md instead of ours.
func ensureConfig() (string, error) {
	work, err := WorkDir()
	if err != nil {
		return "", err
	}
	agentDir := filepath.Join(work, ".opencode", "agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		return "", fmt.Errorf("chat: create workdir: %w", err)
	}
	src := filepath.Join(config.Root(), "chatbot")
	if err := copyFile(filepath.Join(src, "opencode.json"), filepath.Join(work, "opencode.json")); err != nil {
		return "", fmt.Errorf("chat: workdir opencode.json: %w", err)
	}
	if err := copyFile(filepath.Join(src, "agent", "chatbot.md"), filepath.Join(agentDir, "chatbot.md")); err != nil {
		return "", fmt.Errorf("chat: workdir agent: %w", err)
	}
	// The editor agent is the same sandbox plus targeted edit/write of repo docs;
	// it is used only for groups granted can_give_edit_instruction (see perms.go).
	// The docs dir is substituted so the allowlist glob is anchored to this repo.
	if err := copyEditorAgent(filepath.Join(src, "agent", "chatbot-editor.md"), filepath.Join(agentDir, "chatbot-editor.md"), config.Root()); err != nil {
		return "", fmt.Errorf("chat: workdir editor agent: %w", err)
	}
	// Remove the legacy agent/ dir from earlier builds — it did nothing and
	// only confused readers. Never touches .opencode/.
	_ = os.RemoveAll(filepath.Join(work, "agent"))
	if m := config.C.Chatbot.Model; m != "" {
		cfg := filepath.Join(work, "opencode.json")
		raw, err := os.ReadFile(cfg)
		if err != nil {
			return "", err
		}
		// Replace whatever model the template pins with the CHAT_MODEL override.
		next := modelLineRE.ReplaceAllString(string(raw), fmt.Sprintf(`"model": %q`, m))
		if next != string(raw) {
			if err := os.WriteFile(cfg, []byte(next), 0o644); err != nil {
				return "", fmt.Errorf("chat: apply model override: %w", err)
			}
		}
	}
	return work, nil
}

// modelLineRE matches the top-level `"model": "…"` value in opencode.json.
var modelLineRE = regexp.MustCompile(`"model"\s*:\s*"[^"]*"`)

// Generate runs the tool-enabled opencode agent with the prompt and returns its
// plain-text reply (the answer from its final step — see parseEvents) plus
// whether the model consulted the web (websearch/webfetch) — the caller uses that
// flag to stamp the mandatory source disclosure. The child goes through /bin/sh,
// not a direct exec: opencode's embedded server errors out ("Unexpected server
// error") when the CLI is spawned straight by a non-shell parent, while the
// same command through a shell always works. The prompt rides in an env var so
// no shell quoting can mangle it (it contains newlines, apostrophes and glob
// characters).
//
// A reply is rejected (and never posted) when it is not a client-facing sentence:
// either the model printed the tool invocation as its answer text (e.g.
// `whatsapp-tool stock "X"`) or it ran tools but produced no final answer at all
// (it narrated its plan and stopped). Either way it is retried once with an
// explicit nudge; if it is still unusable we post the safe holding message rather
// than leak the narration.
func Generate(prompt string, perms Perms) (string, bool, error) {
	work, err := ensureConfig()
	if err != nil {
		return "", false, err
	}
	bin := config.C.Chatbot.OpenCodeBin
	if resolved, err := exec.LookPath(bin); err == nil {
		bin = resolved
	}
	text, usedWeb, usedTool, err := runOpenCode(work, bin, prompt, perms)
	if err != nil {
		return "", false, err
	}
	text = stripReasoningPrefix(text)
	if badAnswer(text, usedTool) {
		nudge := prompt + "\n\nIMPORTANT: invoke the whatsapp-tool command through your tool interface; " +
			"NEVER print the command itself as your reply, and never describe what you are about to do. " +
			"Reply with the client-facing sentence only."
		if text2, web2, tool2, err2 := runOpenCode(work, bin, nudge, perms); err2 == nil {
			if text2 = stripReasoningPrefix(text2); !badAnswer(text2, tool2) {
				return strings.TrimSpace(text2), web2, nil
			}
		}
		return DeferralReply, false, nil
	}
	return text, usedWeb, nil
}

// badAnswer reports whether text cannot be posted to a client: it is empty after
// the model used tools (it narrated a plan and stopped without answering), it is
// a bare tool invocation rather than a sentence, or it is dominated by the
// model's own reasoning rather than a reply.
func badAnswer(text string, usedTool bool) bool {
	if strings.TrimSpace(text) == "" {
		return usedTool
	}
	return looksLikeToolDump(text) || looksLikeReasoning(text)
}

// runOpenCode executes one opencode generation under llmTimeout and returns the
// final answer text, whether a web tool was used, and whether any tool ran, or
// an error if it failed or produced nothing. An empty answer after a tool run is
// NOT an error — the caller decides (see badAnswer); only a run with no text and
// no tool activity is treated as a failure. The group's capabilities ride to the
// child in WBOT_PERMS (always set, possibly empty) and pick the agent — this is
// the only place the bot's permissions are enforced.
func runOpenCode(work, bin, prompt string, perms Perms) (string, bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), llmTimeout)
	defer cancel()

	promptEnv := "WBOT_PROMPT"
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c",
		`exec `+shQuote(bin)+` run --agent "$1" --pure --format json "$`+promptEnv+`" </dev/null`,
		"chatbot", perms.AgentName())
	cmd.Dir = work
	// The agent may run `whatsapp-tool …` lookups: point it at the repo's
	// bin/ on PATH and let the CLI find .env via WA_ROOT_DIR (the workdir is
	// outside the repo, so its own upward .env search would miss). WBOT_PERMS is
	// always set so the CLI runs restricted for this group. OPENCODE_ENABLE_EXA
	// turns on the built-in websearch tool even for non-OpenCode providers (it
	// is harmless with the default big-pickle provider, which already has it).
	repo := config.Root()
	pathEnv := filepath.Join(repo, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	cmd.Env = append(os.Environ(),
		"NO_COLOR=1",
		promptEnv+"="+prompt,
		groupauth.EnvVar+"="+perms.EnvValue(),
		"WA_ROOT_DIR="+repo,
		"PATH="+pathEnv,
		"OPENCODE_ENABLE_EXA=1",
	)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	runErr := cmd.Run()
	text, usedWeb, usedTool := parseEvents(out.Bytes())
	if text != "" || usedTool {
		return text, usedWeb, usedTool, nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return "", false, false, fmt.Errorf("chat: reply generation timed out after %s", llmTimeout)
	}
	// No text: opencode failed. Prefer the structured reason it printed (so a
	// rate limit reads as such, not a bare "exit status 1").
	if reason := apiErrorReason(out.Bytes()); reason != "" {
		return "", false, false, fmt.Errorf("chat: opencode: %s", reason)
	}
	if runErr != nil {
		return "", false, false, fmt.Errorf("chat: opencode failed: %v%s", runErr, tailStderr(errb.String()))
	}
	return "", false, false, fmt.Errorf("chat: no text in opencode output%s", tailStderr(errb.String()))
}

// apiErrorReason extracts a short, speech-friendly reason from the `type:
// "error"` event opencode prints on failure (e.g. a 429 rate limit or the
// big-pickle free-tier 403). Returns "" when there is no such event.
func apiErrorReason(stdout []byte) string {
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var ev struct {
			Type  string `json:"type"`
			Error struct {
				Data struct {
					Message      string `json:"message"`
					StatusCode   int    `json:"statusCode"`
					ResponseBody string `json:"responseBody"`
				} `json:"data"`
			} `json:"error"`
		}
		if json.Unmarshal(line, &ev) != nil || ev.Type != "error" {
			continue
		}
		d := ev.Error.Data
		blob := strings.ToLower(d.Message + " " + d.ResponseBody)
		switch {
		case d.StatusCode == 429 || strings.Contains(blob, "rate limit") || strings.Contains(blob, "ratelimit"):
			return "rate limit reached (HTTP 429) — the model is throttling us"
		case d.StatusCode == 403 && strings.Contains(blob, "freetier"):
			return "model access denied (HTTP 403 free tier)"
		case d.StatusCode != 0:
			return fmt.Sprintf("model API error (HTTP %d): %s", d.StatusCode, truncateRunes(strings.TrimSpace(d.Message), 120))
		case strings.TrimSpace(d.Message) != "":
			return "model error: " + truncateRunes(strings.TrimSpace(d.Message), 120)
		default:
			return "opencode reported an error"
		}
	}
	return ""
}

// looksLikeToolDump reports whether a model reply is a raw command invocation
// rather than a client-facing sentence — its first non-empty line starts with
// the whatsapp-tool binary. Belt-and-braces guard for the model's occasional
// "tool call as text" slip (see Generate).
func looksLikeToolDump(s string) bool {
	first := strings.TrimSpace(s)
	if i := strings.IndexByte(first, '\n'); i >= 0 {
		first = strings.TrimSpace(first[:i])
	}
	return strings.HasPrefix(first, "whatsapp-tool ") || strings.HasPrefix(first, "./whatsapp-tool ")
}

// shQuote wraps s in single quotes, safe for embedding in a /bin/sh script
// (assumes no single-quote char — fine for binary paths).
func shQuote(s string) string {
	return "'" + s + "'"
}

// parseEvents scans opencode's --format json stream and returns the model's
// FINAL answer text, whether it consulted the web anywhere in the run, and
// whether it invoked any tool. The web flag is set by the built-in `websearch`
// tool AND by any bash call to `whatsapp-tool web fetch …` (opencode's own
// webfetch is disabled, so the allowlisted reader is reached through bash — see
// internal/webfetch).
//
// Only the text emitted AFTER the last tool call is the answer. A model that
// "thinks out loud" emits its planning as ordinary text parts between tool calls
// (e.g. "The user is asking… Let me look it up."); joining every text part would
// leak that narration to the client. So the text buffer is reset whenever a tool
// is used, and a step that ends by calling tools (`reason:"tool-calls"`) never
// contributes. The final step (`reason:"stop"`), or a tool-less reply, yields
// the answer. Verified against opencode 1.19.0: step_start / step_finish frame
// each step, {"type":"text","part":{"type":"text","text":"…"}} is text and
// {"type":"tool_use","part":{"type":"tool","tool":"bash",…}} is a tool.
func parseEvents(stdout []byte) (string, bool, bool) {
	var cur strings.Builder
	var final string
	usedWeb, usedTool := false, false
	for _, line := range bytes.Split(bytes.TrimSpace(stdout), []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var ev struct {
			Type string `json:"type"`
			Part struct {
				Type string `json:"type"`
				Text string `json:"text"`
				Tool string `json:"tool"`
				// step-finish carries the step outcome ("tool-calls"/"stop").
				Reason string `json:"reason"`
				State  struct {
					Input struct {
						Command string `json:"command"`
					} `json:"input"`
				} `json:"state"`
			} `json:"part"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		switch {
		case ev.Type == "step_start":
			// A new step begins: discard any text not already accepted as final.
			cur.Reset()
		case ev.Type == "text" && ev.Part.Type == "text":
			cur.WriteString(ev.Part.Text)
		case ev.Part.Type == "tool":
			usedTool = true
			if isWebTool(ev.Part.Tool) || isWebFetchCommand(ev.Part.State.Input.Command) {
				usedWeb = true
			}
			// Text before a tool call is planning, not the answer.
			cur.Reset()
		case ev.Type == "step_finish":
			if reason := strings.TrimSpace(ev.Part.Reason); reason == "" || reason == "stop" {
				if t := strings.TrimSpace(cur.String()); t != "" {
					final = cur.String()
				}
			}
			cur.Reset()
		}
	}
	if strings.TrimSpace(final) == "" {
		return cur.String(), usedWeb, usedTool
	}
	return final, usedWeb, usedTool
}

// isWebTool reports whether an opencode tool name is one of the web tools the
// chat agent may use — whose results must therefore be disclosed to the client.
// (`webfetch` is disabled in the agent, but a stray attempt is still counted so
// the disclosure can never be skipped.)
func isWebTool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "websearch", "webfetch":
		return true
	default:
		return false
	}
}

// isWebFetchCommand reports whether a bash command reads a page through the
// allowlisted reader, i.e. contains `web fetch` (`whatsapp-tool web fetch …`).
func isWebFetchCommand(cmd string) bool {
	if cmd == "" {
		return false
	}
	return strings.Contains(" "+strings.Join(strings.Fields(cmd), " ")+" ", " web fetch ")
}

func tailStderr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 3 {
		lines = lines[len(lines)-3:]
	}
	return "\nstderr: " + strings.Join(lines, "\n")
}

// copyFile copies src to dst, preserving no ownership — deploy/ownership rules
// for the app's web roots do not apply to a local runtime dir.
func copyFile(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, raw, 0o644)
}

// copyEditorAgent copies the editor agent, substituting __DOCS_DIR__ with the
// repo root so its edit/write allowlist glob is anchored to the docs in THIS
// repo rather than any markdown on the machine.
func copyEditorAgent(src, dst, docsDir string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	subbed := strings.ReplaceAll(string(raw), "__DOCS_DIR__", filepath.ToSlash(docsDir))
	return os.WriteFile(dst, []byte(subbed), 0o644)
}
