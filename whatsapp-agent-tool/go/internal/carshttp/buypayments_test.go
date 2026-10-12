package carshttp

import (
	"strings"
	"testing"
)

func TestBuildBuyPaymentsQuery(t *testing.T) {
	tests := []struct {
		name         string
		filters      BuyPaymentFilters
		wantContains []string
		wantParams   []any
	}{
		{
			name:         "no filters selects all newest first",
			filters:      BuyPaymentFilters{},
			wantContains: []string{"FROM buy_payments bp", "LEFT JOIN buy_bill bb", "LEFT JOIN suppliers s", "ORDER BY bp.date_payment DESC"},
			wantParams:   nil,
		},
		{
			name:         "bill id binds and scopes to one bill",
			filters:      BuyPaymentFilters{BillID: 12},
			wantContains: []string{"bp.id_buy_bill = ?"},
			wantParams:   []any{12},
		},
		{
			name:         "supplier id scopes through the bill join",
			filters:      BuyPaymentFilters{SupplierID: 3},
			wantContains: []string{"bb.id_supplier = ?"},
			wantParams:   []any{3},
		},
		{
			name:         "single payment id",
			filters:      BuyPaymentFilters{ID: 7},
			wantContains: []string{"bp.id = ?"},
			wantParams:   []any{7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildBuyPaymentsQuery(tt.filters)
			for _, want := range tt.wantContains {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL = %q, want it to contain %q", sql, want)
				}
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuyBillHeaderQuery(t *testing.T) {
	for _, want := range []string{
		"FROM buy_bill bb",
		"SUM(amount * QTY)",
		"SUM(amount)",
		"calculated_amount",
		"total_paid",
		"WHERE bb.id = ?",
	} {
		if !strings.Contains(BuyBillHeaderQuery, want) {
			t.Errorf("BuyBillHeaderQuery missing %q", want)
		}
	}
}
