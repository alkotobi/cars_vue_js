// Brand lookup, mirroring the cars app's BrandsModal (useBrands.js). The app
// loads the whole brands table behind the models tab (can_purchase_cars); the
// same fixed SELECT runs here as the AI user, optionally narrowed by brand name.
package carshttp

import (
	"encoding/json"
	"strconv"
)

// BrandFilters narrows the brand list. The app has no server filter (it loads
// every row); this exists so the AI can ask a targeted question.
type BrandFilters struct {
	Brand string `json:"brand,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// brandsSelect mirrors useBrands.js' columns.
const brandsSelect = `
  SELECT
    b.id,
    b.brand,
    b.logo_path
  FROM brands b
`

// BuildBrandsQuery orders by brand ASC like the app, with an optional name match.
func BuildBrandsQuery(f BrandFilters) (string, []any) {
	var params []any
	where := ""
	if f.Brand != "" {
		where = "WHERE b.brand LIKE ?"
		params = append(params, Like(f.Brand))
	}
	sql := brandsSelect + where + "\n    ORDER BY b.brand ASC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchBrands runs the fixed brands query as the AI user.
func SearchBrands(f BrandFilters) ([]json.RawMessage, error) {
	sql, params := BuildBrandsQuery(f)
	return QueryAsAI(sql, params)
}
