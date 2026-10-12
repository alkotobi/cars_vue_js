// Procurement CLI — the ai database's suppliers + supplier-quotation tables
// (db/schema.sql), linked to the pricing catalog (cars.supplier_id, history
// supplier_id/quotation_id, quotation-item car_ref). Unrelated to the cars-app
// suppliers/inquiries. stdout is one compact JSON line per command.
package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/db"
	"whatsapp-agent-tool/internal/db/procurement"
)

func cmdProcurement(args []string) {
	if len(args) == 0 {
		failErr("usage: whatsapp-tool procurement <suppliers|quotations|documents|payments|outstanding> ...")
	}
	switch args[0] {
	case "suppliers":
		procurementSuppliers(args[1:])
	case "quotations":
		procurementQuotations(args[1:])
	case "documents":
		procurementDocuments(args[1:])
	case "payments":
		procurementPayments(args[1:])
	case "outstanding":
		procurementOutstanding(args[1:])
	default:
		failErr("unknown procurement subcommand \"" + args[0] + "\"")
	}
}

// procurementOutstanding prints, per supplier, how much we owe (from the deals /
// quotations), how much we already paid (the payment ledger), and the difference.
func procurementOutstanding(args []string) {
	sid := intFlag(args, "supplier-id", 0)
	list, err := procurement.Outstanding(int64(sid))
	db.ClosePool()
	if err != nil {
		failErr(err.Error())
	}
	ok(pair("action", "outstanding"), pair("count", len(list)), pair("suppliers", list))
}

// ---------------------------------------------------------------- suppliers

func procurementSuppliers(args []string) {
	sub, rest := splitSub(args)
	switch sub {
	case "", "list":
		name := flag(rest, "name")
		limit := intFlag(rest, "limit", 100)
		sup, err := procurement.ListSuppliers(name, limit)
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "suppliers-list"), pair("count", len(sup)), pair("suppliers", sup))
	case "add":
		name := flag(rest, "name")
		if strings.TrimSpace(name) == "" {
			failErr("suppliers add requires --name")
		}
		id, err := procurement.CreateSupplier(name, flag(rest, "contact"), flag(rest, "notes"))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "supplier-add"), pair("id", id), pair("name", strings.TrimSpace(name)))
	case "update":
		id := intFlag(rest, "id", 0)
		if id == 0 {
			failErr("suppliers update requires --id")
		}
		affected, err := procurement.UpdateSupplier(int64(id), flag(rest, "contact"), flag(rest, "notes"))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "supplier-update"), pair("id", id), pair("affected", affected))
	default:
		failErr("unknown suppliers subcommand \"" + sub + "\"")
	}
}

// ---------------------------------------------------------------- quotations

