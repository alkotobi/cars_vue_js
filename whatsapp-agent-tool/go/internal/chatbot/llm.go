// The opencode CLI as the reply generator. To keep the model honest we run it
// in an isolated project dir (<workdir>) whose config disables the WhatsApp MCP
// and whose agent allows exactly one tool: bash, restricted by permission
// pattern to read-only `whatsapp-tool …` lookups (pricing/FAQ/stock). Text in,
// text out — the model can quote the company DB but cannot touch anything else.
// The files are mirrored from the repo's chatbot/ template at each start so the
// prompt stays editable in-repo while the runtime dir stays out of it.
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
// plain-text reply (the join of every text part it emitted). The child goes
// through /bin/sh, not a direct exec: opencode's embedded server errors out
// ("Unexpected server error") when the CLI is spawned straight by a non-shell
// parent, while the same command through a shell always works. The prompt
// rides in an env var so no shell quoting can mangle it (it contains newlines,
// apostrophes and glob characters).
//
// The local 27B model occasionally writes the tool invocation as its answer
// text instead of calling the tool (e.g. it prints `whatsapp-tool stock "X"`).
// That must never reach a client, so a reply that looks like a bare command is
// retried once with an explicit nudge; if it is still a command we post the
// safe holding message rather than leak it.
func Generate(prompt string, perms Perms) (string, error) {
	work, err := ensureConfig()
	if err != nil {
		return "", err
	}
	bin := config.C.Chatbot.OpenCodeBin
	if resolved, err := exec.LookPath(bin); err == nil {
		bin = resolved
	}
	text, err := runOpenCode(work, bin, prompt, perms)
	if err != nil {
		return "", err
	}
	if looksLikeToolDump(text) {
		nudge := prompt + "\n\nIMPORTANT: invoke the whatsapp-tool command through your tool interface; " +
			"NEVER print the command itself as your reply. Reply with the client-facing sentence only."
		if text2, err2 := runOpenCode(work, bin, nudge, perms); err2 == nil && strings.TrimSpace(text2) != "" && !looksLikeToolDump(text2) {
			return text2, nil
		}
		return DeferralReply, nil
	}
	return text, nil
}

// runOpenCode executes one opencode generation under llmTimeout and returns the
// joined text parts, or an error if it failed or produced no text. The group's
// capabilities ride to the child in WBOT_PERMS (always set, possibly empty) and
// pick the agent — this is the only place the bot's permissions are enforced.
func runOpenCode(work, bin, prompt string, perms Perms) (string, error) {
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
	// always set so the CLI runs restricted for this group.
	repo := config.Root()
	pathEnv := filepath.Join(repo, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	cmd.Env = append(os.Environ(),
		"NO_COLOR=1",
		promptEnv+"="+prompt,
		groupauth.EnvVar+"="+perms.EnvValue(),
		"WA_ROOT_DIR="+repo,
		"PATH="+pathEnv,
	)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	runErr := cmd.Run()
	if text := extractTextEvents(out.Bytes()); text != "" {
		return text, nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("chat: reply generation timed out after %s", llmTimeout)
	}
	// No text: opencode failed. Prefer the structured reason it printed (so a
	// rate limit reads as such, not a bare "exit status 1").
	if reason := apiErrorReason(out.Bytes()); reason != "" {
		return "", fmt.Errorf("chat: opencode: %s", reason)
	}
	if runErr != nil {
		return "", fmt.Errorf("chat: opencode failed: %v%s", runErr, tailStderr(errb.String()))
	}
	return "", fmt.Errorf("chat: no text in opencode output%s", tailStderr(errb.String()))
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

// extractTextEvents joins every `part.text` of `type == "text"` events from
// opencode's --format json stream. A single reply usually lands as one text
// event, but a multi-turn response can emit several — they are joined so the
// caller always gets one message.
func extractTextEvents(stdout []byte) string {
	var parts []string
	for _, line := range bytes.Split(bytes.TrimSpace(stdout), []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var ev struct {
			Type string `json:"type"`
			Part struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"part"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		if ev.Type == "text" && ev.Part.Type == "text" && strings.TrimSpace(ev.Part.Text) != "" {
			parts = append(parts, ev.Part.Text)
		}
	}
	return strings.Join(parts, "")
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
