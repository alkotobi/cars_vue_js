package main

// pending — Go twin of src/pending.ts. Run FIRST every turn (§14b).
//
//	whatsapp-tool pending            # trim + promote + scan the "ai questions" group + report
//	whatsapp-tool pending --state    # trim + promote only (no browser, instant)
//
// The report is compact JSON on stdout with no client message bodies; the same
// payload (plus the clients' last messages) is written to data/pending.json.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/statex"
	"whatsapp-agent-tool/internal/timex"
	"whatsapp-agent-tool/internal/wa"
)

const pendingGroup = "ai questions"

// New, profile-qualified tag: [ID: sofia#5] (the profile prefix keeps ids
// collision-free across accounts). Legacy bare tags are never auto-matched.
var (
	idRE      = regexp.MustCompile(`\[ID:\s*([^\]\s#]+)#(\d+)\]`)
	legacyRE  = regexp.MustCompile(`\[ID:\s*\d+\]`)
	legacyAny = regexp.MustCompile(`\[ID:\s*\d+\]`)
)

func cmdPending(cli wa.Client, args []string) {
	stateOnly := has(args, "state")

	qf := statex.LoadQuestions()
	cf := statex.LoadChats()
	aq, ac := statex.TrimAndArchive(&qf, &cf)
	promoted := statex.PromoteReady(&cf, &qf)
	_ = statex.SaveQuestions(qf)
	_ = statex.SaveChats(cf)

	report := pendingReport{}
	report.scanned = false
	report.archivedQ = aq
	report.archivedC = ac
	report.answersFound = []int{}
	report.promotedNow = chatKeys(promoted)
	report.stateFile = stateFileLabel(qf, cf)

	if !stateOnly {
		runPendingGroupScan(cli, &qf, &cf, &report, &promoted)
		_ = statex.SaveQuestions(qf)
		_ = statex.SaveChats(cf)
	}

	report.dataReady = buildDataReady(cf, qf)
	report.waiting = buildWaiting(cf)
	report.toSendToGroup = buildToSend(qf)
	report.openQuestions = countOpen(qf)

	writePendingDump(&report, qf)
	out := report.marshal()
	printCompact(out)
	os.Exit(0)
}

type pendingReport struct {
	scanned       bool
	archivedQ     int
	archivedC     int
	answersFound  []int
	promotedNow   []string
	stateFile     string
	note          string
	noteSet       bool
	dataReady     []*ordered
	waiting       []*ordered
	toSendToGroup []*ordered
	openQuestions int
	fullDump      string
}

func (r *pendingReport) marshal() *ordered {
	pairs := []orderedPair{
		pair("scanned", r.scanned),
		pair("archived", &ordered{Pairs: []orderedPair{
			pair("archivedQuestions", r.archivedQ),
			pair("archivedChats", r.archivedC),
		}}),
		pair("answersFound", r.answersFound),
		pair("promotedNow", r.promotedNow),
		pair("stateFile", r.stateFile),
	}
	if r.noteSet {
		pairs = append(pairs, pair("note", r.note))
	}
	pairs = append(pairs,
		pair("dataReady", r.dataReady),
		pair("waiting", r.waiting),
		pair("toSendToGroup", r.toSendToGroup),
		pair("openQuestions", r.openQuestions),
		pair("fullDump", r.fullDump),
	)
	return &ordered{Pairs: pairs}
}

// runPendingGroupScan mirrors pending.ts' browser branch. On the Stage-5 stub
// every driver call fails, so the report carries Node's "Browser scan failed"
// note and the state-only promote stands.
func runPendingGroupScan(cli wa.Client, qf *statex.QuestionsFile, cf *statex.ChatsFile, r *pendingReport, promoted *[]statex.PendingChat) {
	if err := cli.Init(); err != nil {
		r.note = "Browser scan failed: " + err.Error() + ". State-only promote was done."
		r.noteSet = true
		return
	}
	defer cli.Close()
	if err := cli.RestoreAllFilter(); err != nil {
		r.note = "Browser scan failed: " + err.Error() + ". State-only promote was done."
		r.noteSet = true
		return
	}
	opened, err := cli.OpenChat(pendingGroup)
	if err != nil {
		r.note = "Browser scan failed: " + err.Error() + ". State-only promote was done."
		r.noteSet = true
		return
	}
	if !opened {
		r.note = `Could not open the "` + pendingGroup + `" group — state-only promote done, answers not scanned.`
		r.noteSet = true
		return
	}
	msgs, err := cli.GetLastMessages(50)
	if err != nil {
		r.note = "Browser scan failed: " + err.Error() + ". State-only promote was done."
		r.noteSet = true
		return
	}

	answers := map[int]string{}
	now := timex.NowIso()
	sawLegacy := false
	wanted := strings.ToLower(config.WAProfile())
	for _, m := range msgs {
		if m.From != "them" {
			continue // our own re-posts are not answers
		}
		if legacyAny.MatchString(m.Text) {
			sawLegacy = true
		}
		for _, match := range idRE.FindAllStringSubmatch(m.Text, -1) {
			if strings.ToLower(match[1]) != wanted {
				continue // answer for a different account's question
			}
			id := atoi(match[2])
			var q *statex.AiQuestion
			for i := range qf.Questions {
				if qf.Questions[i].Id == id {
					q = &qf.Questions[i]
					break
				}
			}
			if q == nil || q.Answered {
				continue
			}
			body := strings.TrimSpace(legacyRE.ReplaceAllString(idRE.ReplaceAllString(m.Text, ""), ""))
			if body == "" {
				continue
			}
			answers[id] = body
		}
	}
	if sawLegacy {
		r.note = `Ignored legacy bare "[ID: n]" tags (pre-profile, not unique across accounts). ` +
			`Re-post under "[ID: ` + config.WAProfile() + `#n]" if still open.`
		r.noteSet = true
	}

	touchedScan := false
	for i := range qf.Questions {
		q := &qf.Questions[i]
		if q.Answered || q.SentToGroup == nil || !*q.SentToGroup {
			continue
		}
		if a, ok := answers[q.Id]; ok {
			q.Answered = true
			q.AnsweredAt = &now
			by := "group"
			q.AnsweredBy = &by
			q.Answer = &a
			r.answersFound = append(r.answersFound, q.Id)
		}
		q.GroupLastScanAt = &now
		touchedScan = true
	}

	if len(answers) > 0 || touchedScan {
		*promoted = append(*promoted, statex.PromoteReady(cf, qf)...)
		r.promotedNow = dedup(chatKeys(*promoted))
	}
	r.scanned = true
}

