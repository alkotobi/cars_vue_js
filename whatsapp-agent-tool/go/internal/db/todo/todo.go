// Package todo is the db layer for the ai database's task tables (migration 008):
// todo_users, todos, todo_reminders, todo_status_history.
//
// A todo belongs to the login the tool is running as (todo_users.profile). Subtasks
// nest without limit via parent_id (adjacency list): a parent with children is done
// iff all its children are resolved (done or cancelled) — the cascade in cascade.go.
// A childless todo keeps whatever status a human set on it.
package todo

import (
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// Status values (todos.status).
const (
	StatusOpen      = "open"
	StatusDone      = "done"
	StatusCancelled = "cancelled"
)

// Reminder status values (todo_reminders.status).
const (
	ReminderPending   = "pending"
	ReminderSent      = "sent"
	ReminderCancelled = "cancelled"
)

// Priorities lists the accepted priority values, low to high.
var Priorities = []string{"low", "normal", "high", "urgent"}

// ValidPriority reports whether p is one of the four levels.
func ValidPriority(p string) bool {
	for _, v := range Priorities {
		if p == v {
			return true
		}
	}
	return false
}

// ValidStatus reports whether s is a todos.status value.
func ValidStatus(s string) bool {
	return s == StatusOpen || s == StatusDone || s == StatusCancelled
}

// Todo is one row of todos joined to its owning login.
type Todo struct {
	ID          int64   `json:"id"`
	ParentID    *int64  `json:"parentId"`
	OwnerID     int64   `json:"ownerId"`
	Owner       string  `json:"owner"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueAt       *string `json:"dueAt"`
	RequestedBy *string `json:"requestedBy,omitempty"`
	ChatRef     *string `json:"chatRef,omitempty"`
	DoneAt      *string `json:"doneAt"`
	CreatedAt   *string `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
}

// HistoryRow is one row of todo_status_history.
type HistoryRow struct {
	ID         int64   `json:"id"`
	TodoID     int64   `json:"todoId"`
	FromStatus *string `json:"fromStatus"`
	ToStatus   string  `json:"toStatus"`
	ChangedBy  *string `json:"changedBy"`
	Auto       bool    `json:"auto"`
	Reason     *string `json:"reason"`
	ChangedAt  *string `json:"changedAt"`
}

// todoCols / todoFrom are the shared SELECT tail (always joined to todo_users).
const todoCols = "t.id, t.parent_id, t.owner_id, u.profile AS owner, t.title, t.description, " +
	"t.status, t.priority, t.due_at, t.requested_by, t.chat_ref, t.done_at, t.created_at, t.updated_at"
const todoFrom = " FROM todos t JOIN todo_users u ON u.id = t.owner_id"

func mapTodo(r map[string]any) Todo {
	return Todo{
		ID:          int64(fnum(r["id"])),
		ParentID:    int64PtrN(fnum(r["parent_id"])),
		OwnerID:     int64(fnum(r["owner_id"])),
		Owner:       s(r["owner"]),
		Title:       s(r["title"]),
		Description: optStr(r["description"]),
		Status:      s(r["status"]),
		Priority:    s(r["priority"]),
		DueAt:       db.ToIso(r["due_at"]),
		RequestedBy: optStr(r["requested_by"]),
		ChatRef:     optStr(r["chat_ref"]),
		DoneAt:      db.ToIso(r["done_at"]),
		CreatedAt:   db.ToIso(r["created_at"]),
		UpdatedAt:   db.ToIso(r["updated_at"]),
	}
}

// ---------------------------------------------------------------- helpers

func fnum(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case int:
		return float64(t)
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	}
	return 0
}

func s(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return ""
}

func optStr(v any) *string {
	if v == nil {
		return nil
	}
	out := s(v)
	if out == "" {
		return nil
	}
	return &out
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func int64Ptr(v int64) *int64 { return &v }

// int64PtrN returns a pointer only for positive ids (NULL/0 -> nil).
func int64PtrN(v float64) *int64 {
	if v <= 0 {
		return nil
	}
	out := int64(v)
	return &out
}

// sqlDatetime normalizes a user-supplied date/datetime for a DATETIME(3) column.
// "" stays nil (NULL); a date-only value ("2026-10-20") becomes end-of-day UTC.
func sqlDatetime(v string) any {
	t, ok := parseWhen(v)
	if !ok {
		return nil
	}
	return t.UTC().Format("2006-01-02 15:04:05.000")
}
