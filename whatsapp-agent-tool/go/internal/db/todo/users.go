package todo

import (
	"fmt"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// User is one row of todo_users — a WhatsApp login that owns todos.
type User struct {
	ID          int64   `json:"id"`
	Profile     string  `json:"profile"`
	DisplayName *string `json:"displayName,omitempty"`
	Number      *string `json:"number,omitempty"`
	CreatedAt   *string `json:"createdAt"`
}

// EnsureUser returns the todo_users.id for a login (WA_PROFILE slug), creating the
// row on first use. A non-empty displayName/number refreshes the stored label so
// profiles.json stays the source of truth for the human name. Falls back to
// "whatsapp-profile" for the default account.
func EnsureUser(profile, displayName, number string) (int64, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		profile = "whatsapp-profile"
	}
	if _, err := db.Execute("INSERT IGNORE INTO todo_users (profile, display_name, number, created_at) VALUES (?,?,?,?)",
		profile, nullIfEmpty(displayName), nullIfEmpty(number), nowSQL()); err != nil {
		return 0, err
	}
	if strings.TrimSpace(displayName) != "" || strings.TrimSpace(number) != "" {
		if _, err := db.Execute(
			"UPDATE todo_users SET display_name = COALESCE(?, display_name), number = COALESCE(?, number) WHERE profile = ?",
			nullIfEmpty(displayName), nullIfEmpty(number), profile); err != nil {
			return 0, err
		}
	}
	id, found, err := FindUserID(profile)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("todo user %q could not be resolved", profile)
	}
	return id, nil
}

// FindUserID resolves a login slug to its id (false when unknown).
func FindUserID(profile string) (int64, bool, error) {
	row, err := db.QueryOne("SELECT id FROM todo_users WHERE profile = ?", strings.TrimSpace(profile))
	if err != nil {
		return 0, false, err
	}
	if row == nil {
		return 0, false, nil
	}
	return int64(fnum(row["id"])), true, nil
}

// ListUsers returns every registered login, ordered by profile.
func ListUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, profile, display_name, number, created_at FROM todo_users ORDER BY profile ASC")
	if err != nil {
		return nil, err
	}
	out := make([]User, 0, len(rows))
	for _, r := range rows {
		out = append(out, User{
			ID:          int64(fnum(r["id"])),
			Profile:     s(r["profile"]),
			DisplayName: optStr(r["display_name"]),
			Number:      optStr(r["number"]),
			CreatedAt:   db.ToIso(r["created_at"]),
		})
	}
	return out, nil
}
