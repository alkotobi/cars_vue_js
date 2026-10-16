// Chat helpers for the cars-app chat subsystem (`whythechat` tables
// chat_groups / chat_messages / chat_users). The AI bot uses these to learn
// which groups the AI user belongs to, fetch client messages and post user
// replies — all over the same `query` passthrough the app uses, with fixed
// parameterised SQL mirroring ChatSidebar.vue / ChatMessages.vue.
package carshttp

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ChatGroup is one chat group the AI user is a member of (id + name from
// chat_groups, exactly the two columns the bot needs for context).
type ChatGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GroupMessage is one chat_messages row parsed from the passthrough rows,
// shaped like the app's fetchNewMessages select so the prompt can show who
// wrote what.
type GroupMessage struct {
	ID          int64  `json:"id"`
	IDChatGroup int64  `json:"id_chat_group"`
	Message     string `json:"message"`
	Time        string `json:"time"`
	SenderName  string `json:"sender_username"`
	SenderRole  string `json:"sender_role"`
	SenderType  string `json:"sender_type"`
}

// ParseGroupMessage decodes one raw passthrough row.
func ParseGroupMessage(raw json.RawMessage) (GroupMessage, error) {
	var m GroupMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	return m, nil
}

// MemberGroups lists the chat groups the AI user is an active member of,
// mirroring ChatSidebar.vue's user-mode query (without the aggregate columns).
func MemberGroups(userID int64) ([]ChatGroup, error) {
	const q = "" +
		"SELECT DISTINCT cg.id, cg.name " +
		"FROM chat_groups cg " +
		"INNER JOIN chat_users cu_member ON cg.id = cu_member.id_chat_group " +
		"WHERE cg.is_active = 1 AND cu_member.is_active = 1 AND cu_member.id_user = ? " +
		"ORDER BY cg.name ASC"
	rows, err := QueryAsAI(q, []any{userID})
	if err != nil {
		return nil, err
	}
	groups := make([]ChatGroup, 0, len(rows))
	for _, raw := range rows {
		var g ChatGroup
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, &ApiError{Code: "bad_response", Message: "invalid group row"}
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// AllGroups lists every active chat group in the tenant (id + name), for the
// operator's `group-perms` command to resolve a group name to its id. Distinct
// from MemberGroups: the AI user need not be a member.
func AllGroups() ([]ChatGroup, error) {
	const q = "SELECT id, name FROM chat_groups WHERE is_active = 1 ORDER BY name ASC"
	rows, err := QueryAsAI(q, nil)
	if err != nil {
		return nil, err
	}
	groups := make([]ChatGroup, 0, len(rows))
	for _, raw := range rows {
		var g ChatGroup
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, &ApiError{Code: "bad_response", Message: "invalid group row"}
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// GroupMaxMessageID returns the highest chat_messages id in a group, or 0 for
// an empty group. The bot uses it to seed its cursor on first run so it never
// answers past history.
func GroupMaxMessageID(groupID int64) (int64, error) {
	const q = "SELECT COALESCE(MAX(id), 0) AS max_id " +
		"FROM chat_messages WHERE id_chat_group = CAST(? AS UNSIGNED)"
	rows, err := QueryAsAI(q, []any{groupID})
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	var r struct {
		MaxID json.Number `json:"max_id"`
	}
	if err := json.Unmarshal(rows[0], &r); err != nil {
		return 0, &ApiError{Code: "bad_response", Message: "invalid max_id row"}
	}
	n, err := r.MaxID.Int64()
	if err != nil {
		return 0, &ApiError{Code: "bad_response", Message: "invalid max_id value"}
	}
	return n, nil
}

// NewMessages returns the messages newer than afterID in one group, oldest
// first, authored by anyone EXCEPT the AI account itself (excludeUserID): both
// client messages and staff/user messages. The bot answers whoever writes in a
// group it is in, but never replies to its own posts.
func NewMessages(groupID, afterID, excludeUserID int64) ([]json.RawMessage, error) {
	const q = "" +
		"SELECT cm.id, cm.id_chat_group, cm.message_from_user_id, cm.message_from_client_id, " +
		"cm.message, cm.time, COALESCE(u.username, c.name) AS sender_username, " +
		"COALESCE(r.role_name, 'Client') AS sender_role, " +
		"CASE WHEN cm.message_from_user_id IS NOT NULL THEN 'user' ELSE 'client' END AS sender_type " +
		"FROM chat_messages cm " +
		"LEFT JOIN users u ON cm.message_from_user_id = u.id " +
		"LEFT JOIN roles r ON u.role_id = r.id " +
		"LEFT JOIN clients c ON cm.message_from_client_id = c.id " +
		"WHERE cm.id_chat_group = CAST(? AS UNSIGNED) " +
		"AND cm.id > CAST(? AS UNSIGNED) " +
		"AND (cm.message_from_client_id IS NOT NULL " +
		"OR (cm.message_from_user_id IS NOT NULL AND cm.message_from_user_id <> CAST(? AS UNSIGNED))) " +
		"ORDER BY cm.id ASC"
	return QueryAsAI(q, []any{groupID, afterID, excludeUserID})
}

// GroupHistory returns the last limit messages of a group, oldest first, any
// sender — conversational context for the model. The limit is inlined exactly
// like ChatMessages.vue (`LIMIT ${messagesPerPage}`): MariaDB rejects a bound
// `LIMIT ?` placeholder, and the value is a fixed cap set by the caller, never
// user input.
func GroupHistory(groupID int64, limit int) ([]json.RawMessage, error) {
	q := "" +
		"SELECT * FROM (" +
		"SELECT cm.id, cm.id_chat_group, cm.message_from_user_id, cm.message_from_client_id, " +
		"cm.message, cm.time, COALESCE(u.username, c.name) AS sender_username, " +
		"COALESCE(r.role_name, 'Client') AS sender_role, " +
		"CASE WHEN cm.message_from_user_id IS NOT NULL THEN 'user' ELSE 'client' END AS sender_type " +
		"FROM chat_messages cm " +
		"LEFT JOIN users u ON cm.message_from_user_id = u.id " +
		"LEFT JOIN roles r ON u.role_id = r.id " +
		"LEFT JOIN clients c ON cm.message_from_client_id = c.id " +
		"WHERE cm.id_chat_group = CAST(? AS UNSIGNED) " +
		"ORDER BY cm.id DESC LIMIT " + strconv.Itoa(limit) +
		") z ORDER BY z.id ASC"
	return QueryAsAI(q, []any{groupID})
}

// LatestMessage returns the single newest message in a group (any sender), or
// ok=false for an empty group. The bot uses its timestamp to tell a group that
// just started (answer its last message) from one discovered with stale
// history (seed past it).
func LatestMessage(groupID int64) (GroupMessage, bool, error) {
	rows, err := GroupHistory(groupID, 1)
	if err != nil {
		return GroupMessage{}, false, err
	}
	if len(rows) == 0 {
		return GroupMessage{}, false, nil
	}
	m, err := ParseGroupMessage(rows[0])
	if err != nil {
		return GroupMessage{}, false, &ApiError{Code: "bad_response", Message: "invalid latest message row"}
	}
	return m, true, nil
}

// PostUserMessage sends a user-authored message into a group as the AI user,
// mirroring ChatMessages.vue's user-mode INSERT.
func PostUserMessage(groupID, userID int64, text string) (*WriteResult, error) {
	const q = "" +
		"INSERT INTO chat_messages (id_chat_group, message_from_user_id, message, time) " +
		"VALUES (?, ?, ?, UTC_TIMESTAMP())"
	res, err := ExecAsAI(q, []any{groupID, userID, text})
	if err != nil {
		return nil, err
	}
	if res.LastInsertID == 0 {
		return nil, &ApiError{Code: "query_failed", Message: fmt.Sprintf("message insert for group %d reported no row id", groupID)}
	}
	return res, nil
}

// UserIDOrLogin returns the AI user's id, refreshing the cached login first if
// the store has credentials but no id yet (e.g. a session seeded via
// CARS_API_TOKEN). ok=false when there is genuinely no id to rely on.
func UserIDOrLogin() (int, bool, error) {
	id, ok, err := CurrentUserID()
	if err != nil {
		return 0, false, err
	}
	if ok {
		return id, true, nil
	}
	if _, err := LoginAndCache(); err != nil {
		return 0, false, err
	}
	return CurrentUserID()
}
