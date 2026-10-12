package carshttp

import "testing"

func TestBuildCreateCarNameQuery(t *testing.T) {
	tests := []struct {
		name      string
		carName   string
		idBrand   int
		isBigCar  bool
		notes     string
		wantParam []any
	}{
		{
			name:      "name is trimmed and upcased, big flag as 0",
			carName:   "  gs3 mid ",
			idBrand:   14,
			isBigCar:  false,
			notes:     "",
			wantParam: []any{"GS3 MID", "", 0, 14},
		},
		{
			name:      "big car as 1 and notes passed through",
			carName:   "gs3 max",
			idBrand:   14,
			isBigCar:  true,
			notes:     "flagship",
			wantParam: []any{"GS3 MAX", "flagship", 1, 14},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildCreateCarNameQuery(tt.carName, tt.idBrand, tt.isBigCar, tt.notes)
			const wantSQL = "INSERT INTO cars_names (car_name, notes, is_big_car, id_brand) VALUES (?, ?, ?, ?)"
			if sql != wantSQL {
				t.Errorf("SQL = %q, want %q", sql, wantSQL)
			}
			assertParams(t, params, tt.wantParam)
		})
	}
}
