// Prompt building and reply shaping for the chat AI. Pure string work, fully
// unit-testable. The prompt is assembled from real chat rows only (the model
// never sees SQL), and the final reply always carries the AI disclosure line
// appended deterministically by us — never glued on by the model.
package chatbot

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"whatsapp-agent-tool/internal/config"
)

// AiDisclosureLine is appended to every outgoing AI reply (AGENTS.md §0.9).
const AiDisclosureLine = "هذا الرد مولد بالذكاء الاصطناعي"

// DeferralReply is the safe holding message posted when the model cannot be
// trusted to produce a client-facing sentence (AGENTS.md §13, MISSING/AMBIGUOUS).
const DeferralReply = "سأطلب من المدير وأعود إليك."

// WebSourceLine is appended (before the AI disclosure line) to any reply the
// model produced after consulting the web, so the client always sees that the
// information came from the internet and is unverified. It is added
// deterministically by us — the model cannot be trusted to remember it.
const WebSourceLine = "🌐 هذه المعلومة من الإنترنت (مصدر عمومي) — يُرجى التحقق منها."

// historyLimit is how many recent messages (any sender) the model sees as
// conversational context before the client's new lines.
const historyLimit = 12

// BuildPrompt composes the single text the opencode agent is asked to reply on.
// history is the tail of the group (oldest first), incoming are the new
// messages (oldest first) from anyone except the AI account itself. perms tells
// the model what this group is allowed to do, so it can phrase a short refusal
// for a request it must hand to the team.
func BuildPrompt(groupName, aiUsername string, perms Perms, history, incoming []Message) string {
	var b strings.Builder
	fmt.Fprintf(&b,
		"You are the car-dealership assistant replying inside the dealer platform's chat "+
			"as the user %q, in the group %q.\n", aiUsername, groupName)
	b.WriteString("The AI disclosure line is appended automatically after your reply — never repeat it yourself.\n")
	b.WriteString(capabilityLine(perms))
	if len(history) > 0 {
		b.WriteString("\nRecent conversation (oldest first):\n")
		for _, m := range history {
			b.WriteString(formatMessage(m))
		}
	}
	b.WriteString("\nNew messages to answer (oldest first):\n")
	for _, m := range incoming {
		b.WriteString(formatMessage(m))
	}
	// Channel-agnostic client-reply rules come from ONE file
	// (chatbot/rules/client-reply.md) — the same text injected into both agent
	// personas by ensureConfig. Editing a rule or a CLI flag is then a single
	// edit, not a hand-copy across prompt.go, chatbot.md and chatbot-editor.md.
	b.WriteString("\n")
	b.WriteString(sharedRules())
	return b.String()
}

// sharedRules returns the shared client-reply rules, read once per process from
// chatbot/rules/client-reply.md (relative to the repo root). On a read failure
// it logs and returns a short safe stub rather than an empty rule set — the
// agent still gets the language, no-invented-price, table and no-reasoning
// baselines. The drift test in llm_test.go fails if the file is missing, so the
// stub is a runtime safety net, not a silent fallback.
var sharedRules = sync.OnceValue(func() string {
	path := filepath.Join(config.Root(), "chatbot", "rules", "client-reply.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Printf("chat: shared rules %s unreadable (%v) — using safe fallback", path, err)
		return fallbackRules
	}
	text := strings.TrimRight(string(raw), "\n")
	if strings.TrimSpace(text) == "" {
		log.Printf("chat: shared rules %s is empty — using safe fallback", path)
		return fallbackRules
	}
	return text
})

// fallbackRules is the minimal safe rule set used only if the shared rules file
// cannot be read at runtime. Keep it short; it is not the source of truth.
const fallbackRules = "Reply in the exact language of the client's newest message. " +
	"Never state a price, stock count, availability or delivery figure a command did not return — " +
	"if you cannot look it up, say you will pass the question to the team. " +
	"Present any list of more than one thing as a GitHub-style Markdown table. " +
	"Do NOT include your reasoning, chain-of-thought, or internal thoughts — output only the final client-facing answer."

// formatMessage renders one row for the model.
func formatMessage(m Message) string {
	sender := strings.TrimSpace(m.SenderName)
	role := strings.TrimSpace(m.SenderRole)
	if role == "" || sender == "" {
		sender = "unknown"
	}
	return fmt.Sprintf("%s <%s> %s:\n%s\n", m.Time, sender, role, m.Text)
}

