// Prompt building and reply shaping for the chat AI. Pure string work, fully
// unit-testable. The prompt is assembled from real chat rows only (the model
// never sees SQL), and the final reply always carries the AI disclosure line
// appended deterministically by us — never glued on by the model.
package chatbot

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"whatsapp-agent-tool/internal/carshttp"
)

// AiDisclosureLine is appended to every outgoing AI reply (AGENTS.md §0.9).
const AiDisclosureLine = "هذا الرد مولد بالذكاء الاصطناعي"

// DeferralReply is the safe holding message posted when the model cannot be
// trusted to produce a client-facing sentence (AGENTS.md §13, MISSING/AMBIGUOUS).
const DeferralReply = "سأطلب من المدير وأعود إليك."

// historyLimit is how many recent messages (any sender) the model sees as
// conversational context before the client's new lines.
const historyLimit = 12

// BuildPrompt composes the single text the opencode agent is asked to reply on.
// history is the tail of the group (oldest first), incoming are the new
// messages (oldest first) from anyone except the AI account itself. perms tells
// the model what this group is allowed to do, so it can phrase a short refusal
// for a request it must hand to the team.
func BuildPrompt(groupName, aiUsername string, perms Perms, history, incoming []carshttp.GroupMessage) string {
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
	b.WriteString(`
Reply in the language the newest message is written in — usually Algerian Derja/Arabic.
Rules:
- One text reply of 2-4 short, warm sentences. Address the person who wrote.
- If the message asks about a price, stock, availability or delivery, look it up with
  ` + "`whatsapp-tool`" + ` FIRST (see the tool instructions above), then answer with ONLY
  the figure the tool returned. Never invent, round or estimate one.
- A quote result with status OK always carries a price — give that number. If its
  warnings, stale or per-line expired flags say the rate or freight is old or expired,
  STILL give the number, but add a short note that it may have changed; you may offer
  to confirm with the team. The price is the cheapest applicable shipping line; the
  lines array lists a price per shipping line if you need to compare.
- Only when the tool returns MISSING or AMBIGUOUS (no figure at all, or several cars
  matched) or you cannot identify the exact car, reply briefly in the writer's language
  that you will pass the question to the team (e.g. "سأطلب من المدير وأعود إليك").
- If the message asks about SHIPPING / FREIGHT rates to a port WITHOUT naming a car
  ("كم سعر الشحن إلى الجزائر", "how much is shipping to Oran"), run
  ` + "`whatsapp-tool freight --port <PORT>`" + ` (omit --port to list every Algerian port,
  e.g. for "Algeria" itself) and answer with the cheapest line's containerChargeUsd per
  40ft container and/or the per-car figures (perCarBigUsd for a big car, perCarSmallUsd
  for a small car). Do NOT defer — this command always returns a figure. If the chosen
  line is flagged expired, still give it but note it may have changed.
- Default destination port is ALGIERS: quote with --port ALGIERS and, when you give a
  price, say it is landed at the port of Algiers ("واصلة ميناء الجزائر").
- Never ask the writer for money or payment details; those stay with the team.
- Use Arabic numerals (0..9), never Eastern Arabic numerals.
- Reply with exactly one message, no markdown, and nothing else around it.
`)
	return b.String()
}

// formatMessage renders one row for the model.
func formatMessage(m carshttp.GroupMessage) string {
	sender := strings.TrimSpace(m.SenderName)
	role := strings.TrimSpace(m.SenderRole)
	if role == "" || sender == "" {
		sender = "unknown"
	}
	return fmt.Sprintf("%s <%s> %s:\n%s\n", m.Time, sender, role, m.Message)
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
// character cap, and appends the AI disclosure line. Returns "" for a blank
// reply (caller then skips posting).
func FinalizeReply(raw string, maxChars int) string {
	text := strings.TrimSpace(raw)
	text = stripCodeFences(text)
	if text == "" {
		return ""
	}
	if maxChars > 0 {
		text = truncateRunes(text, maxChars)
	}
	return text + "\n" + AiDisclosureLine
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
