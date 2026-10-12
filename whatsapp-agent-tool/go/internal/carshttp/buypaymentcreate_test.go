package carshttp

import "testing"

func TestBuildCreateBuyPaymentQuery(t *testing.T) {
	userID := 7
	tests := []struct {
		name       string
		billID     int
		date       string
		amount     float64
		swiftPath  string
		notes      string
		userID     *int
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "with user id stamps id_user",
			billID:     1,
			date:       "2026-10-07 11:13:00",
			amount:     23000,
			swiftPath:  "buy_payment_swifts/bill_1_123.pdf",
			notes:      "",
			userID:     &userID,
			wantSQL:    "INSERT INTO buy_payments (id_buy_bill, date_payment, amount, swift_path, notes, id_user) VALUES (?, ?, ?, ?, ?, ?)",
			wantParams: []any{1, "2026-10-07 11:13:00", 23000.0, "buy_payment_swifts/bill_1_123.pdf", "", 7},
		},
		{
			name:       "without user id omits the column",
			billID:     1,
			amount:     500,
			swiftPath:  "s.pdf",
			date:       "2026-10-07 11:13:00",
			userID:     nil,
			wantSQL:    "INSERT INTO buy_payments (id_buy_bill, date_payment, amount, swift_path, notes) VALUES (?, ?, ?, ?, ?)",
			wantParams: []any{1, "2026-10-07 11:13:00", 500.0, "s.pdf", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildCreateBuyPaymentQuery(tt.billID, tt.date, tt.amount, tt.swiftPath, tt.notes, tt.userID)
			if sql != tt.wantSQL {
				t.Errorf("SQL = %q, want %q", sql, tt.wantSQL)
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuildResyncBillPaidQuery(t *testing.T) {
	sql, params := BuildResyncBillPaidQuery(5)
	const wantSQL = "UPDATE buy_bill SET payed = (SELECT COALESCE(SUM(amount), 0) FROM buy_payments WHERE id_buy_bill = ?) WHERE id = ?"
	if sql != wantSQL {
		t.Errorf("SQL = %q, want %q", sql, wantSQL)
	}
	assertParams(t, params, []any{5, 5})
}

func TestBuildUpdateBuyPaymentQuery(t *testing.T) {
	sql, params := BuildUpdateBuyPaymentQuery(2, 77000, "2026-10-11 12:00:00", "buy_payment_swifts/bill_1_9.pdf", "kept note")
	const wantSQL = "UPDATE buy_payments SET amount = ?, date_payment = ?, swift_path = ?, notes = ? WHERE id = ?"
	if sql != wantSQL {
		t.Errorf("SQL = %q, want %q", sql, wantSQL)
	}
	assertParams(t, params, []any{77000.0, "2026-10-11 12:00:00", "buy_payment_swifts/bill_1_9.pdf", "kept note", 2})
}

func TestBuildDeleteBuyPaymentQuery(t *testing.T) {
	sql, params := BuildDeleteBuyPaymentQuery(2)
	if sql != "DELETE FROM buy_payments WHERE id = ?" {
		t.Errorf("SQL = %q", sql)
	}
	assertParams(t, params, []any{2})
}