func procurementQuotations(args []string) {
	sub, rest := splitSub(args)
	switch sub {
	case "", "list":
		f := procurement.QuotationFilters{
			ID:         int64(intFlag(rest, "id", 0)),
			SupplierID: int64(intFlag(rest, "supplier-id", 0)),
			Status:     flag(rest, "status"),
			Reference:  flag(rest, "ref"),
			Limit:      intFlag(rest, "limit", 100),
		}
		if f.Status != "" && !procurement.ValidStatus(f.Status) {
			failErr("invalid --status; one of " + strings.Join(procurement.QuotationStatuses, ", "))
		}
		qs, err := procurement.ListQuotations(f)
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "quotations-list"), pair("count", len(qs)), pair("quotations", qs))

	case "show":
		id := intFlag(rest, "id", firstInt(rest))
		if id == 0 {
			failErr("quotations show requires an id: whatsapp-tool procurement quotations show <id>")
		}
		qt, err := procurement.GetQuotation(int64(id))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		if qt == nil {
			failErr("quotation " + strconv.Itoa(id) + " not found")
		}
		ok(pair("action", "quotation-show"), pair("quotation", qt))

	case "create":
		sid := intFlag(rest, "supplier-id", 0)
		if sid == 0 {
			failErr("quotations create requires --supplier-id")
		}
		date := sqlDt(flag(rest, "date"))
		currency := flag(rest, "currency")
		if currency == "" {
			currency = "USD"
		}
		status := flag(rest, "status")
		if status == "" {
			status = "received"
		}
		if !procurement.ValidStatus(status) {
			failErr("invalid --status; one of " + strings.Join(procurement.QuotationStatuses, ", "))
		}
		var validUntil any = nil
		if v := flag(rest, "valid-until"); v != "" {
			validUntil = v
		}
		id, err := procurement.CreateQuotation(int64(sid), date, flag(rest, "ref"), currency, status, validUntil, optStr(flag(rest, "notes")))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "quotation-create"), pair("id", id), pair("supplierId", sid),
			pair("quotedAt", date), pair("status", status))

	case "add-item":
		qid := intFlag(rest, "quotation-id", 0)
		car := flag(rest, "car")
		if qid == 0 || strings.TrimSpace(car) == "" {
			failErr("add-item requires --quotation-id and --car")
		}
		size := ""
		if has(rest, "big") {
			size = "big"
		} else if has(rest, "small") {
			size = "small"
		} else {
			failErr("add-item requires --big or --small")
		}
		condition := ""
		if has(rest, "used") {
			condition = "used"
		} else if has(rest, "new") {
			condition = "new"
		} else {
			failErr("add-item requires --used or --new")
		}
		var year, month any
		if y := intFlag(rest, "year", 0); y > 0 {
			year = y
		}
		if m := intFlag(rest, "month", 0); m > 0 {
			month = m
		}
		var carRef any
		if ref := strings.TrimSpace(flag(rest, "car-ref")); ref != "" {
			row, err := db.QueryOne("SELECT id FROM cars WHERE car_id = ?", ref)
			if err != nil {
				db.ClosePool()
				failErr(err.Error())
			}
			if row == nil {
				db.ClosePool()
				failErr("--car-ref \"" + ref + "\" is not in the cars catalog — list it with `whatsapp-tool price list-cars`")
			}
			carRef = int64(fn(row["id"]))
		}
		id, err := procurement.AddItem(int64(qid), carRef, car, size, condition, intFlag(rest, "qty", 1), year, month, optStr(flag(rest, "notes")))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "quotation-add-item"), pair("id", id), pair("quotationId", qid),
			pair("carName", strings.TrimSpace(car)))
		if carRef != nil {
			ok(pair("carRef", carRef))
		}

	case "set-price":
		itemID := intFlag(rest, "item-id", 0)
		color := flag(rest, "color")
		priceRaw := flag(rest, "price")
		if itemID == 0 || strings.TrimSpace(color) == "" || priceRaw == "" {
			failErr("set-price requires --item-id, --color and --price")
		}
		price, err := strconv.ParseFloat(priceRaw, 64)
		if err != nil {
			failErr("--price must be a number")
		}
		id, err := procurement.SetItemPrice(int64(itemID), color, price, optStr(flag(rest, "notes")))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "quotation-set-price"), pair("id", id), pair("itemId", itemID),
			pair("color", strings.TrimSpace(color)), pair("unitPrice", price))

	case "set-status":
		id := intFlag(rest, "id", firstInt(rest))
		if id == 0 {
			failErr("set-status requires an id and --status")
		}
		status := flag(rest, "status")
		if !procurement.ValidStatus(status) {
			failErr("invalid --status; one of " + strings.Join(procurement.QuotationStatuses, ", "))
		}
		affected, err := procurement.SetQuotationStatus(int64(id), status)
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "quotation-set-status"), pair("id", id), pair("status", status),
			pair("affectedRows", affected))

	case "delete":
		id := intFlag(rest, "id", firstInt(rest))
		if id == 0 {
			failErr("quotations delete requires an id: whatsapp-tool procurement quotations delete <id>")
		}
		affected, err := procurement.DeleteQuotation(int64(id))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "quotation-delete"), pair("id", id), pair("deleted", affected))

	default:
		failErr("unknown quotations subcommand \"" + sub + "\"")
	}
}

