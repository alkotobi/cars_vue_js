package carshttp

import (
	"bytes"
	"encoding/json"
	"testing"
)

// TestColumnPermissionOrderMatchesMap is the drift guard for the two literals
// in scope.go: every gated column must appear in both and their sizes must
// agree.
func TestColumnPermissionOrderMatchesMap(t *testing.T) {
	if len(ColumnPermission) != len(restrictedColumnOrder) {
		t.Fatalf("ColumnPermission (%d) and restrictedColumnOrder (%d) drifted apart",
			len(ColumnPermission), len(restrictedColumnOrder))
	}
	for _, col := range restrictedColumnOrder {
		if _, ok := ColumnPermission[col]; !ok {
			t.Errorf("restrictedColumnOrder has %q but ColumnPermission has no entry", col)
		}
	}
	for col := range ColumnPermission {
		found := false
		for _, ordered := range restrictedColumnOrder {
			if ordered == col {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ColumnPermission has %q but restrictedColumnOrder does not", col)
		}
	}
}

// objectKeys returns the keys of a JSON object in their original order.
func objectKeys(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		t.Fatalf("not a JSON object: %s", raw)
	}
	var keys []string
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			t.Fatalf("bad key token: %v", err)
		}
		keys = append(keys, kt.(string))
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("bad value: %v", err)
		}
	}
	return keys
}

func TestScrubRestrictedColumns(t *testing.T) {
	row := json.RawMessage(`{"id":18,"vin":"LW9","car_name":"GOLF","color":"GREY","status":"Available","price_cell":"22000.00","freight":null,"rate":"230","cfr_da":"5060.00","buy_price":"18000.00"}`)
	allCols := []string{"id", "vin", "car_name", "color", "status", "price_cell", "freight", "rate", "cfr_da", "buy_price"}

	tests := []struct {
		name  string
		perms []string
		want  []string
	}{
		{name: "no snapshot keeps everything", perms: nil, want: allCols},
		{
			name:  "full admin keeps everything",
			perms: []string{"can_c_freight", "can_c_fob_price", "can_c_cfr_prices", "can_c_buy_price", "can_c_car_stock"},
			want:  allCols,
		},
		{
			name:  "ai role keeps only common columns",
			perms: []string{"can_c_car_stock"},
			want:  []string{"id", "vin", "car_name", "color", "status"},
		},
		{
			name:  "only fob granted keeps price_cell",
			perms: []string{"can_c_fob_price"},
			want:  []string{"id", "vin", "car_name", "color", "status", "price_cell"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := ScrubRestrictedColumns([]json.RawMessage{row}, tt.perms)
			if len(out) != 1 {
				t.Fatalf("expected 1 row, got %d", len(out))
			}
			got := objectKeys(t, out[0])
			if len(got) != len(tt.want) {
				t.Fatalf("keys %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("key[%d] = %q, want %q (full: %v)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

// TestScrubRestrictedColumnsFastPath is the scalability guard from RULES.md:
// when nothing is blocked the original bytes must pass through untouched.
func TestScrubRestrictedColumnsFastPath(t *testing.T) {
	row := json.RawMessage(`{"id":1,"car_name":"A3"}`)
	out := ScrubRestrictedColumns([]json.RawMessage{row}, []string{"can_c_fob_price", "can_c_freight", "can_c_cfr_prices", "can_c_buy_price"})
	if len(out) != 1 || string(out[0]) != string(row) {
		t.Fatalf("fast path should forward the row unchanged: %s", out)
	}
}