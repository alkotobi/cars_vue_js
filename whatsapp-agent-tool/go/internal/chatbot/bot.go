// Package chatbot drives the cars-app chat AI. Each poll cycle it lists the
// groups the AI user belongs to, reads every client message newer than the
// per-group cursor, generates one reply per group with the opencode CLI (a
// tool-less agent) and posts it back as the AI user. The model never touches
// SQL; prompts come from real rows and replies are capped and stamped with the
// AI disclosure line before posting.
package chatbot

import (
	"fmt"
	"log"
	"os"
	"time"

	"whatsapp-agent-tool/internal/carshttp"
	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db/chatbotstate"
	"whatsapp-agent-tool/internal/db/groupperms"
	"whatsapp-agent-tool/internal/keyringx"
)

// Report is the machine-readable per-cycle outcome, printed as one JSON line.
type Report struct {
	Status string `json:"status"`
	Tenant string `json:"tenant"`
	// Owner is this process's lease owner (host:pid:profile); with several
	// machines on one tenant it says who is active. Empty when unset.
	Owner       string `json:"owner,omitempty"`
	Groups      int    `json:"groups"`
	NewMessages int    `json:"newMessages"`
	Replied     int    `json:"replied"`
	Skipped     int    `json:"skipped"`
	Errors      int    `json:"errors"`
	// LastError is the first group's failure message this cycle, when any —
	// surfaced so the daemon can speak a useful alarm instead of a bare count.
	LastError string `json:"lastError,omitempty"`
}

// groupOutcome is one group's contribution to the cycle report.
type groupOutcome struct {
	newMessages int
	replied     int
	skipped     int
}

// PollOnce runs exactly one pass: lease -> poll -> think -> post. Several machines
// may run against one tenant, but only the lease holder does any work: a standby
// returns a STANDBY report without touching the cars app. The holder renews the
// lease on a heartbeat while it works, and never posts after losing it. It never
// blocks for more than one reply generation (see llmTimeout) and returns a report
// even when individual groups failed.
func PollOnce() (Report, error) {
	rep := Report{Tenant: config.C.Cars.Tenant}
	logger := log.New(os.Stderr, "chat-poll ", log.LstdFlags)

	// Single-writer lease. owner is stable per process; the heartbeat keeps the
	// lease alive through long cycles and fails over within ttl if this machine dies.
	owner := config.C.Chatbot.LeaseOwner
	if owner == "" {
		owner = chatbotstate.Owner()
	}
	ttl := config.C.Chatbot.LeaseTTL
	rep.Owner = owner
	held, err := chatbotstate.Acquire(rep.Tenant, owner, ttl)
	if err != nil {
		return rep, fmt.Errorf("chat: acquire lease: %w", err)
	}
	if !held {
		rep.Status = "STANDBY"
		return rep, nil
	}
	keeper := startLeaseKeeper(rep.Tenant, owner, ttl, logger)
	defer keeper.stop()

	userID, ok, err := carshttp.UserIDOrLogin()
	if err != nil {
		return rep, err
	}
	if !ok || userID == 0 {
		return rep, fmt.Errorf("chat: the AI user's id is not resolvable — re-run: whatsapp-tool stock creds --username <user>")
	}

	groups, err := carshttp.MemberGroups(int64(userID))
	if err != nil {
		return rep, err
	}
	rep.Groups = len(groups)
	store, err := LoadCursorStore(rep.Tenant)
	if err != nil {
		return rep, err
	}

	aiUsername := "AI assistant"
	if u, err := keyringx.GetUsername(); err == nil && u != "" {
		aiUsername = u
	}

	for _, g := range groups {
		// Stop before starting another group once the lease is gone: a standby has
		// taken over and would answer the same messages.
		if keeper.Lost() {
			logger.Printf("lease lost — stopping cycle early")
			break
		}
		outcome, err := pollGroup(store, aiUsername, int64(userID), g, keeper.Lost)
		// Tally a group's contribution even when it errored (the cursor may
		// already have advanced past its messages) so the report stays truthful.
		rep.NewMessages += outcome.newMessages
		rep.Skipped += outcome.skipped
		rep.Replied += outcome.replied
		if err != nil {
			rep.Errors++
			if rep.LastError == "" {
				rep.LastError = err.Error()
			}
			logger.Printf("group %q (%d): %v", g.Name, g.ID, err)
		}
	}

	// Always flush cursors: they are monotonic, so recording an answered group
	// cannot move the shared cursor backwards even if we lost the lease.
	if err := store.Save(); err != nil {
		return rep, err
	}
	rep.Status = "OK"
	return rep, nil
}

