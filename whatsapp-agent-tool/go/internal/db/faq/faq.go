// Package faq is the Go twin of src/db/faq.ts — FAQ store in MySQL
// (repeated client questions + approved Arabic answers).
package faq

import (
	"fmt"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

type Entry struct {
	Id        int    `json:"id"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Context   any    `json:"context"`
	Active    bool   `json:"active"`
	AddedAt   any    `json:"addedAt"`
	UpdatedAt any    `json:"updatedAt"`
}

const cols = "id, question, answer, context, active, added_at, updated_at"

func mapRow(r map[string]any) Entry {
	return Entry{
		Id:        int(num(r["id"])),
		Question:  str(r["question"]),
		Answer:    str(r["answer"]),
		Context:   raw(r["context"]),
		Active:    boolish(r["active"]),
		AddedAt:   db.ToIso(r["added_at"]),
		UpdatedAt: db.ToIso(r["updated_at"]),
	}
}

// Search uses FULLTEXT first with a LIKE fallback, mirroring Node's behaviour.
func Search(term string, limit int) ([]Entry, error) {
	t := strings.TrimSpace(term)
	if t == "" {
		return []Entry{}, nil
	}
	rows, err := db.Query(
		"SELECT "+cols+" FROM faq WHERE active = 1 AND MATCH(question, answer) AGAINST (? IN NATURAL LANGUAGE MODE) "+
			"ORDER BY MATCH(question, answer) AGAINST (? IN NATURAL LANGUAGE MODE) DESC LIMIT ?",
		t, t, limit,
	)
	if err == nil && len(rows) > 0 {
		return mapRows(rows), nil
	}
	// FULLTEXT unavailable — fall back to LIKE (Node swallows the error silently).
	like := "%" + t + "%"
	rows, err = db.Query(
		"SELECT "+cols+" FROM faq WHERE active = 1 AND (question LIKE ? OR answer LIKE ? OR context LIKE ?) "+
			"ORDER BY updated_at DESC LIMIT ?",
		like, like, like, limit,
	)
	if err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

func List(includeInactive bool) ([]Entry, error) {
	where := ""
	if !includeInactive {
		where = " WHERE active = 1"
	}
	rows, err := db.Query("SELECT " + cols + " FROM faq" + where + " ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

func Get(id int) (*Entry, error) {
	row, err := db.QueryOne("SELECT "+cols+" FROM faq WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	e := mapRow(row)
	return &e, nil
}

// Add inserts and returns affectedRows (Node's addFaq returns execute()'s
// affectedRows, which the CLI reports as `id`).
func Add(question, answer string, context any) (int64, error) {
	return db.Execute("INSERT INTO faq (question, answer, context) VALUES (?, ?, ?)", question, answer, raw(context))
}

func Update(id int, patch UpdatePatch) (bool, error) {
	var sets []string
	var params []any
	if patch.Question != nil {
		sets = append(sets, "question = ?")
		params = append(params, *patch.Question)
	}
	if patch.Answer != nil {
		sets = append(sets, "answer = ?")
		params = append(params, *patch.Answer)
	}
	if patch.ContextSet {
		sets = append(sets, "context = ?")
		params = append(params, raw(patch.Context))
	}
	if patch.ActiveSet {
		v := 0
		if patch.Active {
			v = 1
		}
		sets = append(sets, "active = ?")
		params = append(params, v)
	}
	if len(sets) == 0 {
		return false, nil
	}
	params = append(params, id)
	affected, err := db.Execute(fmt.Sprintf("UPDATE faq SET %s WHERE id = ?", strings.Join(sets, ", ")), params...)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

type UpdatePatch struct {
	Question   *string
	Answer     *string
	Context    any
	ContextSet bool
	Active     bool
	ActiveSet  bool
}

func mapRows(rows []map[string]any) []Entry {
	out := make([]Entry, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapRow(r))
	}
	return out
}

func num(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolish(v any) bool {
	switch t := v.(type) {
	case float64:
		return t != 0
	case int64:
		return t != 0
	case uint64:
		return t != 0
	case []byte:
		return string(t) != "0" && string(t) != ""
	case bool:
		return t
	}
	return false
}

func raw(v any) any {
	if v == nil {
		return nil
	}
	return v
}
