// Supplier creation, mirroring SuppliersView.vue addSupplier(). The app INSERTs
// into suppliers over the query passthrough with just name/contact_info/notes.
// This is a deliberate write to the cars app (the AI user's role must grant
// can_c_suppliers), so it lives in its own file rather than folded into reads.
package carshttp

// BuildCreateSupplierQuery returns the fixed INSERT for a new supplier. An empty
// contact_info or notes is sent as NULL rather than ” — SuppliersView.vue says
// the same ("nothing on record" stays distinguishable from "a blank space").
func BuildCreateSupplierQuery(name, contactInfo, notes string) (string, []any) {
	return "INSERT INTO suppliers (name, contact_info, notes) VALUES (?, ?, ?)",
		[]any{name, nullIfEmpty(contactInfo), nullIfEmpty(notes)}
}

// CreateSupplier inserts a supplier and returns its new id. The name is unique
// (supplier_name_unic), so a duplicate surfaces as a query_failed error.
func CreateSupplier(name, contactInfo, notes string) (int64, error) {
	sql, params := BuildCreateSupplierQuery(name, contactInfo, notes)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.LastInsertID, nil
}

// nullIfEmpty returns nil for "" so the column is stored as NULL.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
