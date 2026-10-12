package procurement

import (
	"sort"

	"whatsapp-agent-tool/internal/db"
)

// OutstandingDeal is one deal (a supplier quotation) with the amount we owe for it,
// derived from the documents (PI + contract) linked to that quotation. A deal's PI
// and contract carry the same total, so the deal counts once. Amount is nil when the
// deal has no non-cancelled document carrying a total (then it is not summed).
type OutstandingDeal struct {
	QuotationID     int64    `json:"quotationId"`
	Reference       any      `json:"reference"`
	QuotationStatus string   `json:"quotationStatus"`
	Amount          any      `json:"amount"`
	DocNos          []string `json:"docNos"`
}

// OutstandingPayment is one recorded payment toward a supplier.
type OutstandingPayment struct {
	ID        int64   `json:"id"`
	PaymentNo string  `json:"paymentNo"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	Date      any     `json:"date"`
}

// SupplierOutstanding is what we owe a supplier minus what we already paid:
// Owed = sum of that supplier's deal amounts (one per quotation, from its documents);
// Paid = sum of payments with status sent/cleared; Outstanding = Owed - Paid.
// Everything is reported for audit — the deal list and the payment list are included.
type SupplierOutstanding struct {
	SupplierID   int64                `json:"supplierId"`
	SupplierName any                  `json:"supplierName"`
	Currency     string               `json:"currency"`
	Owed         float64              `json:"owed"`
	Paid         float64              `json:"paid"`
	Outstanding  float64              `json:"outstanding"`
	Deals        []OutstandingDeal    `json:"deals"`
	Payments     []OutstandingPayment `json:"payments"`
}

// paidStatuses are the payment statuses that count as money actually out the door.
var paidStatuses = map[string]bool{"sent": true, "cleared": true}

// Outstanding computes, per supplier (or one supplier when supplierID > 0), the
// amount owed from the deal/quotations side, the amount already paid, and the
// difference. It never guesses: a deal with no document total is reported with a
// null amount and excluded from the owed sum.
func Outstanding(supplierID int64) ([]SupplierOutstanding, error) {
	byID := map[int64]*SupplierOutstanding{}
	get := func(id int64, name any, cur string) *SupplierOutstanding {
		so, ok := byID[id]
		if !ok {
			if cur == "" {
				cur = "USD"
			}
			so = &SupplierOutstanding{
				SupplierID:   id,
				SupplierName: name,
				Currency:     cur,
				Deals:        []OutstandingDeal{},
				Payments:     []OutstandingPayment{},
			}
			byID[id] = so
		}
		return so
	}

	dealWhere := ""
	var dealArgs []any
	if supplierID > 0 {
		dealWhere = " WHERE sq.supplier_id = ?"
		dealArgs = append(dealArgs, supplierID)
	}
	rows, err := db.Query(
		"SELECT sq.supplier_id AS sid, s.name AS sname, sq.id AS qid, sq.reference AS ref, "+
			"sq.currency AS cur, sq.status AS qstatus, COUNT(d.id) AS doc_count, MAX(d.total) AS deal_total "+
			"FROM supplier_quotations sq "+
			"LEFT JOIN suppliers s ON s.id = sq.supplier_id "+
			"LEFT JOIN document_quotations dq ON dq.quotation_id = sq.id "+
			"LEFT JOIN documents d ON d.id = dq.document_id AND d.status <> 'cancelled' "+
			dealWhere+" GROUP BY sq.id ORDER BY sq.supplier_id, sq.id", dealArgs...)
	if err != nil {
		return nil, err
	}

	docnos := map[int64][]string{}
	drows, err := db.Query(
		"SELECT dq.quotation_id AS qid, d.doc_no AS doc_no FROM document_quotations dq " +
			"JOIN documents d ON d.id = dq.document_id WHERE d.status <> 'cancelled' " +
			"ORDER BY dq.quotation_id, d.id")
	if err != nil {
		return nil, err
	}
	for _, r := range drows {
		qid := int64(num(r["qid"]))
		docnos[qid] = append(docnos[qid], s(r["doc_no"]))
	}

	for _, r := range rows {
		sid := int64(num(r["sid"]))
		so := get(sid, r["sname"], s(r["cur"]))
		deal := OutstandingDeal{
			QuotationID:     int64(num(r["qid"])),
			Reference:       r["ref"],
			QuotationStatus: s(r["qstatus"]),
			DocNos:          docnos[int64(num(r["qid"]))],
		}
		if deal.DocNos == nil {
			deal.DocNos = []string{}
		}
		if num(r["doc_count"]) > 0 && r["deal_total"] != nil {
			deal.Amount = num(r["deal_total"])
			so.Owed += num(r["deal_total"])
		}
		so.Deals = append(so.Deals, deal)
	}

	payWhere := ""
	var payArgs []any
	if supplierID > 0 {
		payWhere = " WHERE p.id_supplier = ?"
		payArgs = append(payArgs, supplierID)
	}
	prows, err := db.Query(
		"SELECT p.id_supplier AS sid, s.name AS sname, p.id AS id, p.payment_no AS payment_no, "+
			"p.amount AS amount, p.currency AS currency, p.status AS status, p.payment_date AS pay_date "+
			"FROM supplier_payments p LEFT JOIN suppliers s ON s.id = p.id_supplier "+
			payWhere+" ORDER BY p.id_supplier, p.payment_date, p.id", payArgs...)
	if err != nil {
		return nil, err
	}
	for _, r := range prows {
		sid := int64(num(r["sid"]))
		so := get(sid, r["sname"], s(r["currency"]))
		st := s(r["status"])
		amt := num(r["amount"])
		so.Payments = append(so.Payments, OutstandingPayment{
			ID:        int64(num(r["id"])),
			PaymentNo: s(r["payment_no"]),
			Amount:    amt,
			Currency:  s(r["currency"]),
			Status:    st,
			Date:      dateOnly(r["pay_date"]),
		})
		if paidStatuses[st] {
			so.Paid += amt
		}
	}

	out := make([]SupplierOutstanding, 0, len(byID))
	for _, so := range byID {
		so.Outstanding = so.Owed - so.Paid
		out = append(out, *so)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SupplierID < out[j].SupplierID })
	return out, nil
}
