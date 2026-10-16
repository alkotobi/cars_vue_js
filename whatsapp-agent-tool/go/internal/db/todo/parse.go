package todo

import (
	"strings"
	"time"
)

// parseWhen parses the date/datetime forms the CLI accepts. A date-only value
// ("2026-10-20") is treated as end-of-day UTC so "due today" still counts as due
// on that day. All times are interpreted as UTC (the DB stores UTC).
func parseWhen(v string) (time.Time, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t.Add(24*time.Hour - time.Millisecond), true
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05.000",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	} {
		if t, err := time.Parse(layout, v); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// nowSQL is UTC now in the DATETIME(3) literal shape.
func nowSQL() string { return time.Now().UTC().Format("2006-01-02 15:04:05.000") }
