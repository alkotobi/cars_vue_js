package carshttp

import (
	"encoding/json"
	"testing"
)

func TestBuildCreateBrandQuery(t *testing.T) {
	tests := []struct {
		name       string
		brand      string
		logoPath   string
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "brand only leaves logo_path to its default",
			brand:      "TOYOTA",
			logoPath:   "",
			wantSQL:    "INSERT INTO brands (brand) VALUES (?)",
			wantParams: []any{"TOYOTA"},
		},
		{
			name:       "brand with a logo binds both columns",
			brand:      "TOYOTA",
			logoPath:   "uploads/ab12.png",
			wantSQL:    "INSERT INTO brands (brand, logo_path) VALUES (?, ?)",
			wantParams: []any{"TOYOTA", "uploads/ab12.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildCreateBrandQuery(tt.brand, tt.logoPath)
			if sql != tt.wantSQL {
				t.Errorf("SQL = %q, want %q", sql, tt.wantSQL)
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestFlexInt64(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int64
	}{
		{name: "bare number", raw: `12`, want: 12},
		{name: "quoted number", raw: `"12"`, want: 12},
		{name: "empty is zero", raw: ``, want: 0},
		{name: "null is zero", raw: `null`, want: 0},
		{name: "non-numeric is zero", raw: `"nope"`, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flexInt64(json.RawMessage(tt.raw)); got != tt.want {
				t.Errorf("flexInt64(%s) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}
