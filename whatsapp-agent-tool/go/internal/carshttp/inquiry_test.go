package carshttp

import (
	"strings"
	"testing"
)

func TestBuildInquiriesQuery(t *testing.T) {
	tests := []struct {
		name       string
		filters    InquiryFilters
		wantParams []any
		wantSQL    []string
	}{
		{
			name:       "no filter uses default limit and newest-first order",
			filters:    InquiryFilters{},
			wantParams: nil,
			wantSQL:    []string{"FROM inquiry i", "LEFT JOIN suppliers s ON i.id_supplier = s.id", "ORDER BY i.date_inquiry DESC, i.id DESC", "LIMIT 50"},
		},
		{
			name:       "id is a bound equality",
			filters:    InquiryFilters{ID: 7},
			wantParams: []any{7},
			wantSQL:    []string{"WHERE i.id = ?"},
		},
		{
			name:       "supplier and status join with AND",
			filters:    InquiryFilters{SupplierID: 3, Status: "open"},
			wantParams: []any{3, "open"},
			wantSQL:    []string{"WHERE i.id_supplier = ? AND i.status = ?"},
		},
		{
			name:       "ref is a bound LIKE",
			filters:    InquiryFilters{Ref: "INV-1"},
			wantParams: []any{"%INV-1%"},
			wantSQL:    []string{"WHERE i.inquiry_ref LIKE ?"},
		},
		{
			name:       "limit is clamped to the max",
			filters:    InquiryFilters{Limit: 9999},
			wantParams: nil,
			wantSQL:    []string{"LIMIT 500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildInquiriesQuery(tt.filters)
			for _, want := range tt.wantSQL {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL missing %q:\n%s", want, sql)
				}
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuildInquiryCarsQuery(t *testing.T) {
	tests := []struct {
		name       string
		filters    InquiryCarFilters
		wantParams []any
		wantSQL    []string
	}{
		{
			name:       "inquiry id filters and orders by insertion",
			filters:    InquiryCarFilters{InquiryID: 12},
			wantParams: []any{12},
			wantSQL:    []string{"FROM inquiry_details d", "LEFT JOIN cars_names cn ON d.id_car_name = cn.id", "LEFT JOIN colors c ON d.id_color = c.id", "WHERE d.id_inquiry = ?", "ORDER BY d.id ASC"},
		},
		{
			name:       "limit is clamped to the max",
			filters:    InquiryCarFilters{Limit: 9999},
			wantParams: nil,
			wantSQL:    []string{"LIMIT 500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildInquiryCarsQuery(tt.filters)
			for _, want := range tt.wantSQL {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL missing %q:\n%s", want, sql)
				}
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

// TestBuildInquiriesQueryEscapesLike guards against LIKE injection on the ref.
func TestBuildInquiriesQueryEscapesLike(t *testing.T) {
	_, params := BuildInquiriesQuery(InquiryFilters{Ref: "50%_off\\"})
	if len(params) != 1 || params[0] != `%50\%\_off\\%` {
		t.Fatalf("params = %v, want escaped like pattern", params)
	}
}
