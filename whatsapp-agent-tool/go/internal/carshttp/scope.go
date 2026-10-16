// Column scoping for stock rows.
//
// The cars app gates stock columns with permissions client-side (see
// CarStockTable.vue's v-if columns). The API itself returns every column to any
// token, so the AI tool enforces the same visibility from its cached permission
// list: a column whose permission the AI user's role lacks is dropped from the
// rows before they are printed.
package carshttp

import (
	"bytes"
	"encoding/json"
	"strings"
)

// ColumnPermission maps a stock-row column to the app permission that gates it.
// Keep in sync with CarStockTable.vue (can_see_freight / can_see_fob_price /
// can_see_cfr_prices) and with the buy price column the tool adds itself.
var ColumnPermission = map[string]string{
	"freight":    "can_c_freight",
	"price_cell": "can_c_fob_price",
	"rate":       "can_c_cfr_prices",
	"cfr_da":     "can_c_cfr_prices",
	"buy_price":  "can_c_buy_price",
}

// restrictedColumnOrder pins the order in which permission checks are applied.
// A test asserts it matches ColumnPermission exactly (no drift allowed).
var restrictedColumnOrder = []string{
	"price_cell", "freight", "rate", "cfr_da", "buy_price",
}

// ScrubRestrictedColumns drops every permission-gated column the caller lacks.
//
// perms is the AI user's granted permission names. When perms is empty there is
// no permission snapshot (the token came from CARS_API_TOKEN, or the store has
// none), so every column is kept — matching HasPermission's known=false skip.
// Otherwise rows are returned unchanged unless a gated column must be removed,
// so callers pay nothing on the fast path.
func ScrubRestrictedColumns(rows []json.RawMessage, perms []string) []json.RawMessage {
	if len(perms) == 0 {
		return rows
	}

	granted := make(map[string]bool, len(perms))
	for _, p := range perms {
		granted[p] = true
	}

	var blocked map[string]bool
	for _, col := range restrictedColumnOrder {
		if granted[ColumnPermission[col]] {
			continue
		}
		if blocked == nil {
			blocked = make(map[string]bool, len(restrictedColumnOrder))
		}
		blocked[col] = true
	}

	if blocked == nil {
		return rows
	}

	out := make([]json.RawMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, dropColumns(row, blocked))
	}
	return out
}

// dropColumns removes the given keys from one JSON object, preserving the
// remaining keys' original order and formatting (the raw bytes of each value).
func dropColumns(row json.RawMessage, blocked map[string]bool) json.RawMessage {
	dec := json.NewDecoder(bytes.NewReader(row))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		// Not an object (should not happen for stock rows) — forward untouched.
		return row
	}

	var b strings.Builder
	b.WriteByte('{')
	wrote := false
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return row
		}
		key := keyTok.(string)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return row
		}
		if blocked[key] {
			continue
		}
		if wrote {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(strings.ReplaceAll(key, `"`, `\"`))
		b.WriteString(`":`)
		b.Write(raw)
		wrote = true
	}
	b.WriteByte('}')
	return json.RawMessage(b.String())
}
