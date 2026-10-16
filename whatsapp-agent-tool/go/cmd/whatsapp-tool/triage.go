package main

// triage — Go twin of src/triage.ts. Batch unread triage in ONE process.
//
//	whatsapp-tool triage              # 10 newest unread chats, 8 messages each, mark read
//	whatsapp-tool triage 20 5         # 20 chats, 5 messages each
//	whatsapp-tool triage 20 5 --no-read   # don't mutate unread state
//
// Full payload (chat names + messages) -> data/triage.json (read it with ONE Read).
// stdout: compact summary only, never message bodies.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/channel/whatsapp/dom"
	"whatsapp-agent-tool/internal/channel/whatsapp/wa"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/statex"
	"whatsapp-agent-tool/internal/timex"
)

func cmdTriage(cli wa.Client, args []string) {
	markRead := !has(args, "no-read")

	// Node: process.argv.slice(2).filter(a => a !== "--no-read"); args[0]=limit, args[1]=perChat.
	var positionals []string
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			continue
		}
		positionals = append(positionals, a)
	}
	limit := clampInt(positionals, 0, 10, 1, 50)
	perChat := clampInt(positionals, 1, 8, 1, 30)

	out := statex.RunFile("triage.json")

	if err := cli.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "[triage] init failed:", err)
		os.Exit(1)
	}
	sweep, err := cli.GetLastUnreadChats(limit, 90000)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[triage] unread sweep failed:", err)
		_ = cli.Close()
		os.Exit(1)
	}
	chats := sweep.Chats
	if len(chats) > limit {
		chats = chats[:limit]
	}

	results := make([]*ordered, 0, len(chats))
	openedCount, markedCount := 0, 0
	failed := []string{}
	for _, c := range chats {
		opened, _ := cli.OpenChat(c.Name)
		messages := []wa.Message{}
		if opened {
			if msgs, err := cli.GetLastMessages(perChat); err == nil && msgs != nil {
				messages = msgs
			}
		}
		marked := false
		if opened && markRead {
			marked, _ = cli.MarkAsRead()
		}
		// Node object order: name, [unreadCount], opened, markedRead, messages.
		pairs := []orderedPair{pair("name", c.Name)}
		if c.UnreadCount != "" {
			pairs = append(pairs, pair("unreadCount", c.UnreadCount))
		}
		pairs = append(pairs,
			pair("opened", opened),
			pair("markedRead", marked),
			pair("messages", messages),
		)
		results = append(results, &ordered{Pairs: pairs})
		if opened {
			openedCount++
		} else {
			failed = append(failed, c.Name)
		}
		if marked {
			markedCount++
		}
		status := "FAILED"
		if opened {
			status = "ok"
		}
		fmt.Fprintf(os.Stderr, "[triage] %s: %s (%d msgs)\n", status, c.Name, len(messages))
		// Read-only pacing: short human gap, nothing like the 4-9 s send spacing.
		time.Sleep(time.Duration(dom.Rand(500, 1400)) * time.Millisecond)
	}
	_ = cli.Close()

	// data/triage.json — pretty + trailing newline; unreadTotal omitted when undefined.
	filePairs := []orderedPair{pair("fetchedAt", timex.NowIso())}
	if sweep.UnreadTotal != nil {
		filePairs = append(filePairs, pair("unreadTotal", *sweep.UnreadTotal))
	}
	filePairs = append(filePairs,
		pair("collected", sweep.Collected),
		pair("markRead", markRead),
		pair("chats", results),
	)
	if dump, err := jsonx.MarshalIndent(&ordered{Pairs: filePairs}, "", "  "); err == nil {
		_ = os.WriteFile(out, append(dump, '\n'), 0o644)
	}

	// stdout summary — Node order: file, unreadTotal (?? null), collected, opened,
	// failed, markedRead, [note].
	markedOut := 0
	if markRead {
		markedOut = markedCount
	}
	summary := []orderedPair{
		pair("file", out),
		pair("unreadTotal", nullableInt(sweep.UnreadTotal)),
		pair("collected", sweep.Collected),
		pair("opened", openedCount),
		pair("failed", failed),
		pair("markedRead", markedOut),
	}
	if sweep.Note != nil {
		summary = append(summary, pair("note", *sweep.Note))
	}
	printCompact(&ordered{Pairs: summary})
	os.Exit(0)
}

// nullableInt maps a *int to a JSON value: the int, or null when undefined.
func nullableInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

// clampInt reads positionals[i] as an int, falling back to def and clamping to
// [lo, hi]. Node: Math.max(lo, Math.min(hi, Number(x ?? def))).
func clampInt(positionals []string, i, def, lo, hi int) int {
	v := def
	if i < len(positionals) && positionals[i] != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(positionals[i])); err == nil {
			v = n
		}
	}
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}
	return v
}
