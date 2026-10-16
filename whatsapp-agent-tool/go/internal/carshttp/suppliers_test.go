package carshttp

import (
	"strings"
	"testing"
)

func TestBuildSuppliersQuery(t *testing.T) {
	tests := []struct {
		name       string
		filters    SupplierFilters
		wantParams []any
		wantSQL    []string // substrings that must appear
	}{
		{
			name:       "no filter uses default limit and orders by name",
			filters:    SupplierFilters{},
			wantParams: nil,
			wantSQL:    []string{"FROM suppliers s", "ORDER BY s.name ASC", "LIMIT 50"},
		},
		{
			name:       "name is a bound LIKE",
			filters:    SupplierFilters{Name: "ALI"},
			wantParams: []any{"%ALI%"},
			wantSQL:    []string{"WHERE s.name LIKE ?", "ORDER BY s.name ASC"},
		},
		{
			name:       "limit is clamped to the max",
			filters:    SupplierFilters{Limit: 9999},
			wantParams: nil,
			wantSQL:    []string{"LIMIT 500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildSuppliersQuery(tt.filters)
			for _, want := range tt.wantSQL {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL missing %q:\n%s", want, sql)
				}
			}
			if len(params) != len(tt.wantParams) {
				t.Fatalf("params = %v, want %v", params, tt.wantParams)
			}
			for i := range tt.wantParams {
				if params[i] != tt.wantParams[i] {
					t.Errorf("param[%d] = %v, want %v", i, params[i], tt.wantParams[i])
				}
			}
		})
	}
}

// TestBuildSuppliersQueryEscapesLike guards against LIKE injection: a name with
// wildcards must be escaped the same way stock does it.
func TestBuildSuppliersQueryEscapesLike(t *testing.T) {
	_, params := BuildSuppliersQuery(SupplierFilters{Name: "50%_off\\"})
	if len(params) != 1 || params[0] != `%50\%\_off\\%` {
		t.Fatalf("params = %v, want escaped like pattern", params)
	}
}