// capabilityLine tells the model, in one deterministic sentence, which
// capabilities this group has. It is advisory only — the CLI enforces the real
// limit — but it lets the model phrase a short in-language refusal instead of
// attempting a call that will be rejected.
func capabilityLine(p Perms) string {
	allowed := make([]string, 0, 3)
	denied := make([]string, 0, 3)
	add := func(label string, ok bool) {
		if ok {
			allowed = append(allowed, label)
		} else {
			denied = append(denied, label)
		}
	}
	add("giving edit instructions to the app", p.Edit)
	add("seeing supplier names and contacts", p.Suppliers)
	add("seeing our cost/benefit figures", p.Cost)

	var b strings.Builder
	b.WriteString("This group's capabilities: ")
	if len(allowed) == 0 {
		b.WriteString("none")
	} else {
		b.WriteString("may " + strings.Join(allowed, "; may "))
	}
	b.WriteString(".")
	if len(denied) > 0 {
		b.WriteString(" It may NOT " + strings.Join(denied, " and may not ") + ".")
	}
	b.WriteString(" If a request needs a capability the group lacks, do NOT attempt it — reply briefly in the writer's language that you will pass it to the team.\n")
	return b.String()
}

// FinalizeReply trims a model reply, drops any markdown fences, enforces the
// character cap, and appends the source line (when the model consulted the web)
// and the AI disclosure line. Returns "" for a blank reply (caller then skips
// posting). The source line is inserted BEFORE the AI line so the message always
// ends with the AI disclosure.
//
// Two client-reply rules are enforced HERE instead of being trusted to the
// model: Arabic-Indic digits become ASCII (§1/§8) and the ONE money figure in
// the message is rounded UP to the client grid (§0.15: 10,000 DZD / 100 USD).
// A message carrying more than one money figure is left untouched — a
// breakdown's exact numbers must survive (§0.14/§3.8).
func FinalizeReply(raw string, maxChars int, usedWeb bool) string {
	text := strings.TrimSpace(raw)
	text = stripCodeFences(text)
	if text == "" {
		return ""
	}
	text = normalizeDigits(text)
	text = roundUpMoney(text)
	if maxChars > 0 {
		text = truncateRunes(text, maxChars)
	}
	if usedWeb {
		return text + "\n" + WebSourceLine + "\n" + AiDisclosureLine
	}
	return text + "\n" + AiDisclosureLine
}

// normalizeDigits maps the Arabic-Indic (٠-٩) and Extended Arabic-Indic/Persian
// (۰-۹) digit blocks onto ASCII 0-9 (client-reply.md §1/§8).
func normalizeDigits(s string) string {
	return digitReplacer.Replace(s)
}

var digitReplacer = strings.NewReplacer(
	"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
	"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
	"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
	"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
)

// Money figures are a number anchored by a currency token — "2,801,020 دج",
// "2 801 020 دينار", "$2,801,020", "2801020 usd". A bare number is never
// touched (years, quantities, phone fragments).
type moneyPattern struct {
	re        *regexp.Regexp
	numGroup  int // submatch group holding the number
	suffixGrp int // submatch group holding a trailing currency token (0 = none)
	dollar    bool
	grid      float64
}

var moneyPatterns = []moneyPattern{
	{re: regexp.MustCompile(`(?i)([0-9][0-9.,\x{00a0}\x{202f} ]{2,})\s*(دج|دينار|د\.ج|dzd)`), numGroup: 1, suffixGrp: 2, grid: 10000},
	{re: regexp.MustCompile(`([$])\s*([0-9][0-9.,\x{00a0}\x{202f} ]{2,})`), numGroup: 2, dollar: true, grid: 100},
	{re: regexp.MustCompile(`(?i)([0-9][0-9.,\x{00a0}\x{202f} ]{2,})\s*(usd|dollars?|us\$)`), numGroup: 1, suffixGrp: 2, grid: 100},
}

// roundUpMoney rewrites the single money figure in text up to its grid
// (10,000 DZD / 100 USD, comma-grouped). Any message with zero — or two or
// more — currency-anchored figures is left exactly as written: a breakdown's
// numbers must survive (client-reply.md §0.14/§3.8).
func roundUpMoney(text string) string {
	type hit struct {
		start, end int
		val        float64
		grid       float64
		dollar     bool
		suffix     string
	}
	anchors := 0
	var hits []hit
	for _, p := range moneyPatterns {
		anchors += len(p.re.FindAllStringIndex(text, -1))
		for _, m := range p.re.FindAllStringSubmatchIndex(text, -1) {
			v, ok := parseMoneyNum(text[m[2*p.numGroup]:m[2*p.numGroup+1]])
			if !ok {
				continue
			}
			h := hit{start: m[0], end: m[1], val: v, grid: p.grid, dollar: p.dollar}
			if p.suffixGrp > 0 {
				h.suffix = text[m[2*p.suffixGrp]:m[2*p.suffixGrp+1]]
			}
			hits = append(hits, h)
		}
	}
	if anchors != 1 || len(hits) != 1 {
		return text
	}
	h := hits[0]
	// The number class can swallow a trailing separator ("$2801020 ") — keep it.
	end := h.end
	for end > h.start && (text[end-1] == ' ' || text[end-1] == '\t') {
		end--
	}
	grouped := formatGrouped(math.Ceil(h.val/h.grid) * h.grid)
	switch {
	case h.dollar:
		return text[:h.start] + "$" + grouped + text[end:]
	case h.suffix != "":
		return text[:h.start] + grouped + " " + h.suffix + text[end:]
	default:
		return text[:h.start] + grouped + text[end:]
	}
}

