// Clearance read support: per-car customs clearance (DZD, TVN excluded), table
// clearance (migration 036), joined to its model and brand. Reads run as the AI
// user behind can_c_cfr_prices — clearance is the customs part of the CFR/landed
// price — through the same query passthrough as the other cars-app reads.
package carshttp

import (
	"encoding/json"
	"strconv"
	"strings"
)

// ClearanceFilters narrows the clearance list. Every field is optional; the app
// has no clearance screen, so these mirror how the other tables are queried.
type ClearanceFilters struct {
	ID        int    `json:"id,omitempty"`
	CarNameID int    `json:"carNameId,omitempty"`
	Name      string `json:"name,omitempty"`
	Brand     string `json:"brand,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// clearanceSelect mirrors the table's columns plus its model/brand join, so a
// caller can read a clearance without a second lookup for the car name.
const clearanceSelect = `
  SELECT
    cl.id,
    cl.id_car_name,
    cn.car_name,
    cn.id_brand,
    b.brand,
    cl.clearance_dza,
    cl.is_big_car,
    cl.is_used_car,
    cl.notes,
    cl.created_at
  FROM clearance cl
  LEFT JOIN cars_names cn ON cl.id_car_name = cn.id
  LEFT JOIN brands b ON cn.id_brand = b.id
`

// BuildClearanceQuery lists clearance rows ordered by car name (then used/big so
// the variants of one model sit together), with optional id / car-name-id / name
// / brand matches.
func BuildClearanceQuery(f ClearanceFilters) (string, []any) {
	var conditions []string
	var params []any
	if f.ID > 0 {
		conditions = append(conditions, "cl.id = ?")
		params = append(params, f.ID)
	}
	if f.CarNameID > 0 {
		conditions = append(conditions, "cl.id_car_name = ?")
		params = append(params, f.CarNameID)
	}
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
	sql := clearanceSelect + where +
		"\n    ORDER BY cn.car_name ASC, cl.is_used_car ASC, cl.is_big_car ASC\n    LIMIT " +
		strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchClearance runs the fixed clearance query as the AI user.
func SearchClearance(f ClearanceFilters) ([]json.RawMessage, error) {
	sql, params := BuildClearanceQuery(f)
	return QueryAsAI(sql, params)
}
