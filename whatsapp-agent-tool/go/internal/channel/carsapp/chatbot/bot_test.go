package chatbot

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"whatsapp-agent-tool/internal/config"
)

// fakeTransport is an in-memory Transport: it records what the loop asked for
// and replays canned answers, so pollGroup/seedCursor run without the cars app.
type fakeTransport struct {
	identity    int64
	identityOK  bool
	identityErr error

	groups    []Group
	groupsErr error

	perms    Perms
	permsErr error

	read    []Message
	readErr error

	// afterID/excludeUserID are the arguments of the last Read, so a test can
	// assert the loop passes the cursor and its own identity as expected.
	afterID    int64
	excludeUID int64

	history    []Message
	historyErr error

	maxID    int64
	maxIDErr error

	latest    Message
	latestOK  bool
	latestErr error

	sent    []string
	sendErr error
}

func (f *fakeTransport) Identity() (int64, bool, error) {
	return f.identity, f.identityOK, f.identityErr
}

func (f *fakeTransport) Groups(userID int64) ([]Group, error) { return f.groups, f.groupsErr }

func (f *fakeTransport) Perms(groupID int64) (Perms, error) { return f.perms, f.permsErr }

func (f *fakeTransport) Read(groupID, afterID, excludeUserID int64) ([]Message, error) {
	f.afterID, f.excludeUID = afterID, excludeUserID
	return f.read, f.readErr
}

func (f *fakeTransport) History(groupID int64, limit int) ([]Message, error) {
	return f.history, f.historyErr
}

func (f *fakeTransport) MaxID(groupID int64) (int64, error) { return f.maxID, f.maxIDErr }

func (f *fakeTransport) Latest(groupID int64) (Message, bool, error) {
	return f.latest, f.latestOK, f.latestErr
}

func (f *fakeTransport) Send(groupID, userID int64, text string) error {
	f.sent = append(f.sent, text)
	return f.sendErr
}

// fakeStore is an in-memory CursorStore.
type fakeStore struct{ cursors map[int64]int64 }

func newFakeStore(cursors map[int64]int64) *fakeStore {
	if cursors == nil {
		cursors = map[int64]int64{}
	}
	return &fakeStore{cursors: cursors}
}

func (f *fakeStore) Cursor(groupID int64) int64 { return f.cursors[groupID] }

func (f *fakeStore) SetCursor(groupID, lastID int64) { f.cursors[groupID] = lastID }

func (f *fakeStore) Save() error { return nil }

// generateCall records one invocation of the brain step.
type generateCall struct {
	prompt string
	perms  Perms
}

// stubGenerate replaces the opencode call for the duration of a test and
// returns a pointer to the recorded calls. pollGroup must NEVER run unstubbed
// in a test — it would spawn the real CLI.
func stubGenerate(t *testing.T, reply string, err error) *[]generateCall {
	t.Helper()
	calls := &[]generateCall{}
	prev := generateFn
	generateFn = func(prompt string, perms Perms) (string, bool, error) {
		*calls = append(*calls, generateCall{prompt: prompt, perms: perms})
		return reply, false, err
	}
	t.Cleanup(func() { generateFn = prev })
	return calls
}

func testGroup() Group { return Group{ID: 42, Name: "Selection: LIVAN"} }

func TestPollGroupNoNewMessagesLeavesCursorAlone(t *testing.T) {
	tr := &fakeTransport{}
	store := newFakeStore(map[int64]int64{42: 10})
	calls := stubGenerate(t, "should not be used", nil)

	out, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false })
	if err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if out.newMessages != 0 || out.replied != 0 || out.skipped != 0 {
		t.Fatalf("outcome = %+v, want zero", out)
	}
	if len(*calls) != 0 {
		t.Fatalf("brain ran with no new messages: %d calls", len(*calls))
	}
	if store.Cursor(42) != 10 {
		t.Fatalf("cursor = %d, want unchanged 10", store.Cursor(42))
	}
}

