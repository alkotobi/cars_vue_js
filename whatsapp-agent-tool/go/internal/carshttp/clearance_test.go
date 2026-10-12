package carshttp

import (
	"strings"
	"testing"
)

func TestBuildClearanceQuery(t *testing.T) {
	tests := []struct {
		name         string
		filters      ClearanceFilters
		wantContains []string
		wantParams   []any
	}{
		{
			name:         "no filters selects everything ordered by car name",
			filters:      ClearanceFilters{},
			wantContains: []string{"FROM clearance cl", "LEFT JOIN cars_names cn", "ORDER BY cn.car_name ASC"},
			wantParams:   nil,
		},
		{
			name:         "car-name-id and id bind as equality",
			filters:      ClearanceFilters{ID: 9, CarNameID: 4},
			wantContains: []string{"cl.id = ?", "cl.id_car_name = ?"},
			wantParams:   []any{9, 4},
		},
		{
			name:         "name and brand match with like",
			filters:      ClearanceFilters{Name: "COOLRAY", Brand: "GEELY"},
			wantContains: []string{"cn.car_name LIKE ?", "b.brand LIKE ?"},
			wantParams:   []any{"%COOLRAY%", "%GEELY%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildClearanceQuery(tt.filters)
			for _, want := range tt.wantContains {
				if !strings.Contains(sql, want) {
					t.Errorf("SQL = %q, want it to contain %q", sql, want)
				}
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}

func TestBuildSetClearanceQuery(t *testing.T) {
	note := "includes port fees"
	tests := []struct {
		name       string
		idCarName  int
		dza        float64
		isBigCar   bool
		isUsedCar  bool
		notes      *string
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "without notes keeps existing note on update",
			idCarName:  7,
			dza:        950000,
			isBigCar:   false,
			isUsedCar:  false,
			notes:      nil,
			wantSQL:    "INSERT INTO clearance (id_car_name, clearance_dza, is_big_car, is_used_car) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE clearance_dza = VALUES(clearance_dza)",
			wantParams: []any{7, 950000.0, 0, 0},
		},
		{
			name:       "with notes writes and updates the note",
			idCarName:  12,
			dza:        800000.5,
			isBigCar:   true,
			isUsedCar:  true,
			notes:      &note,
			wantSQL:    "INSERT INTO clearance (id_car_name, clearance_dza, is_big_car, is_used_car, notes) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE clearance_dza = VALUES(clearance_dza), notes = VALUES(notes)",
			wantParams: []any{12, 800000.5, 1, 1, "includes port fees"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params := BuildSetClearanceQuery(tt.idCarName, tt.dza, tt.isBigCar, tt.isUsedCar, tt.notes)
			if sql != tt.wantSQL {
				t.Errorf("SQL = %q, want %q", sql, tt.wantSQL)
			}
			assertParams(t, params, tt.wantParams)
		})
	}
}
