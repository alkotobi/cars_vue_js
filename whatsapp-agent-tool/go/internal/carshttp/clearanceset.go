// Clearance write support: set (upsert) a model's customs clearance in the cars
// app, table clearance (migration 036). This is a deliberate write (the AI user's
// role must grant can_c_cfr_prices), so it lives in its own file rather than
// folded into reads. The (id_car_name, is_big_car, is_used_car) unique key makes
// one row per model/variant, so setting an existing key updates it in place.
package carshttp

import "strings"

// BuildSetClearanceQuery returns the fixed upsert for one clearance row: INSERT a
// new row, or ON DUPLICATE KEY UPDATE its amount (and notes, when given) when the
// unique key already exists. The row's key columns and amount are always written;
// notes is optional — omitting it leaves any existing note untouched on update,
// and stores NULL on insert (never an empty string).
func BuildSetClearanceQuery(idCarName int, clearanceDza float64, isBigCar, isUsedCar bool, notes *string) (string, []any) {
	cols := []string{"id_car_name", "clearance_dza", "is_big_car", "is_used_car"}
	params := []any{idCarName, clearanceDza, boolToInt(isBigCar), boolToInt(isUsedCar)}
	if notes != nil {
		cols = append(cols, "notes")
		params = append(params, *notes)
	}
	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	updates := []string{"clearance_dza = VALUES(clearance_dza)"}
	if notes != nil {
		updates = append(updates, "notes = VALUES(notes)")
	}
	sql := "INSERT INTO clearance (" + strings.Join(cols, ", ") + ") VALUES (" +
		strings.Join(placeholders, ", ") + ") ON DUPLICATE KEY UPDATE " +
		strings.Join(updates, ", ")
	return sql, params
}

// SetClearance upserts one clearance row and returns the write result. The
// affected-rows value is 1 for a fresh insert, 2 when an existing row changed, and
// 0 when the update was a no-op (MySQL's ON DUPLICATE KEY semantics).
func SetClearance(idCarName int, clearanceDza float64, isBigCar, isUsedCar bool, notes *string) (*WriteResult, error) {
	sql, params := BuildSetClearanceQuery(idCarName, clearanceDza, isBigCar, isUsedCar, notes)
	return ExecAsAI(sql, params)
}
