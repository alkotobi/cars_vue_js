// Failure alerting for the chat daemon. When reply generation fails the daemon
// beeps on this Mac (a system alert via `osascript`), throttled so a long outage
// does not repeat every poll cycle, plus one shorter beep when it recovers. Set
// CHAT_ALARM_CMD=say to get a spoken voice alarm instead, or any other program to
// run it with the alert text as its argument. The decision logic is separated
// from the alerting so it can be unit-tested without touching the machine.
package chatbot

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// recoveryText is the utterance that marks a recovery (beeped once vs twice).
const recoveryText = "The chatbot has recovered."

// Alarmer decides when to speak an alarm. Zero value is unusable — use
// NewAlarmer.
type Alarmer struct {
	throttle time.Duration
	emit     func(string) // speaks (injectable for tests)
	last     time.Time    // when we last spoke (alarm or recovery)
	failing  bool         // whether the previous cycle failed
}

// NewAlarmer builds an Alarmer that alerts through the configured command
// (default "beep" — a macOS system alert). throttle <= 0 falls back to 10
// minutes; an empty cmd falls back to "beep".
func NewAlarmer(throttle time.Duration, cmd, voice string) *Alarmer {
	if throttle <= 0 {
		throttle = 10 * time.Minute
	}
	if strings.TrimSpace(cmd) == "" {
		cmd = "beep"
	}
	return &Alarmer{throttle: throttle, emit: speaker(cmd, voice)}
}

// Update reports one cycle's outcome. failed is true when the cycle had any
// reply-generation failure; detail is a short human-readable reason (may be
// empty). It speaks at most once per throttle window while failures continue,
// and once when failures stop. Returns true if it spoke.
func (a *Alarmer) Update(now time.Time, failed bool, detail string) bool {
	if !failed {
		if !a.failing {
			return false
		}
		a.failing = false
		a.last = now // also throttle the next alarm if it flaps back
		a.emit("The chatbot has recovered.")
		return true
	}

	first := !a.failing
	a.failing = true
	if first || a.last.IsZero() || now.Sub(a.last) >= a.throttle {
		a.last = now
		a.emit(alarmText(detail))
		return true
	}
	return false
}

// alarmText builds a short, speech-friendly alarm phrase.
func alarmText(detail string) string {
	msg := "Alert. The chatbot is failing."
	if d := strings.TrimSpace(detail); d != "" {
		msg += " " + truncateRunes(d, 160)
	}
	return msg
}

// speaker builds the alerting function for the configured command. "beep" (the
// default) plays system alert beeps — two for a failure, one for a recovery —
// and ignores the text; any other command is run with the alert text as its
// argument (so CHAT_ALARM_CMD=say restores the old spoken voice). Alerts are
// fire-and-forget (`Start`), and their errors are ignored on purpose.
func speaker(cmd, voice string) func(string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		cmd = "beep"
	}
	if cmd == "beep" {
		return func(text string) {
			n := 2
			if text == recoveryText {
				n = 1
			}
			beep(n)
		}
	}
	return func(text string) {
		args := make([]string, 0, 3)
		if voice != "" && cmd == "say" {
			args = append(args, "-v", voice)
		}
		args = append(args, text)
		_ = exec.Command(cmd, args...).Start()
	}
}

// beep plays n macOS system alert beeps via AppleScript. Best-effort — a missing
// osascript (or a headless session) just goes silently.
func beep(n int) {
	if n < 1 {
		n = 1
	}
	_ = exec.Command("osascript", "-e", fmt.Sprintf("beep %d", n)).Start()
}
