// Inquiry read support, mirroring the cars app's BuyView.vue. The app has no
// inquiry screen yet; these tables (inquiry / inquiry_details, migration 034)
// are the purchase-side "supplier quote request" system shaped exactly like
// buy_bill / buy_details. Reads run as the AI user behind can_purchase_cars —
// the same permission the models/buy tabs use — through the query passthrough.
package carshttp

import (
	"encoding/json"
	"strconv"
	"strings"
)

// InquiryFilters narrows the inquiry (header) list. Every field is optional;
// the app loads the whole table, these map to the same columns server-side.
type InquiryFilters struct {
	ID         int    `json:"id,omitempty"`
	SupplierID int    `json:"supplierId,omitempty"`
	Status     string `json:"status,omitempty"`
	Ref        string `json:"ref,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// inquiriesSelect mirrors BuyView's header list columns (LEFT JOIN suppliers).
const inquiriesSelect = `
  SELECT
    i.id,
    i.id_supplier,
    s.name AS supplier_name,
    i.date_inquiry,
    i.inquiry_ref,
    i.amount,
    i.status,
    i.notes
  FROM inquiry i
  LEFT JOIN suppliers s ON i.id_supplier = s.id
`

// BuildInquiriesQuery orders newest first (like the buy list) with optional
// id / supplier / status / ref filters. Every value is bound.
func BuildInquiriesQuery(f InquiryFilters) (string, []any) {
	var conditions []string
	var params []any
	if f.ID > 0 {
		conditions = append(conditions, "i.id = ?")
		params = append(params, f.ID)
	}
	if f.SupplierID > 0 {
		conditions = append(conditions, "i.id_supplier = ?")
		params = append(params, f.SupplierID)
	}
	if f.Status != "" {
		conditions = append(conditions, "i.status = ?")
		params = append(params, f.Status)
	}
	if f.Ref != "" {
		conditions = append(conditions, "i.inquiry_ref LIKE ?")
		params = append(params, Like(f.Ref))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	sql := inquiriesSelect + where + "\n    ORDER BY i.date_inquiry DESC, i.id DESC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchInquiries runs the fixed inquiry query as the AI user.
func SearchInquiries(f InquiryFilters) ([]json.RawMessage, error) {
	sql, params := BuildInquiriesQuery(f)
	return QueryAsAI(sql, params)
}

// InquiryCarFilters narrows the car lines of one inquiry.
type InquiryCarFilters struct {
	InquiryID int `json:"inquiryId,omitempty"`
	Limit     int `json:"limit,omitempty"`
}

// inquiryCarsSelect mirrors BuyView's detail join (cars_names + colors), plus
// the inquiry link.
const inquiryCarsSelect = `
  SELECT
    d.id,
    d.id_inquiry,
    d.id_car_name,
    cn.car_name,
    d.id_color,
    c.color,
    d.QTY,
    d.year,
    d.month,
    d.is_used_car,
    d.is_big_car,
    d.target_price,
    d.notes
  FROM inquiry_details d
  LEFT JOIN cars_names cn ON d.id_car_name = cn.id
  LEFT JOIN colors c ON d.id_color = c.id
`

// BuildInquiryCarsQuery lists one inquiry's lines (by id ASC, insertion order).
func BuildInquiryCarsQuery(f InquiryCarFilters) (string, []any) {
	var params []any
	where := ""
	if f.InquiryID > 0 {
		where = "WHERE d.id_inquiry = ?"
		params = append(params, f.InquiryID)
	}
	sql := inquiryCarsSelect + where + "\n    ORDER BY d.id ASC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchInquiryCars runs the fixed inquiry-cars query as the AI user.
func SearchInquiryCars(f InquiryCarFilters) ([]json.RawMessage, error) {
	sql, params := BuildInquiryCarsQuery(f)
	return QueryAsAI(sql, params)
}