// ---------------------------------------------------------------- helpers

// splitSub peels an optional leading non-flag subcommand off args.
func splitSub(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
		return args[0], args[1:]
	}
	return "", args
}

// firstInt returns the first positional integer in args (0 if none).
func firstInt(args []string) int {
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			continue
		}
		if n, err := strconv.Atoi(a); err == nil {
			return n
		}
	}
	return 0
}

// intFlag parses --name N, returning def when absent or on error.
func intFlag(args []string, name string, def int) int {
	raw := flag(args, name)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		failErr("--" + name + " must be an integer")
	}
	return n
}

// sqlDateStr converts an ISO datetime to "YYYY-MM-DD" for DATE columns;
// empty/unparseable input returns "" (callers decide the default or NULL).
func sqlDateStr(iso string) string {
	t, ok := parseIso(iso)
	if !ok {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

// ---------------------------------------------------------------- documents

func procurementDocuments(args []string) {
	sub, rest := splitSub(args)
	switch sub {
	case "", "list":
		docs, err := procurement.ListDocuments(int64(intFlag(rest, "supplier-id", 0)), flag(rest, "type"), intFlag(rest, "limit", 100))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "documents-list"), pair("count", len(docs)), pair("documents", docs))
	case "create":
		no := flag(rest, "no")
		dtype := strings.TrimSpace(flag(rest, "type"))
		sid := intFlag(rest, "supplier-id", 0)
		if no == "" || dtype == "" || sid == 0 {
			failErr("documents create requires --no, --type and --supplier-id")
		}
		dtype = strings.ToLower(dtype)
		if !procurement.ValidDocType(dtype) {
			failErr("invalid --type; one of " + strings.Join(procurement.DocumentTypes, ", "))
		}
		status := flag(rest, "status")
		if status == "" {
			status = "received"
		}
		if !procurement.ValidDocStatus(status) {
			failErr("invalid --status; one of " + strings.Join(procurement.DocumentStatuses, ", "))
		}
		var total any
		if t := flag(rest, "total"); t != "" {
			f, err := strconv.ParseFloat(t, 64)
			if err != nil {
				failErr("--total must be a number")
			}
			total = f
		}
		text := flag(rest, "text")
		if f := flag(rest, "text-file"); f != "" {
			b, err := os.ReadFile(f)
			if err != nil {
				failErr("--text-file: " + err.Error())
			}
			if len(b) > 5<<20 {
				failErr("--text-file too large (max 5 MB)")
			}
			text = string(b)
		}
		id, err := procurement.CreateDocument(no, dtype, flag(rest, "date"), sid, flag(rest, "counterparty"),
			flag(rest, "currency"), total, flag(rest, "words"), text, flag(rest, "file"),
			status, flag(rest, "notes"))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "document-create"), pair("id", id), pair("docNo", no), pair("docType", dtype),
			pair("supplierId", sid), pair("status", status))
	case "show":
		id := firstInt(rest)
		if id == 0 {
			failErr("documents show requires an id")
		}
		doc, err := procurement.GetDocument(int64(id))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		if doc == nil {
			failErr("no document with id " + strconv.Itoa(id))
		}
		ok(pair("action", "document-show"), pair("document", doc))
	case "link":
		did := intFlag(rest, "document-id", 0)
		qid := intFlag(rest, "quotation-id", 0)
		if did == 0 || qid == 0 {
			failErr("documents link requires --document-id and --quotation-id")
		}
		affected, err := procurement.LinkDocumentQuotation(int64(did), int64(qid))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "document-quotation-link"), pair("documentId", did), pair("quotationId", qid), pair("linked", affected))
	case "unlink":
		did := intFlag(rest, "document-id", 0)
		qid := intFlag(rest, "quotation-id", 0)
		if did == 0 || qid == 0 {
			failErr("documents unlink requires --document-id and --quotation-id")
		}
		affected, err := procurement.UnlinkDocumentQuotation(int64(did), int64(qid))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "document-quotation-unlink"), pair("documentId", did), pair("quotationId", qid), pair("unlinked", affected))
	default:
		failErr("unknown documents subcommand \"" + sub + "\"")
	}
}

