package procurement

import (
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// PaymentStatuses are the allowed supplier_payments.status values (the DB ENUM).
var PaymentStatuses = []string{"planned", "sent", "cleared", "returned", "cancelled"}

// PaymentMethods are the allowed supplier_payments.method values (the DB ENUM).
var PaymentMethods = []string{"swift", "fps", "chats", "cash", "other"}

// ValidPaymentStatus reports whether v is an allowed payment status.
func ValidPaymentStatus(v string) bool {
	for _, t := range PaymentStatuses {
		if t == v {
			return true
		}
	}
	return false
}

// ValidPaymentMethod reports whether v is an allowed payment method.
func ValidPaymentMethod(v string) bool {
	for _, t := range PaymentMethods {
		if t == v {
			return true
		}
	}
	return false
}

// PaymentSwift is one row of payment_swifts — the SWIFT MT103 proof attached to
// a payment. SourceText is the full manuscript; FileURL the PDF link supplied.
type PaymentSwift struct {
	ID                 int64  `json:"id"`
	PaymentID          int64  `json:"paymentId"`
	SwiftRef           string `json:"swiftRef"`
	Uetr               string `json:"uetr"`
	SenderBIC          string `json:"senderBIC"`
	ReceiverBIC        string `json:"receiverBIC"`
	BeneficiaryName    string `json:"beneficiaryName"`
	BeneficiaryAccount string `json:"beneficiaryAccount"`
	FileURL            string `json:"fileURL"`
	FilePath           string `json:"filePath"`
	SourceText         string `json:"sourceText,omitempty"`
	CreatedAt          any    `json:"createdAt"`
	UpdatedAt          any    `json:"updatedAt"`
}

// SupplierPayment is one row of supplier_payments. Show includes the SWIFT proof
// and the quotation(s) the payment settles (payment_quotations, M:N).
type SupplierPayment struct {
	ID           int64             `json:"id"`
	PaymentNo    string            `json:"paymentNo"`
	SupplierID   any               `json:"supplierId"`
	SupplierName any               `json:"supplierName"`
	Amount       float64           `json:"amount"`
	Currency     string            `json:"currency"`
	PaymentDate  any               `json:"paymentDate"`
	ValueDate    any               `json:"valueDate"`
	Method       string            `json:"method"`
	Status       string            `json:"status"`
	Notes        any               `json:"notes"`
	CreatedAt    any               `json:"createdAt"`
	UpdatedAt    any               `json:"updatedAt"`
	Quotations   []LinkedQuotation `json:"quotations,omitempty"`
	Swift        *PaymentSwift     `json:"swift,omitempty"`
}

const paymentSelect = `
  SELECT p.id, p.payment_no, p.id_supplier, s.name AS supplier_name,
         p.amount, p.currency, p.payment_date, p.value_date, p.method, p.status,
         p.notes, p.created_at, p.updated_at
  FROM supplier_payments p
  LEFT JOIN suppliers s ON p.id_supplier = s.id
`

// PaymentFilters narrows ListPayments.
type PaymentFilters struct {
	SupplierID  int64
	QuotationID int64
	Status      string
	Limit       int
}

// ListPayments returns payment headers (no SWIFT proof, no quotations).
func ListPayments(f PaymentFilters) ([]SupplierPayment, error) {
	q := paymentSelect
	var where []string
	var params []any
	if f.QuotationID > 0 {
		where = append(where, "p.id IN (SELECT payment_id FROM payment_quotations WHERE quotation_id = ?)")
		params = append(params, f.QuotationID)
	}
	if f.SupplierID > 0 {
		where = append(where, "p.id_supplier = ?")
		params = append(params, f.SupplierID)
	}
	if strings.TrimSpace(f.Status) != "" {
		where = append(where, "p.status = ?")
		params = append(params, strings.TrimSpace(f.Status))
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY p.payment_date DESC, p.id DESC LIMIT " + strconv.Itoa(clampLimit(f.Limit))
	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	out := make([]SupplierPayment, 0, len(rows))
	for _, r := range rows {
		out = append(out, scanPayment(r))
	}
	return out, nil
}

// GetPayment returns one payment with its SWIFT proof and linked quotations.
func GetPayment(id int64) (*SupplierPayment, error) {
	row, err := db.QueryOne(paymentSelect+" WHERE p.id = ?", id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	p := scanPayment(row)
	swRow, err := db.QueryOne(
		"SELECT id, id_payment, swift_ref, uetr, sender_bic, receiver_bic, beneficiary_name, "+
			"beneficiary_account, file_url, file_path, source_text, created_at, updated_at "+
			"FROM payment_swifts WHERE id_payment = ?", id)
	if err != nil {
		return nil, err
	}
	if swRow != nil {
		sw := scanSwift(swRow)
		p.Swift = &sw
	}
	quos, err := db.Query(
		"SELECT q.id, q.reference, s.name AS supplier_name, q.quoted_at, q.currency, q.status "+
			"FROM payment_quotations pq "+
			"JOIN supplier_quotations q ON pq.quotation_id = q.id "+
			"LEFT JOIN suppliers s ON q.supplier_id = s.id "+
			"WHERE pq.payment_id = ? ORDER BY q.quoted_at DESC, q.id DESC", id)
	if err != nil {
		return nil, err
	}
	p.Quotations = make([]LinkedQuotation, 0, len(quos))
	for _, q := range quos {
		p.Quotations = append(p.Quotations, LinkedQuotation{
			ID:           int64(num(q["id"])),
			Reference:    q["reference"],
			SupplierName: q["supplier_name"],
			QuotedAt:     db.ToIso(q["quoted_at"]),
			Currency:     s(q["currency"]),
			Status:       s(q["status"]),
		})
	}
	return &p, nil
}

func scanPayment(r map[string]any) SupplierPayment {
	return SupplierPayment{
		ID:           int64(num(r["id"])),
		PaymentNo:    s(r["payment_no"]),
		SupplierID:   r["id_supplier"],
		SupplierName: r["supplier_name"],
		Amount:       num(r["amount"]),
		Currency:     s(r["currency"]),
		PaymentDate:  dateOnly(r["payment_date"]),
		ValueDate:    dateOnly(r["value_date"]),
		Method:       s(r["method"]),
		Status:       s(r["status"]),
		Notes:        r["notes"],
		CreatedAt:    db.ToIso(r["created_at"]),
		UpdatedAt:    db.ToIso(r["updated_at"]),
	}
}

func scanSwift(r map[string]any) PaymentSwift {
	return PaymentSwift{
		ID:                 int64(num(r["id"])),
		PaymentID:          int64(num(r["id_payment"])),
		SwiftRef:           s(r["swift_ref"]),
		Uetr:               s(r["uetr"]),
		SenderBIC:          s(r["sender_bic"]),
		ReceiverBIC:        s(r["receiver_bic"]),
		BeneficiaryName:    s(r["beneficiary_name"]),
		BeneficiaryAccount: s(r["beneficiary_account"]),
		FileURL:            s(r["file_url"]),
		FilePath:           s(r["file_path"]),
		SourceText:         s(r["source_text"]),
		CreatedAt:          db.ToIso(r["created_at"]),
		UpdatedAt:          db.ToIso(r["updated_at"]),
	}
}

// NextPaymentNo returns the next sequential integer reference for the ledger: one
// more than the highest numeric payment_no so far. The ledger NEVER keys a payment on
// a bank reference (see AGENTS.md §16) — the bank's own reference, if any, is stored in
// the swift proof's swift_ref, never as payment_no.
func NextPaymentNo() (string, error) {
	row, err := db.QueryOne(
		"SELECT IFNULL(MAX(CAST(payment_no AS UNSIGNED)), 0) AS n FROM supplier_payments")
	if err != nil {
		return "", err
	}
	n := 0
	if row != nil {
		n = int(num(row["n"]))
	}
	return strconv.Itoa(n + 1), nil
}

// CreatePayment inserts a payment. paymentNo, supplierID and amount are required
// (the CLI refuses to run without them); everything else keeps its default or NULL.
// Quotation links are added separately via LinkPaymentQuotation.
func CreatePayment(paymentNo string, supplierID int64, amount float64,
	currency, paymentDate, valueDate, method, status, notes string) (int64, error) {
	if currency == "" {
		currency = "USD"
	}
	if method == "" {
		method = "swift"
	}
	if status == "" {
		status = "sent"
	}
	return db.Insert(
		"INSERT INTO supplier_payments (payment_no, id_supplier, amount, currency, payment_date, value_date, method, status, notes) "+
			"VALUES (?,?,?,?,?,?,?,?,?)",
		strings.TrimSpace(paymentNo), supplierID, amount, currency,
		nullIfEmpty(paymentDate), nullIfEmpty(valueDate), method, status, nullIfEmpty(notes))
}

// PaymentUpdate maps a column name to its new value (nil = NULL). Only the listed
// columns are accepted; other keys are ignored. Quotation links are not columns —
// use LinkPaymentQuotation / UnlinkPaymentQuotation.
func PaymentUpdate(id int64, sets map[string]any) (int64, error) {
	if len(sets) == 0 || id == 0 {
		return 0, nil
	}
	var cols []string
	var params []any
	for _, c := range []string{"payment_no", "id_supplier", "amount",
		"currency", "payment_date", "value_date", "method", "status", "notes"} {
		if v, ok := sets[c]; ok {
			cols = append(cols, c+" = ?")
			params = append(params, v)
		}
	}
	if len(cols) == 0 {
		return 0, nil
	}
	params = append(params, id)
	return db.Execute("UPDATE supplier_payments SET "+strings.Join(cols, ", ")+" WHERE id = ?", params...)
}

// DeletePayment removes a payment; its SWIFT proof and quotation links are
// removed by CASCADE.
func DeletePayment(id int64) (int64, error) {
	return db.Execute("DELETE FROM supplier_payments WHERE id = ?", id)
}

// LinkPaymentQuotation records that a payment settles a quotation. Idempotent.
func LinkPaymentQuotation(paymentID, quotationID int64) (int64, error) {
	if _, err := db.QueryOne("SELECT id FROM supplier_payments WHERE id = ?", paymentID); err != nil {
		return 0, err
	}
	if _, err := db.QueryOne("SELECT id FROM supplier_quotations WHERE id = ?", quotationID); err != nil {
		return 0, err
	}
	return db.Execute("INSERT IGNORE INTO payment_quotations (payment_id, quotation_id) VALUES (?,?)",
		paymentID, quotationID)
}

// UnlinkPaymentQuotation removes a payment↔quotation link; returns affected rows.
func UnlinkPaymentQuotation(paymentID, quotationID int64) (int64, error) {
	return db.Execute("DELETE FROM payment_quotations WHERE payment_id = ? AND quotation_id = ?",
		paymentID, quotationID)
}

// SetPaymentSwift upserts the single SWIFT proof row of a payment: first call
// inserts, re-running overwrites the same row. An empty string clears that column.
func SetPaymentSwift(paymentID int64, sw PaymentSwift) (int64, error) {
	row, err := db.QueryOne("SELECT id FROM supplier_payments WHERE id = ?", paymentID)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, nil
	}
	return db.Execute(
		"INSERT INTO payment_swifts (id_payment, swift_ref, uetr, sender_bic, receiver_bic, "+
			"beneficiary_name, beneficiary_account, file_url, file_path, source_text) VALUES (?,?,?,?,?,?,?,?,?,?) "+
			"ON DUPLICATE KEY UPDATE swift_ref=VALUES(swift_ref), uetr=VALUES(uetr), sender_bic=VALUES(sender_bic), "+
			"receiver_bic=VALUES(receiver_bic), beneficiary_name=VALUES(beneficiary_name), "+
			"beneficiary_account=VALUES(beneficiary_account), file_url=VALUES(file_url), "+
			"file_path=VALUES(file_path), source_text=VALUES(source_text)",
		paymentID, nullIfEmpty(sw.SwiftRef), nullIfEmpty(sw.Uetr), nullIfEmpty(sw.SenderBIC),
		nullIfEmpty(sw.ReceiverBIC), nullIfEmpty(sw.BeneficiaryName), nullIfEmpty(sw.BeneficiaryAccount),
		nullIfEmpty(sw.FileURL), nullIfEmpty(sw.FilePath), nullIfEmpty(sw.SourceText))
}
