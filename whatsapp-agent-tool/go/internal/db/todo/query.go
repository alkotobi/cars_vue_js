package todo

import (
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// ListFilters narrows List. Zero values mean "no filter" (OwnerProfile "" = every
// login). OnlyRoots and ParentID are mutually exclusive.
type ListFilters struct {
	OwnerProfile string
	Status       string
	Priority     string
	OnlyRoots    bool
	ParentID     *int64
	Overdue      bool
	DueBefore    string
	Limit        int
}

// todosOrder keeps open/urgent/due-soon first, then the rest, stable by id.
const todosOrder = " ORDER BY FIELD(t.status,'open','done','cancelled'), " +
	"FIELD(t.priority,'urgent','high','normal','low'), t.due_at IS NULL, t.due_at ASC, t.id ASC"

// List returns todos matching the filters (ordered for action).
func List(f ListFilters) ([]Todo, error) {
	q := "SELECT " + todoCols + todoFrom + " WHERE 1=1"
	var params []any
	if p := strings.TrimSpace(f.OwnerProfile); p != "" {
		q += " AND u.profile = ?"
		params = append(params, p)
	}
	if f.Status != "" {
		q += " AND t.status = ?"
		params = append(params, f.Status)
	}
	if f.Priority != "" {
		q += " AND t.priority = ?"
		params = append(params, f.Priority)
	}
	if f.OnlyRoots {
		q += " AND t.parent_id IS NULL"
	}
	if f.ParentID != nil {
		q += " AND t.parent_id = ?"
		params = append(params, *f.ParentID)
	}
	if f.Overdue {
		q += " AND t.status = 'open' AND t.due_at IS NOT NULL AND t.due_at < ?"
		params = append(params, nowSQL())
	}
	if strings.TrimSpace(f.DueBefore) != "" {
		q += " AND t.due_at IS NOT NULL AND t.due_at <= ?"
		params = append(params, sqlDatetime(f.DueBefore))
	}
	q += todosOrder + " LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

// TodoDetail is a todo plus its direct children, reminders and status history.
type TodoDetail struct {
	Todo      Todo         `json:"todo"`
	Children  []Todo       `json:"children"`
	Reminders []Reminder   `json:"reminders"`
	History   []HistoryRow `json:"history"`
}

// Get returns one todo with its direct children, reminders and history (nil when
// the id does not exist).
func Get(id int64) (*TodoDetail, error) {
	t, err := getTodo(id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	childRows, err := db.Query("SELECT "+todoCols+todoFrom+" WHERE t.parent_id = ?"+todosOrder, id)
	if err != nil {
		return nil, err
	}
	reminders, err := ListReminders(ReminderFilters{TodoID: int64Ptr(id)})
	if err != nil {
		return nil, err
	}
	history, err := History(id)
	if err != nil {
		return nil, err
	}
	return &TodoDetail{Todo: *t, Children: mapRows(childRows), Reminders: reminders, History: history}, nil
}

// Subtree returns a todo and all of its descendants (breadth-first). Nil when the
// id does not exist.
func Subtree(id int64) ([]Todo, error) {
	root, err := getTodo(id)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, nil
	}
	out := []Todo{*root}
	frontier := []int64{id}
	for len(frontier) > 0 {
		ph, args := inClause(frontier)
		rows, err := db.Query("SELECT "+todoCols+todoFrom+" WHERE t.parent_id IN ("+ph+")"+todosOrder, args...)
		if err != nil {
			return nil, err
		}
		children := mapRows(rows)
		if len(children) == 0 {
			break
		}
		out = append(out, children...)
		next := make([]int64, 0, len(children))
		for _, c := range children {
			next = append(next, c.ID)
		}
		frontier = next
	}
	return out, nil
}

// History returns a todo's status history, oldest first.
func History(id int64) ([]HistoryRow, error) {
	rows, err := db.Query(
		"SELECT id, todo_id, from_status, to_status, changed_by, auto, reason, changed_at "+
			"FROM todo_status_history WHERE todo_id = ? ORDER BY changed_at ASC, id ASC", id)
	if err != nil {
		return nil, err
	}
	out := make([]HistoryRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, HistoryRow{
			ID:         int64(fnum(r["id"])),
			TodoID:     int64(fnum(r["todo_id"])),
			FromStatus: optStr(r["from_status"]),
			ToStatus:   s(r["to_status"]),
			ChangedBy:  optStr(r["changed_by"]),
			Auto:       fnum(r["auto"]) != 0,
			Reason:     optStr(r["reason"]),
			ChangedAt:  db.ToIso(r["changed_at"]),
		})
	}
	return out, nil
}

func getTodo(id int64) (*Todo, error) {
	rows, err := db.Query("SELECT "+todoCols+todoFrom+" WHERE t.id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	t := mapTodo(rows[0])
	return &t, nil
}

func mapRows(rows []map[string]any) []Todo {
	out := make([]Todo, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapTodo(r))
	}
	return out
}

func clampLimit(n int) int {
	if n <= 0 {
		return 100
	}
	if n > 500 {
		return 500
	}
	return n
}

func inClause(ids []int64) (string, []any) {
	parts := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		parts[i] = "?"
		args[i] = id
	}
	return strings.Join(parts, ","), args
}
