// Buy-payment read support: payments we make to a supplier against a purchase
// bill (table buy_payments), mirroring the cars app's BuyBillPaymentsView.vue.
// A bill's header carries the supplier and the amount owed (Σ amount×QTY over
// buy_details, the view's calculated_amount) plus how much has been paid; the
// payment rows join the user who recorded them. Reads run as the AI user behind
// can_c_buy_price (payments to suppliers are buy-cost figures) through the same
// query passthrough as the other cars-app reads.
package carshttp

import (
	"encoding/json"
	"strconv"
	"strings"
)

// BuyBillHeaderQuery returns the fixed SELECT for one purchase bill's header,
// mirroring BuyBillPaymentsView.fetchBillInfo() plus a paid-total subquery (the
// view derives it in JS from the payment rows). Stored and calculated amounts are
// both returned so a caller can see if they disagree.
const BuyBillHeaderQuery = `
  SELECT
    bb.id,
    bb.bill_ref,
    bb.id_supplier,
    s.name AS supplier_name,
    bb.date_buy,
    bb.amount AS stored_amount,
    bb.payed,
    COALESCE((
      SELECT SUM(amount * QTY)
      FROM buy_details
      WHERE id_buy_bill = bb.id
    ), 0) AS calculated_amount,
    COALESCE((
      SELECT SUM(amount)
      FROM buy_payments
      WHERE id_buy_bill = bb.id
    ), 0) AS total_paid
  FROM buy_bill bb
  LEFT JOIN suppliers s ON bb.id_supplier = s.id
  WHERE bb.id = ?
`

// SearchBuyBillHeader runs the fixed header query as the AI user; returns the
// raw row (or nil when the id does not exist).
func SearchBuyBillHeader(id int) (json.RawMessage, error) {
	rows, err := QueryAsAI(BuyBillHeaderQuery, []any{id})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// BuyPaymentFilters narrows the payment list. Every field is optional; the app
// only ever queries by one bill (BuyBillPaymentsView), but supplier-wide and
// single-row reads reuse the same fixed SELECT.
type BuyPaymentFilters struct {
	ID         int `json:"id,omitempty"`
	BillID     int `json:"billId,omitempty"`
	SupplierID int `json:"supplierId,omitempty"`
	Limit      int `json:"limit,omitempty"`
}

// buyPaymentsSelect mirrors BuyBillPaymentsView.fetchPayments() (buy_payments
// joined to users), extended with the bill/supplier join so a payment can be read
// without a second lookup for the supplier.
const buyPaymentsSelect = `
  SELECT
    bp.id,
    bp.id_buy_bill,
    bb.bill_ref,
    bb.id_supplier,
    s.name AS supplier_name,
    bp.date_payment,
    bp.amount,
    bp.swift_path,
    bp.notes,
    bp.id_user,
    u.username AS created_by
  FROM buy_payments bp
  LEFT JOIN buy_bill bb ON bp.id_buy_bill = bb.id
  LEFT JOIN suppliers s ON bb.id_supplier = s.id
  LEFT JOIN users u ON bp.id_user = u.id
`

// BuildBuyPaymentsQuery orders by payment date DESC like the app, with optional
// id / bill-id / supplier-id matches.
func BuildBuyPaymentsQuery(f BuyPaymentFilters) (string, []any) {
	var conditions []string
	var params []any
	if f.ID > 0 {
		conditions = append(conditions, "bp.id = ?")
		params = append(params, f.ID)
	}
	if f.BillID > 0 {
		conditions = append(conditions, "bp.id_buy_bill = ?")
		params = append(params, f.BillID)
	}
	if f.SupplierID > 0 {
		conditions = append(conditions, "bb.id_supplier = ?")
		params = append(params, f.SupplierID)
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	sql := buyPaymentsSelect + where +
		"\n    ORDER BY bp.date_payment DESC, bp.id DESC\n    LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	return sql, params
}

// SearchBuyPayments runs the fixed buy-payments query as the AI user.
func SearchBuyPayments(f BuyPaymentFilters) ([]json.RawMessage, error) {
	sql, params := BuildBuyPaymentsQuery(f)
	return QueryAsAI(sql, params)
}

// BuyPaymentRow is one payment as the write commands need it: the columns they
// update/delete plus the bill it belongs to (so the paid-total resync runs
// against the right header).
type BuyPaymentRow struct {
	ID          int    `json:"id"`
	IDBuyBill   int    `json:"id_buy_bill"`
	DatePayment string `json:"date_payment"`
	Amount      string `json:"amount"`
	SwiftPath   string `json:"swift_path"`
	Notes       string `json:"notes"`
}

// GetBuyPayment returns one payment row by id; ok is false when it does not
// exist (used by edit/delete to resolve the owning bill and any field the caller
// did not override).
func GetBuyPayment(id int) (row BuyPaymentRow, ok bool, err error) {
	rows, err := SearchBuyPayments(BuyPaymentFilters{ID: id, Limit: 1})
	if err != nil {
		return BuyPaymentRow{}, false, err
	}
	if len(rows) == 0 {
		return BuyPaymentRow{}, false, nil
	}
	if err := json.Unmarshal(rows[0], &row); err != nil {
		return BuyPaymentRow{}, false, err
	}
	return row, true, nil
}