func TestPollGroupRepliesAndAdvancesCursor(t *testing.T) {
	tr := &fakeTransport{
		read: []Message{
			{ID: 11, Text: "wesh", SenderName: "ahmed", SenderRole: "Client"},
			{ID: 12, Text: "prix?", SenderName: "ahmed", SenderRole: "Client"},
		},
		perms: Perms{Cost: true},
	}
	store := newFakeStore(map[int64]int64{42: 10})
	calls := stubGenerate(t, "مرحبا", nil)

	out, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false })
	if err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if out.newMessages != 2 || out.replied != 2 || out.skipped != 0 {
		t.Fatalf("outcome = %+v, want 2 new/2 replied", out)
	}
	if got := store.Cursor(42); got != 12 {
		t.Fatalf("cursor = %d, want 12 (the newest answered id)", got)
	}
	if tr.afterID != 10 || tr.excludeUID != 7 {
		t.Fatalf("Read(afterID=%d, exclude=%d), want (10, 7)", tr.afterID, tr.excludeUID)
	}
	if len(tr.sent) != 1 {
		t.Fatalf("sent %d replies, want exactly 1", len(tr.sent))
	}
	if !strings.Contains(tr.sent[0], "مرحبا") || !strings.Contains(tr.sent[0], AiDisclosureLine) {
		t.Fatalf("posted reply %q missing body or AI disclosure line", tr.sent[0])
	}
	if len(*calls) != 1 {
		t.Fatalf("brain ran %d times, want 1", len(*calls))
	}
	c := (*calls)[0]
	if !strings.Contains(c.prompt, "Selection: LIVAN") || !strings.Contains(c.prompt, "prix?") {
		t.Fatalf("prompt missing group/messages:\n%s", c.prompt)
	}
	if c.perms != (Perms{Cost: true}) {
		t.Fatalf("perms threaded = %+v, want {Cost:true}", c.perms)
	}
}

func TestPollGroupKeepsCursorWhenReadFails(t *testing.T) {
	tr := &fakeTransport{readErr: errors.New("db down")}
	store := newFakeStore(map[int64]int64{42: 10})
	calls := stubGenerate(t, "x", nil)

	if _, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false }); err == nil {
		t.Fatal("want read error")
	}
	if len(*calls) != 0 {
		t.Fatalf("brain ran after a read failure: %d", len(*calls))
	}
	if store.Cursor(42) != 10 {
		t.Fatalf("cursor = %d, want unchanged 10 (so the messages are retried)", store.Cursor(42))
	}
}

func TestPollGroupKeepsCursorWhenGenerateFails(t *testing.T) {
	tr := &fakeTransport{read: []Message{{ID: 11, Text: "hi"}}}
	store := newFakeStore(map[int64]int64{42: 10})
	stubGenerate(t, "", errors.New("opencode failed"))

	if _, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false }); err == nil {
		t.Fatal("want generate error")
	}
	if len(tr.sent) != 0 {
		t.Fatalf("posted %d replies after a generation failure", len(tr.sent))
	}
	if store.Cursor(42) != 10 {
		t.Fatalf("cursor = %d, want unchanged 10", store.Cursor(42))
	}
}

func TestPollGroupDropsReplyWhenLeaseLost(t *testing.T) {
	tr := &fakeTransport{read: []Message{{ID: 11, Text: "hi"}}}
	store := newFakeStore(map[int64]int64{42: 10})
	stubGenerate(t, "late reply", nil)

	_, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return true })
	if err == nil || !strings.Contains(err.Error(), "lease lost") {
		t.Fatalf("err = %v, want a lease-lost error", err)
	}
	if len(tr.sent) != 0 {
		t.Fatalf("posted %d replies after losing the lease (would double-reply)", len(tr.sent))
	}
	if store.Cursor(42) != 10 {
		t.Fatalf("cursor = %d, want unchanged 10 so the new owner answers them", store.Cursor(42))
	}
}

func TestPollGroupSendErrorStillAdvancesCursor(t *testing.T) {
	// Deliberate: the cursor moves before posting so a failed post is accepted
	// as a miss rather than risking a duplicate answer every cycle.
	tr := &fakeTransport{read: []Message{{ID: 11, Text: "hi"}}, sendErr: errors.New("write failed")}
	store := newFakeStore(map[int64]int64{42: 10})
	stubGenerate(t, "reply", nil)

	if _, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false }); err == nil {
		t.Fatal("want send error")
	}
	if store.Cursor(42) != 11 {
		t.Fatalf("cursor = %d, want 11 (advance-before-send)", store.Cursor(42))
	}
}

func TestPollGroupEmptyReplyAdvancesCursorWithoutPosting(t *testing.T) {
	tr := &fakeTransport{read: []Message{{ID: 11, Text: "hi"}, {ID: 12, Text: "yo"}}}
	store := newFakeStore(map[int64]int64{42: 10})
	stubGenerate(t, "   \n ", nil)

	out, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false })
	if err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if out.skipped != 2 || out.replied != 0 {
		t.Fatalf("outcome = %+v, want 2 skipped / 0 replied", out)
	}
	if len(tr.sent) != 0 {
		t.Fatalf("posted an empty reply: %v", tr.sent)
	}
	if store.Cursor(42) != 12 {
		t.Fatalf("cursor = %d, want 12", store.Cursor(42))
	}
}

func TestPollGroupHistoryFailureIsTolerated(t *testing.T) {
	tr := &fakeTransport{
		read:       []Message{{ID: 11, Text: "hi"}},
		historyErr: errors.New("history unavailable"),
	}
	store := newFakeStore(map[int64]int64{42: 10})
	calls := stubGenerate(t, "ok", nil)

	out, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false })
	if err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if out.replied != 1 {
		t.Fatalf("outcome = %+v, want a reply despite missing history", out)
	}
	if len(*calls) != 1 {
		t.Fatalf("brain ran %d times, want 1", len(*calls))
	}
	if strings.Contains((*calls)[0].prompt, "Recent conversation") {
		t.Fatal("prompt claims to have history when the transport failed")
	}
}

