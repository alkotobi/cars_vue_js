// Brand creation, mirroring useBrands.js addBrand(). The app INSERTs into
// brands over the query passthrough, with logo_path only when a logo was
// uploaded first. This is the tool's one deliberate write to the cars app, so
// it is a separate, explicitly-named file rather than folded into reads.
package carshttp

// BuildCreateBrandQuery returns the fixed INSERT for a new brand. The column
// list is chosen by whether a logo path was provided, never built from input;
// both values are bound parameters.
func BuildCreateBrandQuery(brand, logoPath string) (string, []any) {
	if logoPath != "" {
		return "INSERT INTO brands (brand, logo_path) VALUES (?, ?)", []any{brand, logoPath}
	}
	return "INSERT INTO brands (brand) VALUES (?)", []any{brand}
}

// CreateBrand inserts a brand and returns its new id. logoPath may be empty,
// in which case the row carries no logo (the column is left to its default).
func CreateBrand(brand, logoPath string) (int64, error) {
	sql, params := BuildCreateBrandQuery(brand, logoPath)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.LastInsertID, nil
}
