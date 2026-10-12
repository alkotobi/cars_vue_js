// Car-name (model) creation, mirroring useCarNames.js addCarName(). The app
// INSERTs into cars_names over the query passthrough. This is a deliberate write
// to the cars app (the AI user's role must grant can_purchase_cars), so it lives
// in its own file rather than folded into reads.
package carshttp

import "strings"

// BuildCreateCarNameQuery returns the fixed INSERT for a new car name. It
// mirrors the app exactly: the name is trimmed and upcased, notes defaults to
// the empty string the form sends, and is_big_car is stored as 0/1.
func BuildCreateCarNameQuery(carName string, idBrand int, isBigCar bool, notes string) (string, []any) {
	return "INSERT INTO cars_names (car_name, notes, is_big_car, id_brand) VALUES (?, ?, ?, ?)",
		[]any{strings.ToUpper(strings.TrimSpace(carName)), notes, boolToInt(isBigCar), idBrand}
}

// CreateCarName inserts a car name and returns its new id. id_brand must be an
// existing brand (cars_names.id_brand references brands.id).
func CreateCarName(carName string, idBrand int, isBigCar bool, notes string) (int64, error) {
	sql, params := BuildCreateCarNameQuery(carName, idBrand, isBigCar, notes)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.LastInsertID, nil
}
