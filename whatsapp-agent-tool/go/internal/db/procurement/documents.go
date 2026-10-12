package procurement

import (
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// DocumentTypes are the allowed documents.doc_type values (the DB ENUM).
var DocumentTypes = []string{"proforma_invoice", "sales_contract", "packing_list", "bill_of_lading", "other"}

// DocumentStatuses are the allowed documents.status values (the DB ENUM).
var DocumentStatuses = []string{"received", "accepted", "paid", "cancelled"}

// ValidDocType reports whether v is an allowed doc_type.
func ValidDocType(v string) bool {
	for _, t := range DocumentTypes {
		if t == v {
			return true
		}
	}
	return false
}

// ValidDocStatus reports whether v is an allowed document status.
func ValidDocStatus(v string) bool {
	for _, t := range DocumentStatuses {
		if t == v {
			return true
		}
	}
	return false
}

// LinkedQuotation is a quotation referenced by a document (via document_quotations).
type LinkedQuotation struct {
	ID           int64  `json:"id"`
	Reference    any    `json:"reference"`
	SupplierName any    `json:"supplierName"`
	QuotedAt     any    `json:"quotedAt"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
}

// Document is one row of the ai `documents` table (a real supplier document:
// proforma invoice, sales contract, …). source_text is the transcription, so the
// original wording stays queryable by AI.
type Document struct {
	ID            int64             `json:"id"`
	DocNo         string            `json:"docNo"`
	DocType       string            `json:"docType"`
	DocDate       any               `json:"docDate"`
	SupplierID    any               `json:"supplierId"`
	SupplierName  any               `json:"supplierName"`
	Counterparty  any               `json:"counterparty"`
	Currency      string            `json:"currency"`
	Total         any               `json:"total"`
	AmountInWords any               `json:"amountInWords"`
	SourceText    any               `json:"sourceText"`
	FilePath      any               `json:"filePath"`
	Status        string            `json:"status"`
	Notes         any               `json:"notes"`
	CreatedAt     any               `json:"createdAt"`
	UpdatedAt     any               `json:"updatedAt"`
	Quotations    []LinkedQuotation `json:"quotations,omitempty"`
}

const documentSelect = `
  SELECT d.id, d.doc_no, d.doc_type, d.doc_date, d.supplier_id, s.name AS supplier_name,
         d.counterparty, d.currency, d.total, d.amount_in_words, d.file_path, d.status,
         d.notes, d.created_at, d.updated_at
  FROM documents d
  LEFT JOIN suppliers s ON d.supplier_id = s.id
`

// ListDocuments returns document headers (no source_text, no quotations).
func ListDocuments(supplierID int64, docType string, limit int) ([]Document, error) {
	q := documentSelect
	var where []string
	var params []any
	if supplierID > 0 {
		where = append(where, "d.supplier_id = ?")
		params = append(params, supplierID)
	}
	if strings.TrimSpace(docType) != "" {
		where = append(where, "d.doc_type = ?")
		params = append(params, strings.ToLower(strings.TrimSpace(docType)))
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY d.doc_date DESC, d.id DESC LIMIT " + strconv.Itoa(clampLimit(limit))
	rows, err := db.Query(q, params...)
	if err != nil {
		return nil, err
	}
	out := make([]Document, 0, len(rows))
	for _, r := range rows {
		out = append(out, scanDocument(r))
	}
	return out, nil
}

// GetDocument returns one document with its sourceText and linked quotations.
func GetDocument(id int64) (*Document, error) {
	row, err := db.QueryOne(documentSelect+" WHERE d.id = ?", id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	doc := scanDocument(row)
	src, err := db.QueryOne("SELECT source_text FROM documents WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if src != nil {
		doc.SourceText = src["source_text"]
	}
	quos, err := db.Query(
		"SELECT q.id, q.reference, s.name AS supplier_name, q.quoted_at, q.currency, q.status "+
			"FROM document_quotations dq "+
			"JOIN supplier_quotations q ON dq.quotation_id = q.id "+
			"LEFT JOIN suppliers s ON q.supplier_id = s.id "+
			"WHERE dq.document_id = ? ORDER BY q.quoted_at DESC, q.id DESC", id)
	if err != nil {
		return nil, err
	}
	doc.Quotations = make([]LinkedQuotation, 0, len(quos))
	for _, q := range quos {
		doc.Quotations = append(doc.Quotations, LinkedQuotation{
			ID:           int64(num(q["id"])),
			Reference:    q["reference"],
			SupplierName: q["supplier_name"],
			QuotedAt:     db.ToIso(q["quoted_at"]),
			Currency:     s(q["currency"]),
			Status:       s(q["status"]),
		})
	}
	return &doc, nil
}

func scanDocument(r map[string]any) Document {
	return Document{
		ID:            int64(num(r["id"])),
		DocNo:         s(r["doc_no"]),
		DocType:       s(r["doc_type"]),
		DocDate:       dateOnly(r["doc_date"]),
		SupplierID:    r["supplier_id"],
		SupplierName:  r["supplier_name"],
		Counterparty:  r["counterparty"],
		Currency:      s(r["currency"]),
		Total:         r["total"],
		AmountInWords: r["amount_in_words"],
		FilePath:      r["file_path"],
		Status:        s(r["status"]),
		Notes:         r["notes"],
		CreatedAt:     db.ToIso(r["created_at"]),
		UpdatedAt:     db.ToIso(r["updated_at"]),
	}
}

// CreateDocument inserts a document header. docNo is required, everything else is
// optional (stored NULL when empty). Returns the new id.
func CreateDocument(docNo, docType, docDate string, supplierID any, counterparty, currency string,
	total any, amountInWords, sourceText, filePath, status, notes string) (int64, error) {
	if currency == "" {
		currency = "USD"
	}
	if status == "" {
		status = "received"
	}
	return db.Insert(
		"INSERT INTO documents (doc_no, doc_type, doc_date, supplier_id, counterparty, currency, total, amount_in_words, source_text, file_path, status, notes) "+
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
		strings.TrimSpace(docNo), strings.TrimSpace(docType), nullIfEmpty(docDate), supplierID,
		nullIfEmpty(counterparty), currency, total, nullIfEmpty(amountInWords),
		nullIfEmpty(sourceText), nullIfEmpty(filePath), status, nullIfEmpty(notes))
}

// LinkDocumentQuotation records that a document backs a quotation. Idempotent.
func LinkDocumentQuotation(documentID, quotationID int64) (int64, error) {
	_, err := db.QueryOne("SELECT id FROM documents WHERE id = ?", documentID)
	if err != nil {
		return 0, err
	}
	_, err = db.QueryOne("SELECT id FROM supplier_quotations WHERE id = ?", quotationID)
	if err != nil {
		return 0, err
	}
	return db.Execute("INSERT IGNORE INTO document_quotations (document_id, quotation_id) VALUES (?,?)",
		documentID, quotationID)
}

// UnlinkDocumentQuotation removes a document↔quotation link; returns affected rows.
func UnlinkDocumentQuotation(documentID, quotationID int64) (int64, error) {
	return db.Execute("DELETE FROM document_quotations WHERE document_id = ? AND quotation_id = ?",
		documentID, quotationID)
}
