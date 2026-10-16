package chatbot

import (
	"testing"
	"time"
)

// TestWithinSeedWindow pins the boundary of the "is this group freshly started
// or discovered with stale history?" decision that drives first-sight seeding.
func TestWithinSeedWindow(t *testing.T) {
	now := time.Date(2026, 10, 15, 6, 0, 0, 0, time.UTC)
	window := 10 * time.Minute

	cases := []struct {
		name string
		at   time.Time
		want bool
	}{
		{"just now", now, true},
		{"5 min old", now.Add(-5 * time.Minute), true},
		{"exactly at the window", now.Add(-window), true},
		{"1s past the window", now.Add(-window - time.Second), false},
		{"a day old", now.Add(-24 * time.Hour), false},
		{"future clock skew counts as fresh", now.Add(time.Minute), true},
	}
	for _, c := range cases {
		if got := withinSeedWindow(c.at, now, window); got != c.want {
			t.Errorf("%s: withinSeedWindow = %v, want %v", c.name, got, c.want)
		}
	}
}
