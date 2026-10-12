// Car-name lookup, mirroring the cars app's CarNamesView (useCarNames.js). The
// app loads the whole cars_names table (joined to brands) behind the models tab
// (can_purchase_cars); the same fixed SELECT runs here as the AI user, optionally
// narrowed by car name and/or brand.
package carshttp

import (
	"encoding/json"
	"strconv"
	"strings"
)

// CarNameFilters narrows the car-name list. The app filters client-side
// (carNamesFilter.js); these map to the same columns server-side.
type CarNameFilters struct {
	Name  string `json:"name,omitempty"`
	Brand string `json:"brand,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// carNamesSelect mirrors useCarNames.js' columns and its brands LEFT JOIN.
const carNamesSelect = `
  SELECT
    cn.id,
    cn.car_name,
    cn.notes,
    cn.is_big_car,
    cn.id_brand,
    b.brand
  FROM cars_names cn
  LEFT JOIN brands b ON cn.id_brand = b.id
`

// BuildCarNamesQuery orders by car_name ASC like the app, with optional name and
// brand matches.
func BuildCarNamesQuery(f CarNameFilters) (string, []any) {
	var conditions []string
	var params []any
	if f.Name != "" {
		conditions = append(conditions, "cn.car_name LIKE ?")
		params = append(params, Like(f.Name))
	}
	if f.Brand != "" {
		conditions = append(conditions, "b.brand LIKE ?")
		params = append(params, Like(f.Brand))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	sql := carNamesSelect + where + "\n    ORDER BY cn.car_name ASC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchCarNames runs the fixed car-names query as the AI user.
func SearchCarNames(f CarNameFilters) ([]json.RawMessage, error) {
	sql, params := BuildCarNamesQuery(f)
	return QueryAsAI(sql, params)
}