// ---------------------------------------------------------------- payments

func procurementPayments(args []string) {
	sub, rest := splitSub(args)
	switch sub {
	case "", "list":
		f := procurement.PaymentFilters{
			SupplierID:  int64(intFlag(rest, "supplier-id", 0)),
			QuotationID: int64(intFlag(rest, "quotation-id", 0)),
			Status:      flag(rest, "status"),
			Limit:       intFlag(rest, "limit", 100),
		}
		if f.Status != "" && !procurement.ValidPaymentStatus(f.Status) {
			failErr("invalid --status; one of " + strings.Join(procurement.PaymentStatuses, ", "))
		}
		ps, err := procurement.ListPayments(f)
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payments-list"), pair("count", len(ps)), pair("payments", ps))

	case "show":
		id := intFlag(rest, "id", firstInt(rest))
		if id == 0 {
			failErr("payments show requires an id: whatsapp-tool procurement payments show <id>")
		}
		p, err := procurement.GetPayment(int64(id))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		if p == nil {
			failErr("payment " + strconv.Itoa(id) + " not found")
		}
		ok(pair("action", "payment-show"), pair("payment", p))

	case "create":
		sid := intFlag(rest, "supplier-id", 0)
		amountRaw := flag(rest, "amount")
		if sid == 0 || amountRaw == "" {
			failErr("payments create requires --supplier-id and --amount")
		}
		amount, err := strconv.ParseFloat(amountRaw, 64)
		if err != nil {
			failErr("--amount must be a number")
		}
		currency := flag(rest, "currency")
		if currency == "" {
			currency = "USD"
		}
		date := sqlDateStr(flag(rest, "date"))
		if date == "" {
			date = time.Now().UTC().Format("2006-01-02")
		}
		no := strings.TrimSpace(flag(rest, "payment-no"))
		if no == "" {
			no, err = procurement.NextPaymentNo()
			if err != nil {
				failErr(err.Error())
			}
		}
		method := flag(rest, "method")
		if method == "" {
			method = "swift"
		}
		if !procurement.ValidPaymentMethod(method) {
			failErr("invalid --method; one of " + strings.Join(procurement.PaymentMethods, ", "))
		}
		status := flag(rest, "status")
		if status == "" {
			status = "sent"
		}
		if !procurement.ValidPaymentStatus(status) {
			failErr("invalid --status; one of " + strings.Join(procurement.PaymentStatuses, ", "))
		}
		id, err := procurement.CreatePayment(no, int64(sid), amount, currency,
			date, sqlDateStr(flag(rest, "value-date")), method, status, flag(rest, "notes"))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payment-create"), pair("id", id), pair("paymentNo", no),
			pair("supplierId", sid), pair("amount", amount), pair("currency", currency),
			pair("status", status))

	case "edit":
		id := intFlag(rest, "id", firstInt(rest))
		if id == 0 {
			failErr("payments edit requires an id: whatsapp-tool procurement payments edit --id <id>")
		}
		sets := map[string]any{}
		if no := strings.TrimSpace(flag(rest, "payment-no")); no != "" {
			sets["payment_no"] = no
		}
		if sid := intFlag(rest, "supplier-id", 0); sid > 0 {
			sets["id_supplier"] = int64(sid)
		}
		if raw := flag(rest, "amount"); raw != "" {
			amount, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				failErr("--amount must be a number")
			}
			sets["amount"] = amount
		}
		if c := flag(rest, "currency"); c != "" {
			sets["currency"] = c
		}
		if has(rest, "date") {
			if d := sqlDateStr(flag(rest, "date")); d != "" {
				sets["payment_date"] = d
			} else {
				sets["payment_date"] = nil
			}
		}
		if has(rest, "value-date") {
			if d := sqlDateStr(flag(rest, "value-date")); d != "" {
				sets["value_date"] = d
			} else {
				sets["value_date"] = nil
			}
		}
		if m := flag(rest, "method"); m != "" {
			if !procurement.ValidPaymentMethod(m) {
				failErr("invalid --method; one of " + strings.Join(procurement.PaymentMethods, ", "))
			}
			sets["method"] = m
		}
		if st := flag(rest, "status"); st != "" {
			if !procurement.ValidPaymentStatus(st) {
				failErr("invalid --status; one of " + strings.Join(procurement.PaymentStatuses, ", "))
			}
			sets["status"] = st
		}
		if has(rest, "notes") {
			sets["notes"] = optStr(flag(rest, "notes"))
		}
		affected, err := procurement.PaymentUpdate(int64(id), sets)
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payment-edit"), pair("id", id), pair("affectedRows", affected))

	case "delete":
		id := intFlag(rest, "id", firstInt(rest))
		if id == 0 {
			failErr("payments delete requires an id: whatsapp-tool procurement payments delete <id>")
		}
		affected, err := procurement.DeletePayment(int64(id))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payment-delete"), pair("id", id), pair("deleted", affected))

	case "link":
		pid := intFlag(rest, "payment-id", 0)
		qid := intFlag(rest, "quotation-id", 0)
		if pid == 0 || qid == 0 {
			failErr("payments link requires --payment-id and --quotation-id")
		}
		affected, err := procurement.LinkPaymentQuotation(int64(pid), int64(qid))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payment-quotation-link"), pair("paymentId", pid), pair("quotationId", qid), pair("linked", affected))

	case "unlink":
		pid := intFlag(rest, "payment-id", 0)
		qid := intFlag(rest, "quotation-id", 0)
		if pid == 0 || qid == 0 {
			failErr("payments unlink requires --payment-id and --quotation-id")
		}
		affected, err := procurement.UnlinkPaymentQuotation(int64(pid), int64(qid))
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payment-quotation-unlink"), pair("paymentId", pid), pair("quotationId", qid), pair("unlinked", affected))

	case "set-swift":
		pid := intFlag(rest, "payment-id", 0)
		if pid == 0 {
			failErr("payments set-swift requires --payment-id")
		}
		text := flag(rest, "text")
		if f := flag(rest, "text-file"); f != "" {
			b, err := os.ReadFile(f)
			if err != nil {
				failErr("--text-file: " + err.Error())
			}
			if len(b) > 5<<20 {
				failErr("--text-file too large (max 5 MB)")
			}
			text = string(b)
		}
		sw := procurement.PaymentSwift{
			SwiftRef:           flag(rest, "swift-ref"),
			Uetr:               flag(rest, "uetr"),
			SenderBIC:          flag(rest, "sender-bic"),
			ReceiverBIC:        flag(rest, "receiver-bic"),
			BeneficiaryName:    flag(rest, "beneficiary-name"),
			BeneficiaryAccount: flag(rest, "beneficiary-account"),
			FileURL:            flag(rest, "file-url"),
			FilePath:           flag(rest, "file"),
			SourceText:         text,
		}
		affected, err := procurement.SetPaymentSwift(int64(pid), sw)
		db.ClosePool()
		if err != nil {
			failErr(err.Error())
		}
		ok(pair("action", "payment-set-swift"), pair("paymentId", pid), pair("affectedRows", affected))

	default:
		failErr("unknown payments subcommand \"" + sub + "\"")
	}
}
