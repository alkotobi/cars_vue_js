package carshttp

import (
	"strings"
	"testing"
)

func TestBuildBrandsQuery(t *testing.T) {
	tests := []struct {
		name       string
		filters    BrandFilters
		wantParams []any
		wantSQL    []string
	}{
		{
			name:       "no filter uses default limit and orders by brand",
			filters:    BrandFilters{},
			wantParams: nil,
			wantSQL:    []string{"FROM brands b", "ORDER BY b.brand ASC", "LIMIT 50"},
		},
		{
			name:       "brand is a bound LIKE",
			filters:    BrandFilters{Brand: "TOYOTA"},
			wantParams: []any{"%TOYOTA%"},
			wantSQL:    []string{"WHERE b.brand LIKE ?", "ORDER BY b.brand ASC"},
		},
		{
			name:       "limit is clamped to the max",
			filters:    BrandFilters{Limit: 9999},
			wantParams: nil,
			wantSQL:    []string{"LIMIT 500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildBrandsQuery(tt.filters)
			for _, want := range tt.wantSQL {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL missing %q:\n%s", want, sql)
				}
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuildCarNamesQuery(t *testing.T) {
	tests := []struct {
		name       string
		filters    CarNameFilters
		wantParams []any
		wantSQL    []string
	}{
		{
			name:       "no filter uses default limit and joins brands",
			filters:    CarNameFilters{},
			wantParams: nil,
			wantSQL:    []string{"FROM cars_names cn", "LEFT JOIN brands b ON cn.id_brand = b.id", "ORDER BY cn.car_name ASC", "LIMIT 50"},
		},
		{
			name:       "name is a bound LIKE",
			filters:    CarNameFilters{Name: "GOLF"},
			wantParams: []any{"%GOLF%"},
			wantSQL:    []string{"WHERE cn.car_name LIKE ?", "ORDER BY cn.car_name ASC"},
		},
		{
			name:       "brand joins the where with AND",
			filters:    CarNameFilters{Name: "GOLF", Brand: "VW"},
			wantParams: []any{"%GOLF%", "%VW%"},
			wantSQL:    []string{"WHERE cn.car_name LIKE ? AND b.brand LIKE ?"},
		},
		{
			name:       "limit is clamped to the max",
			filters:    CarNameFilters{Limit: 9999},
			wantParams: nil,
			wantSQL:    []string{"LIMIT 500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildCarNamesQuery(tt.filters)
			for _, want := range tt.wantSQL {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL missing %q:\n%s", want, sql)
				}
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBrandQueriesEscapeLike(t *testing.T) {
	_, brandParams := BuildBrandsQuery(BrandFilters{Brand: "50%off"})
	if len(brandParams) != 1 || brandParams[0] != `%50\%off%` {
		t.Fatalf("brand params = %v, want escaped like pattern", brandParams)
	}
	_, cnParams := BuildCarNamesQuery(CarNameFilters{Name: "a_b", Brand: `c\d`})
	if len(cnParams) != 2 || cnParams[0] != `%a\_b%` || cnParams[1] != `%c\\d%` {
		t.Fatalf("car-name params = %v, want escaped like patterns", cnParams)
	}
}

func assertParams(t *testing.T, got, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("params = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("param[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}