// Package procurement is the db layer for the ai database's supplier +
// supplier-quotation tables (see db/schema.sql), which are LINKED to the pricing
// catalog: cars.supplier_id, car_price_history.supplier_id/quotation_id and
// supplier_quotation_items.car_ref. It is unrelated to the cars-app suppliers
// (internal/carshttp).
package procurement

import (
	"fmt"
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// Supplier is one row of the ai `suppliers` table.
type Supplier struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ContactInfo any    `json:"contactInfo"`
	Notes       any    `json:"notes"`
}

// ListSuppliers returns suppliers ordered by name, optionally filtered by a
// case-insensitive name substring.
func ListSuppliers(name string, limit int) ([]Supplier, error) {
	q := "SELECT id, name, contact_info, notes FROM suppliers"
	var params []any
	if strings.TrimSpace(name) != "" {
		q += " WHERE name LIKE ?"
		params = append(params, "%"+name+"%")
	}
	q += " ORDER BY name ASC LIMIT " + strconv.Itoa(clampLimit(limit))
	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	out := make([]Supplier, 0, len(rows))
	for _, r := range rows {
		out = append(out, Supplier{
			ID:          int64(num(r["id"])),
			Name:        s(r["name"]),
			ContactInfo: r["contact_info"],
			Notes:       r["notes"],
		})
	}
	return out, nil
}

// CreateSupplier inserts a supplier. Empty contact/notes are stored NULL. The
// name is unique, so a duplicate surfaces as a driver error.
func CreateSupplier(name, contactInfo, notes string) (int64, error) {
	return db.Insert("INSERT INTO suppliers (name, contact_info, notes) VALUES (?,?,?)",
		nullIfEmpty(name), nullIfEmpty(contactInfo), nullIfEmpty(notes))
}

// UpdateSupplier changes a supplier's contact_info and/or notes. Only the fields
// with a non-empty value are updated (as the web form does for its text boxes);
// pass "strings.TrimSpace(contact) == ” && notes == ”" to clear both to NULL.
func UpdateSupplier(id int64, contact, notes string) (int64, error) {
	sets := []string{}
	var params []any
	if contact != "" {
		sets = append(sets, "contact_info = ?")
		params = append(params, contact)
	}
	if notes != "" {
		sets = append(sets, "notes = ?")
		params = append(params, notes)
	}
	if len(sets) == 0 {
		return 0, fmt.Errorf("nothing to update: pass --contact and/or --notes")
	}
	params = append(params, id)
	return db.Execute("UPDATE suppliers SET "+strings.Join(sets, ", ")+" WHERE id = ?", params...)
}

// FindSupplierID resolves a supplier by exact (case-insensitive) name. Returns
// id, found. Used by the pricing writers to attribute a price to its supplier.
func FindSupplierID(name string) (int64, bool, error) {
	rows, err := db.Query("SELECT id, name FROM suppliers WHERE LOWER(name) = LOWER(?) LIMIT 1", name)
	if err != nil {
		return 0, false, err
	}
	if len(rows) == 0 {
		return 0, false, nil
	}
	return int64(num(rows[0]["id"])), true, nil
}

// ---------------------------------------------------------------- helpers

func num(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case int:
		return float64(t)
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	}
	return 0
}

func s(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return ""
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func clampLimit(n int) int {
	if n <= 0 {
		return 100
	}
	if n > 500 {
		return 500
	}
	return n
}
