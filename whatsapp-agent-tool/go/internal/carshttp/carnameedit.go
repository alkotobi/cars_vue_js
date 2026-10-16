// Car-name edit: UPDATE the notes field in cars_names. This is a deliberate write
// to the cars app (the AI user's role must grant can_purchase_cars).
package carshttp

import "encoding/json"

// UpdateCarNameNotes runs UPDATE cars_names SET notes = ? WHERE id = ?. It returns
// the affected rows (0 = id not found, 1 = updated).
func UpdateCarNameNotes(id int, notes string) (int64, error) {
	result, err := ExecAsAI("UPDATE cars_names SET notes = ? WHERE id = ?", []any{notes, id})
	if err != nil {
		return 0, err
	}
	return result.AffectedRows, nil
}

// GetCarName returns one car-name row by id (for verification after edit).
func GetCarName(id int) (json.RawMessage, error) {
	rows, err := QueryAsAI("SELECT cn.id, cn.car_name, cn.notes, cn.is_big_car, cn.id_brand, b.brand FROM cars_names cn LEFT JOIN brands b ON cn.id_brand = b.id WHERE cn.id = ?", []any{id})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}
