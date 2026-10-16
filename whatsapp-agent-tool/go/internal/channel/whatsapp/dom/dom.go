// Package dom is the Go twin of src/whatsapp/dom.ts — pure DOM/text helpers for
// parsing WhatsApp Web. Everything here is side-effect free and operates on
// strings or raw scrape shapes, so it is unit-testable without a browser.
package dom

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// LoginURL is the WhatsApp Web entry point.
const LoginURL = "https://web.whatsapp.com/"

// LaunchArgs mirrors dom.ts LAUNCH_ARGS (used by the legacy launch path).
var LaunchArgs = []string{
	"--no-sandbox",
	"--disable-setuid-sandbox",
	"--disable-blink-features=AutomationControlled",
}

var (
	timeRe         = regexp.MustCompile(`^(\d{1,2}:\d{2}(\s?[APap][Mm])?|\d{1,2}[\/.]\d{1,2}[\/.,]?\d{0,4}|أمس|امس|اليوم|أبارحة|امبارحة|Yesterday|Today|\d{4}\/\d{1,2}\/\d{1,2}|الآن|Now)$`)
	unreadBannerRe = regexp.MustCompile(`(?i)(غير\s*مقروء|unread)`)
	voiceRe        = regexp.MustCompile(`(?i)(voice|audio|mic|صوت|رسالة\s*صوتية|مكالمة\s*صوتية|call.*voice|voice.*call|ptt)`)
	callRe         = regexp.MustCompile(`(?i)(call|missed|مكالمة|اتصال|فائت|مفقود|ended|ringing|outgoing|incoming)`)
	missedRe       = regexp.MustCompile(`(?i)(missed|فائت|مفقود)`)

	bidiRe = regexp.MustCompile("[\u200E\u200F\u202A-\u202E\u2066-\u2069]")
	nbspRe = regexp.MustCompile("\u00a0")
	wsRe   = regexp.MustCompile("[ \t]+")

	// normalizeTitle strips bidi marks/whitespace and lowercases.
	titleStripRe = regexp.MustCompile("[\u200E\u200F\u202A-\u202E\u2066-\u2069\\s]")

	arabicIndicDigits  = regexp.MustCompile("[٠-٩]")
	easternIndicDigits = regexp.MustCompile("[۰-۹]")
	digitRuns          = regexp.MustCompile(`\d+`)
	trailingNumRe      = regexp.MustCompile(`\s+\d+\s*$`)

	allChipRe    = regexp.MustCompile(`(?i)(الكل|^all$)`)
	unreadChipRe = regexp.MustCompile(`(?i)(غير مقروء|unread)`)
)

// Rand returns min..max inclusive (matches dom.ts rand()).
func Rand(min, max int) int {
	if max < min {
		max = min
	}
	return rand.Intn(max-min+1) + min
}

// Jitter returns ms +/- variance*ms.
func Jitter(ms int, variance float64) int {
	v := float64(ms) * variance
	return ms + Rand(-int(v), int(v))
}

// CleanText strips bidi marks, turns NBSP into space and collapses spaces/tabs.
func CleanText(s string) string {
	s = bidiRe.ReplaceAllString(s, "")
	s = nbspRe.ReplaceAllString(s, " ")
	return wsRe.ReplaceAllString(s, " ")
}

