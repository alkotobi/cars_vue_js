package chatbot

import (
	"testing"
	"time"
)

// recorder captures what the alarmer would have spoken.
type recorder struct{ said []string }

func (r *recorder) emit(s string) { r.said = append(r.said, s) }

func newTestAlarmer(throttle time.Duration) (*Alarmer, *recorder) {
	rec := &recorder{}
	a := &Alarmer{throttle: throttle, emit: rec.emit}
	return a, rec
}

func TestAlarmerFirstFailureSpeaksOnce(t *testing.T) {
	a, rec := newTestAlarmer(10 * time.Minute)
	now := time.Unix(0, 0)

	if !a.Update(now, true, "boom") {
		t.Fatal("first failure should speak")
	}
	// Still failing, within throttle → silent.
	if a.Update(now.Add(time.Minute), true, "boom") {
		t.Fatal("second failure within throttle should be silent")
	}
	if len(rec.said) != 1 {
		t.Fatalf("want 1 utterance, got %d: %v", len(rec.said), rec.said)
	}
	if got := rec.said[0]; got != "Alert. The chatbot is failing. boom" {
		t.Fatalf("unexpected alarm text: %q", got)
	}
}

func TestAlarmerSpeaksAgainAfterThrottle(t *testing.T) {
	a, rec := newTestAlarmer(10 * time.Minute)
	now := time.Unix(0, 0)
	a.Update(now, true, "boom")
	if a.Update(now.Add(11*time.Minute), true, "boom") != true {
		t.Fatal("failure past throttle should speak again")
	}
	if len(rec.said) != 2 {
		t.Fatalf("want 2 utterances, got %d", len(rec.said))
	}
}

func TestAlarmerRecoverySpeaksOnce(t *testing.T) {
	a, rec := newTestAlarmer(time.Minute)
	now := time.Unix(0, 0)
	a.Update(now, true, "boom")
	if !a.Update(now.Add(time.Second), false, "") {
		t.Fatal("recovery should speak")
	}
	// Healthy cycles after recovery are silent.
	if a.Update(now.Add(2*time.Second), false, "") {
		t.Fatal("healthy cycle should be silent")
	}
	if len(rec.said) != 2 || rec.said[1] != "The chatbot has recovered." {
		t.Fatalf("unexpected utterances: %v", rec.said)
	}
}

func TestAlarmerHealthyFromStartSilent(t *testing.T) {
	a, rec := newTestAlarmer(time.Minute)
	if a.Update(time.Unix(0, 0), false, "") {
		t.Fatal("healthy first cycle should be silent")
	}
	if len(rec.said) != 0 {
		t.Fatalf("want silence, got %v", rec.said)
	}
}