func TestPollGroupIncludesHistoryWhenAvailable(t *testing.T) {
	tr := &fakeTransport{
		read:    []Message{{ID: 11, Text: "new one"}},
		history: []Message{{ID: 9, Text: "earlier", SenderName: "sofia", SenderRole: "Admin"}},
	}
	store := newFakeStore(map[int64]int64{42: 10})
	calls := stubGenerate(t, "ok", nil)

	if _, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false }); err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("brain ran %d times, want 1", len(*calls))
	}
	p := (*calls)[0].prompt
	if !strings.Contains(p, "Recent conversation") || !strings.Contains(p, "earlier") {
		t.Fatalf("prompt missing history:\n%s", p)
	}
}

// A group the daemon has never seen must be seeded, never answered on the spot.

func TestSeedCursorSeedsPastStaleHistory(t *testing.T) {
	prev := config.C.Chatbot.SeedWindow
	config.C.Chatbot.SeedWindow = 0
	t.Cleanup(func() { config.C.Chatbot.SeedWindow = prev })

	tr := &fakeTransport{
		maxID:    7,
		latestOK: true,
		latest:   Message{ID: 7, Time: "2020-01-01 00:00:00"},
	}
	store := newFakeStore(nil)
	calls := stubGenerate(t, "x", nil)

	out, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false })
	if err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("answered a group it had never seen: %d brain calls", len(*calls))
	}
	if out.newMessages != 0 {
		t.Fatalf("outcome = %+v, want no new messages while seeding", out)
	}
	if store.Cursor(42) != 7 {
		t.Fatalf("cursor = %d, want seeded past history (7)", store.Cursor(42))
	}
}

func TestSeedCursorAnswersAGroupThatJustStarted(t *testing.T) {
	prev := config.C.Chatbot.SeedWindow
	config.C.Chatbot.SeedWindow = time.Hour
	t.Cleanup(func() { config.C.Chatbot.SeedWindow = prev })

	tr := &fakeTransport{
		maxID:    7,
		latestOK: true,
		latest:   Message{ID: 7, Time: time.Now().UTC().Format(messageTimeLayout)},
	}
	store := newFakeStore(nil)
	stubGenerate(t, "x", nil)

	if _, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false }); err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	// Seeded to maxID-1, so the next cycle reads that fresh message and answers it.
	if store.Cursor(42) != 6 {
		t.Fatalf("cursor = %d, want 6 (so the fresh message is answered next cycle)", store.Cursor(42))
	}
}

func TestSeedCursorEmptyGroupStaysAtZero(t *testing.T) {
	tr := &fakeTransport{}
	store := newFakeStore(nil)
	calls := stubGenerate(t, "x", nil)

	if _, err := pollGroup(tr, store, "ai", 7, testGroup(), func() bool { return false }); err != nil {
		t.Fatalf("pollGroup: %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("brain ran on an empty group: %d", len(*calls))
	}
	if store.Cursor(42) != 0 {
		t.Fatalf("cursor = %d, want 0 (re-check until the first message)", store.Cursor(42))
	}
}

func TestParseMessages(t *testing.T) {
	rows := []json.RawMessage{
		json.RawMessage(`{"id":1,"id_chat_group":9,"message":"hi","time":"2026-10-14 10:00:00",` +
			`"sender_username":"ahmed","sender_role":"Client","sender_type":"client"}`),
	}
	msgs, err := parseMessages(rows)
	if err != nil {
		t.Fatalf("parseMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	m := msgs[0]
	if m.ID != 1 || m.GroupID != 9 || m.Text != "hi" || m.SenderName != "ahmed" || m.SenderRole != "Client" {
		t.Fatalf("decoded = %+v, want every field of the row", m)
	}
	if m.Time != "2026-10-14 10:00:00" || m.SenderType != "client" {
		t.Fatalf("decoded = %+v, want the original time/type", m)
	}

	// A malformed row must be an error, never a silently dropped message: Read's
	// caller then leaves the cursor put and retries next cycle.
	if _, err := parseMessages([]json.RawMessage{json.RawMessage(`{bad json`)}); err == nil {
		t.Fatal("want an error for a malformed row")
	}
}

func TestHistoryOfReturnsNilWhenTransportFails(t *testing.T) {
	tr := &fakeTransport{historyErr: errors.New("boom")}
	if got := historyOf(tr, 42); got != nil {
		t.Fatalf("historyOf = %v, want nil so the prompt falls back to no history", got)
	}
}