func buildDataReady(cf statex.ChatsFile, qf statex.QuestionsFile) []*ordered {
	out := []*ordered{}
	for _, c := range cf.Chats {
		if c.Status != "data_ready" {
			continue
		}
		qs := []*ordered{}
		for _, id := range c.QuestionIds {
			var qq *statex.AiQuestion
			for i := range qf.Questions {
				if qf.Questions[i].Id == id {
					qq = &qf.Questions[i]
					break
				}
			}
			if qq == nil {
				continue
			}
			qs = append(qs, &ordered{Pairs: []orderedPair{
				pair("id", qq.Id),
				pair("question", qq.Question),
				pair("answer", nilOrStr(qq.Answer)),
				pair("answeredBy", nilOrStr(qq.AnsweredBy)),
				pair("askedAt", qq.AskedAt),
				pair("answeredAt", nilOrStr(qq.AnsweredAt)),
			}})
		}
		out = append(out, &ordered{Pairs: []orderedPair{
			pair("chatKey", c.ChatKey),
			pair("chatName", nilOrStr(c.ChatName)),
			pair("needed", strSlice(c.Needed)),
			pair("reason", nilOrStr(c.Reason)),
			pair("askedAt", nilOrStr(c.AskedAt)),
			pair("dataReceivedAt", nilOrStr(c.DataReceivedAt)),
			pair("questions", qs),
		}})
	}
	return out
}

func buildWaiting(cf statex.ChatsFile) []*ordered {
	out := []*ordered{}
	for _, c := range cf.Chats {
		if c.Status != "waiting_for_data" {
			continue
		}
		out = append(out, &ordered{Pairs: []orderedPair{
			pair("chatKey", c.ChatKey),
			pair("needed", strSlice(c.Needed)),
			pair("askedAt", nilOrStr(c.AskedAt)),
		}})
	}
	return out
}

func buildToSend(qf statex.QuestionsFile) []*ordered {
	out := []*ordered{}
	for _, q := range qf.Questions {
		sent := q.SentToGroup != nil && *q.SentToGroup
		if q.Answered || sent {
			continue
		}
		var chatKey any
		needed := []string{}
		if q.Context != nil {
			if q.Context.ChatKey != "" {
				chatKey = q.Context.ChatKey
			}
			needed = strSlice(q.Context.Needed)
		}
		out = append(out, &ordered{Pairs: []orderedPair{
			pair("id", q.Id),
			pair("profile", config.WAProfile()),
			pair("groupKey", statex.GroupKey(q.Id)),
			pair("question", q.Question),
			pair("chatKey", chatKey),
			pair("needed", needed),
		}})
	}
	return out
}

func writePendingDump(r *pendingReport, qf statex.QuestionsFile) {
	open := []statex.AiQuestion{}
	for _, q := range qf.Questions {
		if !q.Answered {
			open = append(open, q)
		}
	}
	dump := &ordered{Pairs: []orderedPair{
		pair("generatedAt", timex.NowIso()),
		pair("dataReady", r.dataReady),
		pair("waiting", r.waiting),
		pair("toSendToGroup", r.toSendToGroup),
		pair("openQuestions", open),
	}}
	data, err := jsonx.MarshalIndent(dump, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(statex.RunFile("pending.json"), append(data, '\n'), 0o644)
	r.fullDump = statex.RunFile("pending.json") + " (only if the client's original message is needed)"
}

// ---------------------------------------------------------------- helpers

func chatKeys(chats []statex.PendingChat) []string {
	out := make([]string, 0, len(chats))
	for _, c := range chats {
		out = append(out, c.ChatKey)
	}
	return out
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func stateFileLabel(qf statex.QuestionsFile, cf statex.ChatsFile) string {
	return filepath.Base(statex.QuestionsPath()) + "=" + itoaInt(len(qf.Questions)) +
		", " + filepath.Base(statex.ChatsPath()) + "=" + itoaInt(len(cf.Chats))
}

func countOpen(qf statex.QuestionsFile) int {
	n := 0
	for _, q := range qf.Questions {
		if !q.Answered {
			n++
		}
	}
	return n
}

func strSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nilOrStr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func itoaInt(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