// Lines returns the non-empty trimmed lines of s.
func Lines(s string) []string {
	parts := strings.Split(CleanText(s), "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// TimeReTest reports whether s looks like a sidebar time/date stamp.
func TimeReTest(s string) bool { return timeRe.MatchString(s) }

// BuildSweepNote explains a sweep result that came back short (nil when fine).
func BuildSweepNote(found, requested int, reachedEnd bool, filter string, unreadTotal *int) *string {
	var notes []string
	if found < requested {
		if reachedEnd {
			notes = append(notes, "List exhausted after "+itoa(found)+" chats.")
		} else {
			notes = append(notes, "Timed out after "+itoa(found)+" chats.")
		}
	}
	if filter == "unread" && unreadTotal != nil && found < *unreadTotal {
		notes = append(notes,
			"WhatsApp counts "+itoa(*unreadTotal)+" unread chats but the walk collected "+itoa(found)+". "+
				`The gap is archived chats (data-testid="chatlist-panel-archived-button"), which this tool does not open.`)
	}
	if len(notes) == 0 {
		return nil
	}
	joined := strings.Join(notes, " ")
	return &joined
}

// ParseArabicDigits converts Arabic-Indic / Eastern Arabic-Indic digits to ASCII
// and returns the last integer found, or nil.
func ParseArabicDigits(input string) *int {
	if input == "" {
		return nil
	}
	normalized := arabicIndicDigits.ReplaceAllStringFunc(input, func(d string) string {
		return itoa(int([]rune(d)[0]) - 0x0660)
	})
	normalized = easternIndicDigits.ReplaceAllStringFunc(normalized, func(d string) string {
		return itoa(int([]rune(d)[0]) - 0x06f0)
	})
	runs := digitRuns.FindAllString(normalized, -1)
	if len(runs) == 0 {
		return nil
	}
	n, err := strconv.Atoi(runs[len(runs)-1])
	if err != nil {
		return nil
	}
	return &n
}

// RawMessage is one scraped bubble.
type RawMessage struct {
	Text         string
	Meta         string
	TailOut      bool
	TailIn       bool
	BubbleCenter float64
	Width        float64
}

// RawPanel is the scraped message panel.
type RawPanel struct {
	Dir      string
	HasPanel bool
	Center   float64
}

// IsOutgoing mirrors dom.ts isOutgoing: bubble geometry survives the RTL mirror.
func IsOutgoing(m RawMessage, panel RawPanel) bool {
	if panel.HasPanel && m.Width > 0 {
		isLeft := m.BubbleCenter < panel.Center
		if panel.Dir == "rtl" {
			return isLeft
		}
		return !isLeft
	}
	if panel.Dir == "rtl" {
		return m.TailIn
	}
	return m.TailOut
}

// ChatNameFrom picks the first title line that is not the unread banner.
func ChatNameFrom(rawTitle string) string {
	for _, l := range Lines(rawTitle) {
		if !unreadBannerRe.MatchString(l) {
			return l
		}
	}
	return ""
}

// NormalizeTitle strips bidi marks/whitespace and lowercases for comparison.
func NormalizeTitle(s string) string {
	return strings.ToLower(titleStripRe.ReplaceAllString(s, ""))
}

// MessageExtra carries optional bubble signals for type inference.
type MessageExtra struct {
	HasAudio *bool
	HasImage *bool
	HasVideo *bool
	IsSystem *bool
}

// InferMessageType mirrors dom.ts inferMessageType.
func InferMessageType(text string, extra *MessageExtra) string {
	t := strings.TrimSpace(text)
	if t == "" && extra == nil {
		return "unknown"
	}
	if extra != nil {
		if extra.IsSystem != nil && *extra.IsSystem {
			return "system"
		}
		if extra.HasAudio != nil && *extra.HasAudio {
			if voiceRe.MatchString(t) {
				return "voice"
			}
			return "audio"
		}
		if extra.HasImage != nil && *extra.HasImage {
			return "image"
		}
		if extra.HasVideo != nil && *extra.HasVideo {
			return "video"
		}
	}
	if callRe.MatchString(t) {
		if missedRe.MatchString(t) {
			return "missed_call"
		}
		return "call"
	}
	if voiceRe.MatchString(t) {
		return "voice"
	}
	if t != "" {
		return "text"
	}
	return "unknown"
}

// TrimPreviewNumber strips the trailing unread-count number from a preview.
func TrimPreviewNumber(preview string) string {
	return strings.TrimSpace(trailingNumRe.ReplaceAllString(preview, ""))
}

// MatchAll reports whether a filter chip label is the "All" chip.
func MatchAll(label string) bool { return allChipRe.MatchString(label) }

// MatchUnread reports whether a filter chip label is the "Unread" chip.
func MatchUnread(label string) bool { return unreadChipRe.MatchString(label) }

func itoa(n int) string { return strconv.Itoa(n) }