// parseMoneyNum reads a grouped figure as a number. The LAST separator decides:
// a group of 3 digits after it is a thousands mark ("2,801,020" / "2.801.020"),
// anything else is a decimal point ("2,801,020.25") — so a lone exact figure
// can still be rounded up to the client grid.
func parseMoneyNum(s string) (float64, bool) {
	trimmed := strings.TrimSpace(s)
	intPart, frac := trimmed, ""
	if i := strings.LastIndexAny(trimmed, ",."); i >= 0 && len(trimmed)-i-1 != 3 {
		intPart, frac = trimmed[:i], trimmed[i+1:]
		if frac == "" {
			return 0, false
		}
	}
	strip := func(r rune) rune {
		switch r {
		case ',', '.', ' ', '\u00a0', '\u202f':
			return -1
		}
		return r
	}
	clean := strings.Map(strip, intPart)
	if frac != "" {
		clean += "." + strings.Map(strip, frac)
	}
	if clean == "" || clean == "." {
		return 0, false
	}
	v, err := strconv.ParseFloat(clean, 64)
	if err != nil || v <= 0 {
		return 0, false
	}
	return v, true
}

// formatGrouped renders an integer with thousands separators ("2,810,000").
func formatGrouped(v float64) string {
	s := strconv.FormatInt(int64(math.Round(v)), 10)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// reasoningCues are the paragraph openers that mark a model's private planning
// ("thinking out loud") rather than a client-facing sentence. Some models (e.g.
// opencode/big-pickle) put that deliberation inside the final text part, so the
// daemon has to strip it before posting.
var reasoningCues = []string{
	"let me", "let's", "i'll ", "i will ", "i should ", "i need to ", "i must ",
	"i want to ", "i think i", "i can't ", "i couldn't ", "now i ", "so i ",
	"then i ", "ok,", "okay", "alright", "hmm", "actually,", "wait,",
	"should i ", "maybe i ", "the user ", "the user's", "the client ",
	"the message ", "the instruction", "the request ", "i'm going to",
	"i should probably", "let me write", "but wait",
}

// reasoningMinParagraphs is how many reasoning-flagged paragraphs must appear
// before we treat a reply as contaminated. A single "I'll…" opener is normal
// prose; a block of them is the model thinking out loud.
const reasoningMinParagraphs = 3

// isReasoningParagraph reports whether a paragraph opens like the model's own
// deliberation rather than a reply addressed to the client.
func isReasoningParagraph(p string) bool {
	l := strings.ToLower(strings.TrimSpace(p))
	for _, cue := range reasoningCues {
		if strings.HasPrefix(l, cue) {
			return true
		}
	}
	return false
}

// stripReasoningPrefix drops a leading "thinking out loud" block from a model
// reply. Some models emit their private reasoning as ordinary paragraphs before
// the real answer (same text part), which would leak the narration to the
// client. The answer is taken as the text after the last reasoning paragraph
// (never the final paragraph, so a client-facing "Let me know…" tail is not a
// cut point). The strip only applies when at least reasoningMinParagraphs
// paragraphs look like reasoning and something non-reasoning follows — otherwise
// the text is returned unchanged (ordinary prose is not worth hacking apart).
func stripReasoningPrefix(text string) string {
	paras := splitParagraphs(text)
	if len(paras) < 2 {
		return text
	}
	last, count := -1, 0
	for i := 0; i < len(paras)-1; i++ { // exclude the final paragraph
		if isReasoningParagraph(paras[i]) {
			last, count = i, count+1
		}
	}
	if last < 0 || count < reasoningMinParagraphs {
		return text
	}
	kept := strings.TrimSpace(strings.Join(paras[last+1:], "\n\n"))
	if kept == "" || isReasoningParagraph(kept) {
		return text
	}
	return kept
}

// looksLikeReasoning reports whether a reply is dominated by the model's
// deliberation (>=reasoningMinParagraphs reasoning paragraphs) — i.e. it never
// produced a real client-facing answer. Such a reply must not be posted; the
// caller nudges then defers instead.
func looksLikeReasoning(text string) bool {
	count := 0
	for _, p := range splitParagraphs(text) {
		if isReasoningParagraph(p) {
			count++
		}
	}
	return count >= reasoningMinParagraphs
}

// splitParagraphs normalises CRLF and splits on blank lines.
func splitParagraphs(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	raw := strings.Split(s, "\n\n")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// stripCodeFences removes ``` fence delimiter lines (the model may wrap its
// answer in a markdown code block). The block's content is kept — the fences
// are decoration; throwing the whole answer away would leave nothing to post.
func stripCodeFences(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	out := strings.ReplaceAll(b.String(), "`", "")
	return strings.TrimSpace(out)
}

func truncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	out := string([]rune(s)[:max])
	if lastBreak := strings.LastIndexAny(out, " \n.;،؟"); lastBreak > max/2 {
		return out[:lastBreak]
	}
	return out
}
