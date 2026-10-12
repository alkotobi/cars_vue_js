// Supplier lookup, mirroring the cars app's SuppliersView. The app fetches the
// whole table (SELECT * FROM suppliers ORDER BY name ASC) behind the
// can_purchase_cars permission; here the same fixed, parameterised SELECT runs
// as the AI user, optionally narrowed by name.
package carshttp

import (
	"encoding/json"
	"strconv"
)

// SupplierFilters narrows the supplier list. The app has no filter of its own
// (it loads every row); these exist so the AI can ask a targeted question
// without pulling the whole table.
type SupplierFilters struct {
	Name  string `json:"name,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// suppliersSelect keeps the app's column set (id/name/contact_info/notes) as an
// explicit list rather than `*`, so the payload stays stable if the table gains
// columns the view does not show.
const suppliersSelect = `
  SELECT
    s.id,
    s.name,
    s.contact_info,
    s.notes
  FROM suppliers s
`

// BuildSuppliersQuery mirrors the app's ordering (name ASC) with an optional
// name match. The limit is clamped the same way stock is.
func BuildSuppliersQuery(f SupplierFilters) (string, []any) {
	var params []any
	where := ""
	if f.Name != "" {
		where = "WHERE s.name LIKE ?"
		params = append(params, Like(f.Name))
	}
	sql := suppliersSelect + where + "\n    ORDER BY s.name ASC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchSuppliers runs the fixed supplier query as the AI user.
func SearchSuppliers(f SupplierFilters) ([]json.RawMessage, error) {
	sql, params := BuildSuppliersQuery(f)
	return QueryAsAI(sql, params)
}