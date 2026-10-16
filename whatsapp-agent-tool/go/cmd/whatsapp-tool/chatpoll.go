// chat-poll — run the cars-app chat AI.
//
//	whatsapp-tool chat-poll [--once] [--interval N]
//
// Polls the CARS_TENANT database for client messages in the groups the AI user
// belongs to, generates a reply per group with the opencode CLI, and posts it
// back as the AI user. Default: run forever, sleeping CHAT_POLL_INTERVAL
// seconds between cycles. --once runs a single pass and exits (for smoke
// tests / cron). Each cycle prints one compact JSON report line on stdout;
// diagnostics go to stderr.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"whatsapp-agent-tool/internal/chatbot"
	"whatsapp-agent-tool/internal/config"
)

func cmdChatPoll(args []string) {
	once := has(args, "once")
	interval := config.C.Chatbot.PollInterval
	if s := flag(args, "interval"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			interval = time.Duration(n) * time.Second
		}
	}
	alarmer := chatbot.NewAlarmer(config.C.Chatbot.AlertThrottle, config.C.Chatbot.AlarmCommand, config.C.Chatbot.AlarmVoice)

	for {
		rep, err := chatbot.PollOnce()

		// Speak (throttled) whenever generation failed for any reason; --once
		// runs stay silent so smoke tests / cron don't alarm.
		if !once {
			detail := rep.LastError
			if err != nil {
				detail = err.Error()
			}
			alarmer.Update(time.Now(), err != nil || rep.Errors > 0, detail)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "chat-poll: %v\n", err)
			printCompact(&ordered{Pairs: []orderedPair{pair("status", "ERROR"), pair("error", err.Error())}})
			if once {
				os.Exit(1)
			}
			time.Sleep(interval)
			continue
		}
		printCompact(rep)
		if once {
			return
		}
		time.Sleep(interval)
	}
}
