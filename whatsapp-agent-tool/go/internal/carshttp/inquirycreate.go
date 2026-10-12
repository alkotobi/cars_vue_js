// Inquiry write support: create a header, attach car lines, delete. Shaped
// exactly like the app's BuyView.vue writes (INSERT INTO buy_bill / buy_details)
// but against inquiry / inquiry_details (migration 034). Deleting a header
// removes its lines through the FK's ON DELETE CASCADE, so one DELETE statement
// suffices over the single-statement query passthrough.
package carshttp

import "strings"

// insertQuery builds "INSERT INTO <table> (cols) VALUES (?, ...)" with one
// placeholder per column, or the all-defaults form when there are no columns.
// table and cols are chosen by this package's code from fixed literals, never
// from input; every value is a bound parameter.
func insertQuery(table string, cols []string) string {
	if len(cols) == 0 {
		return "INSERT INTO " + table + " () VALUES ()"
	}
	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return "INSERT INTO " + table + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
}

// boolToInt maps a Go bool to the 0/1 the tables' tinyint(1) columns use.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// BuildCreateInquiryQuery returns the fixed INSERT for a new inquiry header.
// Only the columns with a value are named (the rest fall to their defaults); no
// column list is ever built from input strings.
func BuildCreateInquiryQuery(idSupplier *int, dateInquiry, ref, status, notes string) (string, []any) {
	var cols []string
	var params []any
	if idSupplier != nil {
		cols = append(cols, "id_supplier")
		params = append(params, *idSupplier)
	}
	if dateInquiry != "" {
		cols = append(cols, "date_inquiry")
		params = append(params, dateInquiry)
	}
	if ref != "" {
		cols = append(cols, "inquiry_ref")
		params = append(params, ref)
	}
	if status != "" {
		cols = append(cols, "status")
		params = append(params, status)
	}
	if notes != "" {
		cols = append(cols, "notes")
		params = append(params, notes)
	}
	return insertQuery("inquiry", cols), params
}

// CreateInquiry inserts an inquiry header and returns its new id.
func CreateInquiry(idSupplier *int, dateInquiry, ref, status, notes string) (int64, error) {
	sql, params := BuildCreateInquiryQuery(idSupplier, dateInquiry, ref, status, notes)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.LastInsertID, nil
}

// InquiryCar is one car line to attach to an inquiry. InquiryID and CarNameID
// are required; every other field is optional (nil / "" leaves the column to its
// default), mirroring the buy-details form.
type InquiryCar struct {
	InquiryID   int
	CarNameID   int
	ColorID     *int
	QTY         *int
	Year        *int
	Month       *int
	IsUsedCar   *bool
	IsBigCar    *bool
	TargetPrice *float64
	Notes       string
}

// BuildAddInquiryCarQuery returns the fixed INSERT for one inquiry car line.
func BuildAddInquiryCarQuery(c InquiryCar) (string, []any) {
	cols := []string{"id_inquiry", "id_car_name"}
	params := []any{c.InquiryID, c.CarNameID}
	if c.ColorID != nil {
		cols = append(cols, "id_color")
		params = append(params, *c.ColorID)
	}
	if c.QTY != nil {
		cols = append(cols, "QTY")
		params = append(params, *c.QTY)
	}
	if c.Year != nil {
		cols = append(cols, "year")
		params = append(params, *c.Year)
	}
	if c.Month != nil {
		cols = append(cols, "month")
		params = append(params, *c.Month)
	}
	if c.IsUsedCar != nil {
		cols = append(cols, "is_used_car")
		params = append(params, boolToInt(*c.IsUsedCar))
	}
	if c.IsBigCar != nil {
		cols = append(cols, "is_big_car")
		params = append(params, boolToInt(*c.IsBigCar))
	}
	if c.TargetPrice != nil {
		cols = append(cols, "target_price")
		params = append(params, *c.TargetPrice)
	}
	if c.Notes != "" {
		cols = append(cols, "notes")
		params = append(params, c.Notes)
	}
	return insertQuery("inquiry_details", cols), params
}

// AddInquiryCar inserts one car line and returns its new id.
func AddInquiryCar(c InquiryCar) (int64, error) {
	sql, params := BuildAddInquiryCarQuery(c)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.LastInsertID, nil
}

// BuildDeleteInquiryQuery returns the fixed DELETE for one inquiry header. Its
// lines go with it via ON DELETE CASCADE.
func BuildDeleteInquiryQuery(id int) (string, []any) {
	return "DELETE FROM inquiry WHERE id = ?", []any{id}
}

// DeleteInquiry removes an inquiry (and, by cascade, its lines); returns the
// number of header rows removed (0 when the id did not exist).
func DeleteInquiry(id int) (int64, error) {
	sql, params := BuildDeleteInquiryQuery(id)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.AffectedRows, nil
}
