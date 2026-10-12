package procurement

import (
	"strconv"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/db"
)

// QuotationStatuses are the allowed supplier_quotations.status values (the DB
// ENUM, kept here so the CLI can reject a bad value before it reaches MySQL).
var QuotationStatuses = []string{"requested", "received", "accepted", "rejected", "expired"}

// ValidStatus reports whether s is one of the allowed status values.
func ValidStatus(s string) bool {
	for _, v := range QuotationStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// ItemPrice is one colour price of a quotation item.
type ItemPrice struct {
	Color     string  `json:"color"`
	UnitPrice float64 `json:"unitPrice"`
	Notes     any     `json:"notes"`
}

// Item is one car line of a quotation.
type Item struct {
	ID        int64       `json:"id"`
	CarRef    any         `json:"carRef"` // cars.id — null until linked to the catalog
	CarName   string      `json:"carName"`
	Size      string      `json:"size"`
	Condition string      `json:"condition"`
	Quantity  int         `json:"quantity"`
	Year      any         `json:"year"`
	Month     any         `json:"month"`
	Notes     any         `json:"notes"`
	Prices    []ItemPrice `json:"prices"`
}

// Quotation is a supplier quotation header, optionally with its items.
type Quotation struct {
	ID           int64  `json:"id"`
	SupplierID   int64  `json:"supplierId"`
	SupplierName any    `json:"supplierName"`
	QuotedAt     any    `json:"quotedAt"`
	Reference    any    `json:"reference"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ValidUntil   any    `json:"validUntil"`
	Notes        any    `json:"notes"`
	Items        []Item `json:"items,omitempty"`
}

// QuotationFilters narrows a quotation list.
type QuotationFilters struct {
	ID         int64
	SupplierID int64
	Status     string
	Reference  string
	Limit      int
}

const quotationSelect = `
  SELECT q.id, q.supplier_id, s.name AS supplier_name, q.quoted_at, q.reference,
         q.currency, q.status, q.valid_until, q.notes
  FROM supplier_quotations q
  LEFT JOIN suppliers s ON q.supplier_id = s.id
`

// ListQuotations returns quotation headers (no items) newest first.
func ListQuotations(f QuotationFilters) ([]Quotation, error) {
	q := quotationSelect
	var where []string
	var params []any
	if f.ID > 0 {
		where = append(where, "q.id = ?")
		params = append(params, f.ID)
	}
	if f.SupplierID > 0 {
		where = append(where, "q.supplier_id = ?")
		params = append(params, f.SupplierID)
	}
	if f.Status != "" {
		where = append(where, "q.status = ?")
		params = append(params, f.Status)
	}
	if f.Reference != "" {
		where = append(where, "q.reference LIKE ?")
		params = append(params, "%"+f.Reference+"%")
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY q.quoted_at DESC, q.id DESC LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	out := make([]Quotation, 0, len(rows))
	for _, r := range rows {
		out = append(out, scanQuotation(r))
	}
	return out, nil
}

// GetQuotation returns one quotation header with its items and per-colour prices.
func GetQuotation(id int64) (*Quotation, error) {
	row, err := db.QueryOne(quotationSelect+" WHERE q.id = ?", id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	qt := scanQuotation(row)

	items, err := db.Query("SELECT id, car_ref, car_name, size, `condition`, quantity, `year`, `month`, notes "+
		"FROM supplier_quotation_items WHERE quotation_id = ? ORDER BY id ASC", id)
	if err != nil {
		return nil, err
	}
	prices, err := db.Query("SELECT p.item_id, p.color, p.unit_price, p.notes "+
		"FROM supplier_quotation_item_prices p "+
		"JOIN supplier_quotation_items i ON p.item_id = i.id "+
		"WHERE i.quotation_id = ? ORDER BY p.id ASC", id)
	if err != nil {
		return nil, err
	}
	pricesByItem := map[int64][]ItemPrice{}
	for _, p := range prices {
		ref := int64(num(p["item_id"]))
		pricesByItem[ref] = append(pricesByItem[ref], ItemPrice{
			Color:     s(p["color"]),
			UnitPrice: num(p["unit_price"]),
			Notes:     p["notes"],
		})
	}
	qt.Items = make([]Item, 0, len(items))
	for _, it := range items {
		iid := int64(num(it["id"]))
		if pricesByItem[iid] == nil {
			pricesByItem[iid] = []ItemPrice{}
		}
		qt.Items = append(qt.Items, Item{
			ID:        iid,
			CarRef:    it["car_ref"],
			CarName:   s(it["car_name"]),
			Size:      s(it["size"]),
			Condition: s(it["condition"]),
			Quantity:  int(num(it["quantity"])),
			Year:      it["year"],
			Month:     it["month"],
			Notes:     it["notes"],
			Prices:    pricesByItem[iid],
		})
	}
	return &qt, nil
}

func scanQuotation(r map[string]any) Quotation {
	return Quotation{
		ID:           int64(num(r["id"])),
		SupplierID:   int64(num(r["supplier_id"])),
		SupplierName: r["supplier_name"],
		QuotedAt:     db.ToIso(r["quoted_at"]),
		Reference:    r["reference"],
		Currency:     s(r["currency"]),
		Status:       s(r["status"]),
		ValidUntil:   dateOnly(r["valid_until"]),
		Notes:        r["notes"],
	}
}

// CreateQuotation inserts a header and returns its id.
func CreateQuotation(supplierID int64, quotedAt, reference, currency, status string, validUntil, notes any) (int64, error) {
	return db.Insert("INSERT INTO supplier_quotations (supplier_id, quoted_at, reference, currency, status, valid_until, notes) "+
		"VALUES (?,?,?,?,?,?,?)",
		supplierID, quotedAt, nullIfEmpty(reference), currency, status, validUntil, notes)
}

// AddItem appends one car line to a quotation and returns its id. carRef is the
// cars.id the item links to (nil when the car is not in the catalog).
func AddItem(quotationID int64, carRef any, carName, size, condition string, quantity int, year, month any, notes any) (int64, error) {
	_, err := db.QueryOne("SELECT id FROM supplier_quotations WHERE id = ?", quotationID)
	if err != nil {
		return 0, err
	}
	return db.Insert("INSERT INTO supplier_quotation_items (quotation_id, car_ref, car_name, size, `condition`, quantity, `year`, `month`, notes) "+
		"VALUES (?,?,?,?,?,?,?,?,?)",
		quotationID, carRef, strings.TrimSpace(carName), size, condition, quantity, year, month, notes)
}

// SetItemPrice upserts the price of one colour of an item.
func SetItemPrice(itemID int64, color string, unitPrice float64, notes any) (int64, error) {
	return db.Insert("INSERT INTO supplier_quotation_item_prices (item_id, color, unit_price, notes) "+
		"VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE unit_price = VALUES(unit_price), notes = VALUES(notes)",
		itemID, strings.TrimSpace(color), unitPrice, notes)
}

// SetQuotationStatus updates a quotation's status; returns affected rows.
func SetQuotationStatus(id int64, status string) (int64, error) {
	return db.Execute("UPDATE supplier_quotations SET status = ? WHERE id = ?", status, id)
}

// DeleteQuotation removes a header; the FK cascade drops its items and prices.
func DeleteQuotation(id int64) (int64, error) {
	return db.Execute("DELETE FROM supplier_quotations WHERE id = ?", id)
}

// dateOnly renders a DATE column as "YYYY-MM-DD" (nil stays nil).
func dateOnly(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case time.Time:
		return t.Format("2006-01-02")
	case []byte:
		return trunc10(string(t))
	case string:
		return trunc10(t)
	}
	return v
}

func trunc10(s string) string {
	if len(s) > 10 {
		return s[:10]
	}
	return s
}
