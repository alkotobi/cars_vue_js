package chatbot

import (
	"encoding/json"
	"fmt"

	"whatsapp-agent-tool/internal/carshttp"
	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db/groupperms"
)

// carsTransport is the cars-app implementation of Transport. It is the ONLY
// place the loop's channel I/O is bound to carshttp: everything else in this
// package speaks Group/Message/Perms, so a second channel is added by writing
// one more Transport, not by touching the loop.
type carsTransport struct{}

// NewTransport returns the cars-app chat transport the chat-poll daemon runs.
func NewTransport() Transport { return carsTransport{} }

// Identity resolves the AI user's id, refreshing the cached login first when
// there are credentials but no id yet.
func (carsTransport) Identity() (int64, bool, error) {
	id, ok, err := carshttp.UserIDOrLogin()
	return int64(id), ok, err
}

// Groups lists the chat groups the AI user is an active member of.
func (carsTransport) Groups(userID int64) ([]Group, error) {
	gs, err := carshttp.MemberGroups(userID)
	if err != nil {
		return nil, err
	}
	out := make([]Group, 0, len(gs))
	for _, g := range gs {
		out = append(out, Group{ID: g.ID, Name: g.Name})
	}
	return out, nil
}

// Perms reads the group's capabilities from the ai DB. A missing row is fully
// restricted (all false) — groupperms.Get's safe default — so an unconfigured
// group can never write or see cost/suppliers.
func (carsTransport) Perms(groupID int64) (Perms, error) {
	p, err := groupperms.Get(config.C.Cars.Tenant, groupID)
	if err != nil {
		return Perms{}, fmt.Errorf("load group perms: %w", err)
	}
	return Perms{Edit: p.Edit, Suppliers: p.Suppliers, Cost: p.Cost}, nil
}

// Read fetches the new client/staff messages of one group (bot's own excluded).
func (carsTransport) Read(groupID, afterID, excludeUserID int64) ([]Message, error) {
	rows, err := carshttp.NewMessages(groupID, afterID, excludeUserID)
	if err != nil {
		return nil, err
	}
	return parseMessages(rows)
}

// History fetches the group's recent context.
func (carsTransport) History(groupID int64, limit int) ([]Message, error) {
	rows, err := carshttp.GroupHistory(groupID, limit)
	if err != nil {
		return nil, err
	}
	return parseMessages(rows)
}

// MaxID is the newest message id of a group (0 when empty).
func (carsTransport) MaxID(groupID int64) (int64, error) {
	return carshttp.GroupMaxMessageID(groupID)
}

// Latest is the newest message of a group, used by the cursor seeding rule.
func (carsTransport) Latest(groupID int64) (Message, bool, error) {
	m, ok, err := carshttp.LatestMessage(groupID)
	if err != nil || !ok {
		return Message{}, false, err
	}
	return Message{
		ID:         m.ID,
		GroupID:    m.IDChatGroup,
		Text:       m.Message,
		Time:       m.Time,
		SenderName: m.SenderName,
		SenderRole: m.SenderRole,
		SenderType: m.SenderType,
	}, true, nil
}

// Send inserts the reply as the AI user, mirroring ChatMessages.vue.
func (carsTransport) Send(groupID, userID int64, text string) error {
	_, err := carshttp.PostUserMessage(groupID, userID, text)
	return err
}

// parseMessages decodes the raw passthrough rows into the loop's Message type.
// The tags line up with the select the cars app returns, so no field mapping is
// needed — this is the only decoding step of the pipeline.
func parseMessages(rows []json.RawMessage) ([]Message, error) {
	out := make([]Message, 0, len(rows))
	for _, raw := range rows {
		var m Message
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("parse message row: %w", err)
		}
		out = append(out, m)
	}
	return out, nil
}
