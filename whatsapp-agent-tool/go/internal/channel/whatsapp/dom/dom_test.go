package dom

import "testing"

func TestCleanText(t *testing.T) {
	got := CleanText("hello\u200F world\u00a0 \u200Eenter\u200E  x\ty")
	if got != "hello world enter x y" {
		t.Fatalf("got %q", got)
	}
}

func TestLines(t *testing.T) {
	got := Lines("a\n \nb\u00a0\n\ne")
	want := []string{"a", "b", "e"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestParseArabicDigits(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"٠٢٤٥", 245},
		{"٣", 3},
		{"۱۲۳", 123},
		{"110", 110},
		{"الكل:", 0},
		{"غير مقروء ٥١", 51},
	}
	for _, c := range cases {
		got := ParseArabicDigits(c.in)
		if got == nil {
			if c.want != 0 {
				t.Fatalf("%q: nil", c.in)
			}
			continue
		}
		if *got != c.want {
			t.Fatalf("%q: got %d want %d", c.in, *got, c.want)
		}
	}
}

func TestIsOutgoing(t *testing.T) {
	// LTR: bubbles on the right are outgoing.
	p := RawPanel{Dir: "ltr", HasPanel: true, Center: 100}
	if !IsOutgoing(RawMessage{BubbleCenter: 150, Width: 50}, p) {
		t.Fatal("right bubble should be outgoing in ltr")
	}
	if IsOutgoing(RawMessage{BubbleCenter: 50, Width: 50}, p) {
		t.Fatal("left bubble should be incoming in ltr")
	}
	// RTL: bubbles on the left are outgoing.
	pr := RawPanel{Dir: "rtl", HasPanel: true, Center: 100}
	if !IsOutgoing(RawMessage{BubbleCenter: 50, Width: 50}, pr) {
		t.Fatal("left bubble should be outgoing in rtl")
	}
	// No panel: fall back to tail flags.
	if !IsOutgoing(RawMessage{TailOut: true}, RawPanel{Dir: "ltr"}) {
		t.Fatal("tail-out should mark outgoing")
	}
	if IsOutgoing(RawMessage{TailIn: true}, RawPanel{Dir: "ltr"}) {
		t.Fatal("tail-in alone should not be outgoing in ltr")
	}
}

func TestChatNameFrom(t *testing.T) {
	got := ChatNameFrom("Ali\nغير مقروء")
	if got != "Ali" {
		t.Fatalf("got %q", got)
	}
}

func TestInferMessageType(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Voice message (0:17)", "voice"},
		{"🎤 Voice message", "voice"},
		{"Missed call", "missed_call"},
		{"Call ended", "call"},
		{"كلمني", "text"},
		{"", "unknown"},
	}
	for _, c := range cases {
		if got := InferMessageType(c.in, nil); got != c.want {
			t.Fatalf("%q: got %s want %s", c.in, got, c.want)
		}
	}

	// DOM flags (browser scrape): a media/system bubble has no text at all.
	b := func(v bool) *bool { return &v }
	extraCases := []struct {
		in    string
		extra MessageExtra
		want  string
	}{
		{"", MessageExtra{HasAudio: b(true)}, "audio"},
		{"0:17", MessageExtra{HasAudio: b(true)}, "audio"},
		{"Missed call", MessageExtra{IsSystem: b(true)}, "missed_call"}, // call rows win over system
		{"Ali changed the subject", MessageExtra{IsSystem: b(true)}, "system"},
		{"", MessageExtra{HasImage: b(true)}, "image"},
		{"", MessageExtra{HasVideo: b(true)}, "video"},
	}
	for _, c := range extraCases {
		if got := InferMessageType(c.in, &c.extra); got != c.want {
			t.Fatalf("%q +extra: got %s want %s", c.in, got, c.want)
		}
	}
}

func TestTrimPreviewNumber(t *testing.T) {
	if got := TrimPreviewNumber("hi there 2"); got != "hi there" {
		t.Fatalf("got %q", got)
	}
	if got := TrimPreviewNumber("plain"); got != "plain" {
		t.Fatalf("got %q", got)
	}
}

func TestChipMatchers(t *testing.T) {
	if !MatchAll("الكل") || !MatchAll("All") {
		t.Fatal("all chip labels should match")
	}
	if !MatchUnread("غير مقروء") || !MatchUnread("غير مقروء 51") || !MatchUnread("Unread") {
		t.Fatal("unread chip labels should match")
	}
}

func TestBuildSweepNote(t *testing.T) {
	if n := BuildSweepNote(30, 50, true, "all", nil); n == nil || *n != "List exhausted after 30 chats." {
		t.Fatalf("got %v", n)
	}
	if BuildSweepNote(50, 50, false, "all", nil) != nil {
		t.Fatal("full sweep should have no note")
	}
	ut := 80
	n := BuildSweepNote(30, 50, true, "unread", &ut)
	if n == nil {
		t.Fatal("expected note")
	}
}

func TestTimeRe(t *testing.T) {
	for _, s := range []string{"10:30", "09:45 PM", "اليوم", "Yesterday", "14/03", "2026/10/11", "الآن"} {
		if !TimeReTest(s) {
			t.Fatalf("%q should look like a time", s)
		}
	}
	if TimeReTest("buy the car now") {
		t.Fatal("sentence should not look like a time")
	}
}
