// Package groupperms is the store of per-chat-group permissions for the cars-app
// chat AI (table chat_group_permissions in the shared `ai` DB). A group with no
// row is fully restricted — every capability is false — so the zero value is
// safe to hand to the bot. Writes go through the operator-only `group-perms` CLI.
package groupperms

import (
	"strconv"

	"whatsapp-agent-tool/internal/db"
)

// Perms is one group's capability set. The JSON field names are the exact
// permission names the CLI takes and the bot exports (WBOT_PERMS).
type Perms struct {
	Tenant    string `json:"tenant"`
	GroupID   int64  `json:"id_chat_group"`
	Edit      bool   `json:"can_give_edit_instruction"`
	Suppliers bool   `json:"can_c_suppliers_info"`
	Cost      bool   `json:"can_c_cost"`
}

const cols = "tenant, id_chat_group, can_give_edit_instruction, can_c_suppliers_info, can_c_cost"

// Get returns the group's permissions, or an all-false Perms (nil error) when no
// row exists — an unconfigured group is restricted, never granted.
func Get(tenant string, groupID int64) (Perms, error) {
	row, err := db.QueryOne("SELECT "+cols+" FROM chat_group_permissions WHERE tenant = ? AND id_chat_group = ?", tenant, groupID)
	if err != nil {
		return Perms{}, err
	}
	if row == nil {
		return Perms{Tenant: tenant, GroupID: groupID}, nil
	}
	return mapRow(row), nil
}

// List returns the configured rows for one tenant, or for every tenant when
// tenant is empty.
func List(tenant string) ([]Perms, error) {
	var rows []map[string]any
	var err error
	if tenant == "" {
		rows, err = db.Query("SELECT " + cols + " FROM chat_group_permissions ORDER BY tenant, id_chat_group")
	} else {
		rows, err = db.Query("SELECT "+cols+" FROM chat_group_permissions WHERE tenant = ? ORDER BY id_chat_group", tenant)
	}
	if err != nil {
		return nil, err
	}
	out := make([]Perms, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapRow(r))
	}
	return out, nil
}

// Patch is a partial update: only non-nil flags are changed, so `group-perms set`
// can flip one capability without touching the others.
type Patch struct {
	Edit      *bool
	Suppliers *bool
	Cost      *bool
}

// Set applies patch to a group and returns the resulting row. It reads the
// current flags first (upserting a fresh all-false row is equivalent) and writes
// all three back, so an omitted flag keeps its stored value.
func Set(tenant string, groupID int64, patch Patch) (Perms, error) {
	cur, err := Get(tenant, groupID)
	if err != nil {
		return Perms{}, err
	}
	if patch.Edit != nil {
		cur.Edit = *patch.Edit
	}
	if patch.Suppliers != nil {
		cur.Suppliers = *patch.Suppliers
	}
	if patch.Cost != nil {
		cur.Cost = *patch.Cost
	}
	_, err = db.Execute(
		"INSERT INTO chat_group_permissions "+
			"(tenant, id_chat_group, can_give_edit_instruction, can_c_suppliers_info, can_c_cost) "+
			"VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE "+
			"can_give_edit_instruction = VALUES(can_give_edit_instruction), "+
			"can_c_suppliers_info = VALUES(can_c_suppliers_info), "+
			"can_c_cost = VALUES(can_c_cost)",
		cur.Tenant, cur.GroupID, boolInt(cur.Edit), boolInt(cur.Suppliers), boolInt(cur.Cost))
	if err != nil {
		return Perms{}, err
	}
	return cur, nil
}

func mapRow(r map[string]any) Perms {
	return Perms{
		Tenant:    str(r["tenant"]),
		GroupID:   int64(num(r["id_chat_group"])),
		Edit:      boolish(r["can_give_edit_instruction"]),
		Suppliers: boolish(r["can_c_suppliers_info"]),
		Cost:      boolish(r["can_c_cost"]),
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func num(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
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

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return ""
}

func boolish(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case float32:
		return t != 0
	case int64:
		return t != 0
	case uint64:
		return t != 0
	case int:
		return t != 0
	case []byte:
		s := string(t)
		return s != "0" && s != ""
	case string:
		return t != "0" && t != ""
	}
	return false
}
