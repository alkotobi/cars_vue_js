package carshttp

import "testing"

func TestBuildCreateSupplierQuery(t *testing.T) {
	tests := []struct {
		name        string
		supplier    string
		contactInfo string
		notes       string
		wantParams  []any
	}{
		{
			name:       "name only stores NULL contact and notes",
			supplier:   "abi",
			wantParams: []any{"abi", nil, nil},
		},
		{
			name:        "contact and notes bind as given",
			supplier:    "abi",
			contactInfo: "+213555000000",
			notes:       "demo supplier",
			wantParams:  []any{"abi", "+213555000000", "demo supplier"},
		},
		{
			name:       "notes only leaves contact NULL",
			supplier:   "abi",
			notes:      "hello",
			wantParams: []any{"abi", nil, "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildCreateSupplierQuery(tt.supplier, tt.contactInfo, tt.notes)
			const wantSQL = "INSERT INTO suppliers (name, contact_info, notes) VALUES (?, ?, ?)"
			if sql != wantSQL {
				t.Errorf("SQL = %q, want %q", sql, wantSQL)
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestNullIfEmpty(t *testing.T) {
	if nullIfEmpty("") != nil {
		t.Error(`nullIfEmpty("") should be nil`)
	}
	if nullIfEmpty("x") != "x" {
		t.Errorf(`nullIfEmpty("x") = %v, want "x"`, nullIfEmpty("x"))
	}
}
