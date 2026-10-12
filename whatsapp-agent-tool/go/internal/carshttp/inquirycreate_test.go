package carshttp

import "testing"

func intPtr(n int) *int           { return &n }
func boolPtr(b bool) *bool        { return &b }
func floatPtr(f float64) *float64 { return &f }

func TestBuildCreateInquiryQuery(t *testing.T) {
	tests := []struct {
		name       string
		idSupplier *int
		date       string
		ref        string
		status     string
		notes      string
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "all empty inserts all defaults",
			wantSQL:    "INSERT INTO inquiry () VALUES ()",
			wantParams: nil,
		},
		{
			name:       "supplier only",
			idSupplier: intPtr(4),
			wantSQL:    "INSERT INTO inquiry (id_supplier) VALUES (?)",
			wantParams: []any{4},
		},
		{
			name:       "every column named in table order",
			idSupplier: intPtr(4),
			date:       "2026-10-11",
			ref:        "INQ-1",
			status:     "open",
			notes:      "hello",
			wantSQL:    "INSERT INTO inquiry (id_supplier, date_inquiry, inquiry_ref, status, notes) VALUES (?, ?, ?, ?, ?)",
			wantParams: []any{4, "2026-10-11", "INQ-1", "open", "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildCreateInquiryQuery(tt.idSupplier, tt.date, tt.ref, tt.status, tt.notes)
			if sql != tt.wantSQL {
				t.Errorf("SQL = %q, want %q", sql, tt.wantSQL)
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuildAddInquiryCarQuery(t *testing.T) {
	tests := []struct {
		name       string
		car        InquiryCar
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "required columns only",
			car:        InquiryCar{InquiryID: 2, CarNameID: 9},
			wantSQL:    "INSERT INTO inquiry_details (id_inquiry, id_car_name) VALUES (?, ?)",
			wantParams: []any{2, 9},
		},
		{
			name: "all optionals in fixed column order",
			car: InquiryCar{
				InquiryID: 2, CarNameID: 9,
				ColorID: intPtr(1), QTY: intPtr(3), Year: intPtr(2020), Month: intPtr(6),
				IsUsedCar: boolPtr(true), IsBigCar: boolPtr(false),
				TargetPrice: floatPtr(12500.5), Notes: "note",
			},
			wantSQL:    "INSERT INTO inquiry_details (id_inquiry, id_car_name, id_color, QTY, year, month, is_used_car, is_big_car, target_price, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			wantParams: []any{2, 9, 1, 3, 2020, 6, 1, 0, 12500.5, "note"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildAddInquiryCarQuery(tt.car)
			if sql != tt.wantSQL {
				t.Errorf("SQL = %q, want %q", sql, tt.wantSQL)
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuildDeleteInquiryQuery(t *testing.T) {
	sql, params := BuildDeleteInquiryQuery(5)
	if sql != "DELETE FROM inquiry WHERE id = ?" {
		t.Errorf("SQL = %q", sql)
	}
	assertParams(t, params, []any{5})
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 || boolToInt(false) != 0 {
		t.Fatalf("boolToInt: true=%d false=%d, want 1/0", boolToInt(true), boolToInt(false))
	}
}