// pollGroup runs the read->think->post pipeline for one group. The cursor only
// advances once the model has produced a reply (or legitimately returned
// nothing): a generation failure, a permission rejection or a lookup error
// leaves the cursor put so the same client message is retried on the next
// cycle instead of being silently skipped forever. The reply is posted after
// the cursor moves, so a failed post is accepted as a miss rather than risking
// a duplicate. lost reports whether the lease was lost while generating — in
// which case the reply is dropped and the cursor left put, so the new owner
// answers the same messages exactly once.
func pollGroup(store CursorStore, aiUsername string, userID int64, g carshttp.ChatGroup, lost func() bool) (groupOutcome, error) {
	var outcome groupOutcome

	cursor := store.Cursor(g.ID)
	if cursor == 0 {
		return seedCursor(store, g)
	}

	incoming, err := carshttp.NewMessages(g.ID, cursor, userID)
	if err != nil {
		return outcome, err
	}
	if len(incoming) == 0 {
		return outcome, nil
	}
	msgs := make([]carshttp.GroupMessage, 0, len(incoming))
	newMax := cursor
	for _, raw := range incoming {
		m, err := carshttp.ParseGroupMessage(raw)
		if err != nil {
			return outcome, fmt.Errorf("parse incoming: %w", err)
		}
		if m.ID > newMax {
			newMax = m.ID
		}
		msgs = append(msgs, m)
	}
	outcome.newMessages = len(msgs)

	hist := historyOf(g.ID)

	perms, err := loadPerms(config.C.Cars.Tenant, g.ID)
	if err != nil {
		return outcome, err
	}

	reply, err := Generate(BuildPrompt(g.Name, aiUsername, perms, hist, msgs), perms)
	if err != nil {
		// Nothing was posted, so retrying cannot duplicate a reply: leave the
		// cursor where it is and try the same messages again next cycle.
		return outcome, err
	}

	// The lease may have been taken over while we generated (long local models):
	// posting now would double-reply with the new owner. Drop the reply and leave
	// the cursor put so the new owner answers these messages exactly once.
	if lost != nil && lost() {
		return outcome, fmt.Errorf("lease lost during generation — reply dropped")
	}

	// The model produced a reply: the messages are now considered handled, so
	// advance the cursor before posting to avoid re-answering them.
	store.SetCursor(g.ID, newMax)

	final := FinalizeReply(reply, config.C.Chatbot.MaxReplyChars)
	if final == "" {
		outcome.skipped = len(msgs)
		return outcome, nil
	}
	if _, err := carshttp.PostUserMessage(g.ID, userID, final); err != nil {
		return outcome, fmt.Errorf("post reply: %w", err)
	}
	outcome.replied = len(msgs)
	return outcome, nil
}

// seedCursor sets the cursor for a group the daemon has not seen before. A
// group discovered with only stale messages is seeded past all of them — the
// bot never answers a pre-existing conversation. But if its newest message is
// recent (within SeedWindow) the group is treated as freshly started: the
// cursor lands just before that message so the next cycle answers it, instead
// of silently skipping a message someone wrote minutes ago. An empty group
// stays at 0 and is re-checked until its first message arrives.
func seedCursor(store CursorStore, g carshttp.ChatGroup) (groupOutcome, error) {
	var outcome groupOutcome
	maxID, err := carshttp.GroupMaxMessageID(g.ID)
	if err != nil {
		return outcome, fmt.Errorf("seed cursor: %w", err)
	}
	if maxID == 0 {
		store.SetCursor(g.ID, 0)
		return outcome, nil
	}
	fresh, err := latestIsFresh(g.ID)
	if err != nil {
		// A failed freshness probe must not start answering old history: fall
		// back to the conservative seed-past-everything behaviour.
		log.Printf("seed cursor: group %q (%d): %v", g.Name, g.ID, err)
	}
	if fresh {
		store.SetCursor(g.ID, maxID-1)
		return outcome, nil
	}
	store.SetCursor(g.ID, maxID)
	return outcome, nil
}

// latestIsFresh reports whether a group's newest message arrived within
// SeedWindow. message time is stored by the app as UTC_TIMESTAMP(), so it is
// parsed as UTC and compared against the current UTC time.
func latestIsFresh(groupID int64) (bool, error) {
	m, ok, err := carshttp.LatestMessage(groupID)
	if err != nil || !ok {
		return false, err
	}
	at, err := time.ParseInLocation(messageTimeLayout, m.Time, time.UTC)
	if err != nil {
		return false, fmt.Errorf("parse message time %q: %w", m.Time, err)
	}
	return withinSeedWindow(at, time.Now().UTC(), config.C.Chatbot.SeedWindow), nil
}

// messageTimeLayout matches the app's chat_messages.time format (UTC_TIMESTAMP()).
const messageTimeLayout = "2006-01-02 15:04:05"

// withinSeedWindow reports whether a message written at `at` is recent enough
// (within window of now) to be worth answering when its group is discovered.
// A future timestamp (clock skew) counts as fresh.
func withinSeedWindow(at, now time.Time, window time.Duration) bool {
	return now.Sub(at) <= window
}

// loadPerms reads the group's capabilities from the ai DB. A missing row is
// fully restricted (all false), so an unconfigured group can never write or see
// cost/suppliers — the safe default.
func loadPerms(tenant string, groupID int64) (Perms, error) {
	p, err := groupperms.Get(tenant, groupID)
	if err != nil {
		return Perms{}, fmt.Errorf("load group perms: %w", err)
	}
	return Perms{Edit: p.Edit, Suppliers: p.Suppliers, Cost: p.Cost}, nil
}

// historyOf fetches the group's recent context, accepting row-level failures
// (the model still gets the client's new messages without it).
func historyOf(groupID int64) []carshttp.GroupMessage {
	raw, err := carshttp.GroupHistory(groupID, historyLimit)
	if err != nil || len(raw) == 0 {
		return nil
	}
	out := make([]carshttp.GroupMessage, 0, len(raw))
	for _, r := range raw {
		if m, err := carshttp.ParseGroupMessage(r); err == nil {
			out = append(out, m)
		}
	}
	return out
}
