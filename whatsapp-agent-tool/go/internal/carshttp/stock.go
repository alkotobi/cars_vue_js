// Package carshttp also implements read-only stock lookup with a fixed SELECT
// (never free-form SQL from the caller). The query mirrors the cars app's own
// CarStockTable view so the AI sees exactly what a logged-in user sees.
package carshttp

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Filters struct {
	Car           string `json:"car,omitempty"`
	Color         string `json:"color,omitempty"`
	Vin           string `json:"vin,omitempty"`
	Port          string `json:"port,omitempty"`
	Status        string `json:"status"`
	Limit         int    `json:"limit,omitempty"`
	IncludeHidden bool   `json:"includeHidden"`
}

const defaultLimit = 50
const maxLimit = 500

// Like escapes LIKE metacharacters, mirroring stock.ts like().
func Like(term string) string {
	var b strings.Builder
	for _, ch := range term {
		switch ch {
		case '\\', '%', '_':
			b.WriteByte('\\')
		}
		b.WriteRune(ch)
	}
	return "%" + b.String() + "%"
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

const stockSelect = `
  SELECT
    cs.id,
    cs.vin,
    cs.car_name,
    cs.color,
    cs.hexa,
    cs.status,
    cs.loading_port,
    cs.discharge_port,
    cs.warehouse_name,
    cs.container_ref,
    cs.date_loding,
    cs.export_lisence_ref,
    cs.buy_bill_ref,
    cs.sell_bill_ref,
    cs.client_name,
    cs.is_big_car,
    cs.is_used_car,
    cs.hidden,
    cs.price_cell,
    cs.freight,
    cs.rate,
    cs.cfr_da,
    cs.buy_price
  FROM (
    SELECT
      cs.id, cs.vin, cs.price_cell, cs.freight, cs.rate, cs.cfr_da,
      cs.date_loding, cs.container_ref, cs.export_lisence_ref, cs.hidden,
      cs.is_big_car, cs.is_used_car,
      CASE WHEN cs.id_sell IS NOT NULL THEN 'Sold' ELSE 'Available' END AS status,
      c.name AS client_name,
      cn.car_name, clr.color, clr.hexa,
      lp.loading_port, dp.discharge_port, w.warhouse_name AS warehouse_name,
      bb.bill_ref AS buy_bill_ref, sb.bill_ref AS sell_bill_ref,
      bd.amount AS buy_price
    FROM cars_stock cs
    LEFT JOIN clients c ON cs.id_client = c.id
    LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
    LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
    LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
    LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
    LEFT JOIN colors clr ON cs.id_color = clr.id
    LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
    LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
    LEFT JOIN warehouses w ON cs.id_warehouse = w.id
  ) cs
`

// BuildStockQuery mirrors buildStockQuery().
func BuildStockQuery(f Filters) (string, []any) {
	var conditions []string
	var params []any

	if !f.IncludeHidden {
		conditions = append(conditions, "cs.hidden = 0")
	}
	if f.Car != "" {
		conditions = append(conditions, "cs.car_name LIKE ?")
		params = append(params, Like(f.Car))
	}
	if f.Color != "" {
		conditions = append(conditions, "cs.color LIKE ?")
		params = append(params, Like(f.Color))
	}
	if f.Vin != "" {
		conditions = append(conditions, "cs.vin LIKE ?")
		params = append(params, Like(f.Vin))
	}
	if f.Port != "" {
		conditions = append(conditions, "(cs.loading_port LIKE ? OR cs.discharge_port LIKE ?)")
		params = append(params, Like(f.Port), Like(f.Port))
	}
	if f.Status == "available" {
		conditions = append(conditions, "cs.status = 'Available'")
	} else if f.Status == "sold" {
		conditions = append(conditions, "cs.status = 'Sold'")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	sql := stockSelect + where + "\n    ORDER BY cs.status = 'Available' DESC, cs.id DESC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchStock mirrors searchStock().
func SearchStock(f Filters) ([]json.RawMessage, error) {
	sql, params := BuildStockQuery(f)
	return QueryAsAI(sql, params)
}
