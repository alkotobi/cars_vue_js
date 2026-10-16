package todo

import (
	"fmt"
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// Reminder is one row of todo_reminders.
type Reminder struct {
	ID        int64   `json:"id"`
	TodoID    int64   `json:"todoId"`
	RemindAt  *string `json:"remindAt"`
	Channel   string  `json:"channel"`
	Status    string  `json:"status"`
	SentAt    *string `json:"sentAt"`
	Note      *string `json:"note,omitempty"`
	CreatedAt *string `json:"createdAt"`
}

// ReminderStatus values.
var ReminderStatuses = []string{"pending", "sent", "cancelled"}

func validReminderStatus(s string) bool {
	for _, v := range ReminderStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// ReminderFilters narrows ListReminders.
type ReminderFilters struct {
	TodoID *int64
	Status string
	Limit  int
}

// ListReminders returns reminders, soonest first.
func ListReminders(f ReminderFilters) ([]Reminder, error) {
	q := "SELECT id, todo_id, remind_at, channel, status, sent_at, note, created_at FROM todo_reminders WHERE 1=1"
	var params []any
	if f.TodoID != nil {
		q += " AND todo_id = ?"
		params = append(params, *f.TodoID)
	}
	if f.Status != "" {
		q += " AND status = ?"
		params = append(params, f.Status)
	}
	q += " ORDER BY remind_at ASC, id ASC LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	out := make([]Reminder, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapReminder(r))
	}
	return out, nil
}

// AddReminder attaches a reminder to a todo. remindAt accepts a date or datetime.
func AddReminder(todoID int64, remindAt, channel, note string) (int64, error) {
	when, ok := parseWhen(remindAt)
	if !ok {
		return 0, fmt.Errorf("--at must be a date or datetime (e.g. 2026-10-20 09:30)")
	}
	if strings.TrimSpace(channel) == "" {
		channel = "whatsapp"
	}
	if channel != "whatsapp" && channel != "local" {
		return 0, fmt.Errorf("--channel must be whatsapp or local")
	}
	t, err := getTodo(todoID)
	if err != nil {
		return 0, err
	}
	if t == nil {
		return 0, fmt.Errorf("todo %d not found", todoID)
	}
	return db.Insert("INSERT INTO todo_reminders (todo_id, remind_at, channel, status, note, created_at) VALUES (?,?,?,?,?,?)",
		todoID, when.UTC().Format("2006-01-02 15:04:05.000"), channel, "pending", nullIfEmpty(note), nowSQL())
}

// SetReminderStatus flips a reminder's status (marks sent_at when sent).
func SetReminderStatus(id int64, status string) (int64, error) {
	if !validReminderStatus(status) {
		return 0, fmt.Errorf("invalid reminder status %q; one of %s", status, strings.Join(ReminderStatuses, ", "))
	}
	var affected int64
	var err error
	if status == "sent" {
		affected, err = db.Execute("UPDATE todo_reminders SET status = ?, sent_at = ? WHERE id = ?", status, nowSQL(), id)
	} else {
		affected, err = db.Execute("UPDATE todo_reminders SET status = ?, sent_at = NULL WHERE id = ?", status, id)
	}
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, fmt.Errorf("reminder %d not found", id)
	}
	return affected, nil
}

// DueReminder is a pending reminder joined to its (still open) todo — the payload
// the `pending` report and `todo due` surface.
type DueReminder struct {
	ReminderID  int64   `json:"reminderId"`
	TodoID      int64   `json:"todoId"`
	RemindAt    *string `json:"remindAt"`
	Channel     string  `json:"channel"`
	Note        *string `json:"note,omitempty"`
	Title       string  `json:"title"`
	Owner       string  `json:"owner"`
	Priority    string  `json:"priority"`
	DueAt       *string `json:"dueAt"`
	RequestedBy *string `json:"requestedBy,omitempty"`
	ChatRef     *string `json:"chatRef,omitempty"`
}

// Due returns pending reminders whose time has passed for todos that are still
// open. before "" = now.
func Due(before string) ([]DueReminder, error) {
	cutoff := nowSQL()
	if strings.TrimSpace(before) != "" {
		cutoff = ""
		if w, ok := parseWhen(before); ok {
			cutoff = w.UTC().Format("2006-01-02 15:04:05.000")
		} else {
			return nil, fmt.Errorf("--before must be a date or datetime")
		}
	}
	rows, err := db.Query(
		"SELECT r.id AS reminder_id, r.todo_id, r.remind_at, r.channel, r.note, "+
			"t.title, t.priority, t.due_at, t.requested_by, t.chat_ref, u.profile AS owner "+
			"FROM todo_reminders r "+
			"JOIN todos t ON t.id = r.todo_id "+
			"JOIN todo_users u ON u.id = t.owner_id "+
			"WHERE r.status = 'pending' AND r.remind_at <= ? AND t.status = 'open' "+
			"ORDER BY r.remind_at ASC, r.id ASC LIMIT 200", cutoff)
	if err != nil {
		return nil, err
	}
	out := make([]DueReminder, 0, len(rows))
	for _, r := range rows {
		out = append(out, DueReminder{
			ReminderID:  int64(fnum(r["reminder_id"])),
			TodoID:      int64(fnum(r["todo_id"])),
			RemindAt:    db.ToIso(r["remind_at"]),
			Channel:     s(r["channel"]),
			Note:        optStr(r["note"]),
			Title:       s(r["title"]),
			Owner:       s(r["owner"]),
			Priority:    s(r["priority"]),
			DueAt:       db.ToIso(r["due_at"]),
			RequestedBy: optStr(r["requested_by"]),
			ChatRef:     optStr(r["chat_ref"]),
		})
	}
	return out, nil
}

func mapReminder(r map[string]any) Reminder {
	return Reminder{
		ID:        int64(fnum(r["id"])),
		TodoID:    int64(fnum(r["todo_id"])),
		RemindAt:  db.ToIso(r["remind_at"]),
		Channel:   s(r["channel"]),
		Status:    s(r["status"]),
		SentAt:    db.ToIso(r["sent_at"]),
		Note:      optStr(r["note"]),
		CreatedAt: db.ToIso(r["created_at"]),
	}
}
