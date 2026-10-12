// Package timex — ISO-8601 helpers mirroring the JS `Date` idioms used across
// the codebase (new Date().toISOString(), Date.parse()).
package timex

import "time"

// NowIso returns UTC now as `2026-10-10T20:22:28.853Z` (Node's toISOString shape).
func NowIso() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

// ParseIso parses an ISO timestamp from the state files (ms precision or plain Z).
// Returns zero time + false for anything unparseable.
func ParseIso(iso string) (time.Time, bool) {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, iso); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// AgeMs returns how long ago iso was, or max int if unparseable
// (Node: Date.parse -> NaN -> Infinity).
func AgeMs(iso string) int64 {
	t, ok := ParseIso(iso)
	if !ok {
		return int64(^uint64(0) >> 1) // MaxInt64
	}
	return time.Since(t).Milliseconds()
}
