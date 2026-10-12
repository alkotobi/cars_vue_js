// buy-payments — payments we make to a supplier against a purchase bill (table
// buy_payments), mirroring BuyBillPaymentsView.vue. Reads go through the same AI
// user / query passthrough as the other cars-app reads; `add`, `edit` and `delete`
// are deliberate writes. All are gated on can_pay_purchases (payments to
// suppliers are buy-cost figures).
//
//	whatsapp-tool buy-payments --bill-id N                     # one bill: header + its payments
//	whatsapp-tool buy-payments --supplier-id N [--limit N]     # a supplier's payments across bills
//	whatsapp-tool buy-payments --id N                          # a single payment row
//	whatsapp-tool buy-payments [--limit N]                     # all payments, newest first
//	whatsapp-tool buy-payments add --bill-id N --amount V --date ".." --swift <path> [--notes ..] [--user-id N]
//	whatsapp-tool buy-payments edit --id N --amount V --date ".." [--swift <path>] [--notes ..]
//	whatsapp-tool buy-payments delete --id N
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/carshttp"
)

// payPurchasesPermission gates the supplier-payment area (payments we make to a
// supplier are buy-cost figures). It is a dedicated permission (migration 037)
// rather than can_c_buy_price, so a role can be granted just this.
const payPurchasesPermission = "can_pay_purchases"

func cmdBuyPayments(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "add":
			cmdBuyPaymentAdd(args[1:])
			return
		case "edit":
			cmdBuyPaymentEdit(args[1:])
			return
		case "delete":
			cmdBuyPaymentDelete(args[1:])
			return
		}
	}

	filters := carshttp.BuyPaymentFilters{}
	if n, ok := optInt(args, "id"); ok {
		filters.ID = n
	}
	billID, hasBill := optInt(args, "bill-id")
	if hasBill {
		filters.BillID = billID
	}
	if n, ok := optInt(args, "supplier-id"); ok {
		filters.SupplierID = n
	}
	if n, ok := optInt(args, "limit"); ok {
		filters.Limit = n
	}

	requireCarsPermission(payPurchasesPermission)

	rows, err := carshttp.SearchBuyPayments(filters)
	if err != nil {
		carsFail(err)
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}

	// A bill-scoped read also returns the header (supplier + amounts), like the
	// view's fetchBillInfo(). Other scopes have no single header.
	var bill json.RawMessage
	if hasBill {
		bill, err = carshttp.SearchBuyBillHeader(billID)
		if err != nil {
			carsFail(err)
		}
	}

	ok(pair("action", "buy-payments"), pair("filters", filters),
		pair("count", len(rows)), pair("bill", bill), pair("payments", rows))
}

// cmdBuyPaymentAdd records a payment to a supplier against a purchase bill,
// mirroring BuyBillPaymentsView.savePayment() and its buy_bill.payed resync. The
// app's validateForm makes amount, date and the swift document obligatory, so all
// three are required here — never guessed or defaulted. The bill is required too
// (the row is always scoped to one bill), and the swift file is uploaded to
// buy_payment_swifts first (named like the app: bill_<id>_<ts>.<ext>). id_user is
// the AI user (cached at login), overridable with --user-id.
func cmdBuyPaymentAdd(args []string) {
	billID, hasBill := optInt(args, "bill-id")
	if !hasBill {
		failErr("buy-payments add requires --bill-id <id>")
	}
	amount, hasAmount := optFloat(args, "amount")
	if !hasAmount {
		failErr("buy-payments add requires --amount <value>")
	}
	date := strings.TrimSpace(flag(args, "date"))
	if date == "" {
		failErr("buy-payments add requires --date \"YYYY-MM-DD HH:MM:SS\" (the app never posts a blank payment date)")
	}
	swift := strings.TrimSpace(flag(args, "swift"))
	if swift == "" {
		failErr("buy-payments add requires --swift <path> (the swift document; the app requires it too)")
	}

	var userID *int
	if n, ok := optInt(args, "user-id"); ok {
		userID = &n
	} else if id, ok, _ := carshttp.CurrentUserID(); ok {
		userID = &id
	}

	requireCarsPermission(payPurchasesPermission)

	ext := filepath.Ext(swift)
	fileName := fmt.Sprintf("bill_%d_%d%s", billID, time.Now().UnixMilli(), ext)
	swiftPath, err := carshttp.UploadFileNamed(swift, "buy_payment_swifts", fileName)
	if err != nil {
		carsFail(err)
	}

	payID, err := carshttp.CreateBuyPayment(billID, date, amount, swiftPath, flag(args, "notes"), userID)
	if err != nil {
		carsFail(err)
	}

	affected, err := carshttp.ResyncBillPaid(billID)
	if err != nil {
		carsFail(err)
	}

	var userOut any
	if userID != nil {
		userOut = *userID
	}
	ok(pair("action", "buy-payment-add"), pair("id", payID), pair("billId", billID),
		pair("amount", amount), pair("datePayment", date), pair("swiftPath", swiftPath),
		pair("userId", userOut), pair("billPayedResynced", affected))
}

// cmdBuyPaymentEdit edits a payment in place, mirroring
// BuyBillPaymentsView.savePayment(true) + its buy_bill.payed resync. `--id`,
// `--amount` and `--date` are required (the form requires amount and date). The
// swift is optional: `--swift <path>` uploads a replacement to
// buy_payment_swifts, while omitting it keeps the stored document. `--notes`
// omitted keeps the stored note; pass `--notes ""` to clear it. The owning bill
// is resolved from the payment, so its paid total is resynced after the edit.
func cmdBuyPaymentEdit(args []string) {
	id, hasID := optInt(args, "id")
	if !hasID {
		failErr("buy-payments edit requires --id <payment id>")
	}
	amount, hasAmount := optFloat(args, "amount")
	if !hasAmount {
		failErr("buy-payments edit requires --amount <value>")
	}
	date := strings.TrimSpace(flag(args, "date"))
	if date == "" {
		failErr("buy-payments edit requires --date \"YYYY-MM-DD HH:MM:SS\"")
	}

	requireCarsPermission(payPurchasesPermission)

	current, found, err := carshttp.GetBuyPayment(id)
	if err != nil {
		carsFail(err)
	}
	if !found {
		failErr(fmt.Sprintf("buy-payments edit: no payment with id %d", id))
	}

	swiftPath := current.SwiftPath
	if swift := strings.TrimSpace(flag(args, "swift")); swift != "" {
		ext := filepath.Ext(swift)
		fileName := fmt.Sprintf("bill_%d_%d%s", current.IDBuyBill, time.Now().UnixMilli(), ext)
		swiftPath, err = carshttp.UploadFileNamed(swift, "buy_payment_swifts", fileName)
		if err != nil {
			carsFail(err)
		}
	}
	notes := current.Notes
	if has(args, "--notes") {
		notes = flag(args, "notes")
	}

	affected, err := carshttp.UpdateBuyPayment(id, amount, date, swiftPath, notes)
	if err != nil {
		carsFail(err)
	}
	resynced, err := carshttp.ResyncBillPaid(current.IDBuyBill)
	if err != nil {
		carsFail(err)
	}

	ok(pair("action", "buy-payment-edit"), pair("id", id), pair("billId", current.IDBuyBill),
		pair("amount", amount), pair("datePayment", date), pair("swiftPath", swiftPath),
		pair("notes", notes), pair("affectedRows", affected), pair("billPayedResynced", resynced))
}

// cmdBuyPaymentDelete removes a payment, mirroring
// BuyBillPaymentsView.handleDeletePayment() (DELETE by id, then resync the bill's
// paid total). `--id` is required; the owning bill is resolved from the payment
// before it is removed. Run it only when the user asks — deletion is destructive.
func cmdBuyPaymentDelete(args []string) {
	id, hasID := optInt(args, "id")
	if !hasID {
		failErr("buy-payments delete requires --id <payment id>")
	}

	requireCarsPermission(payPurchasesPermission)

	current, found, err := carshttp.GetBuyPayment(id)
	if err != nil {
		carsFail(err)
	}
	if !found {
		failErr(fmt.Sprintf("buy-payments delete: no payment with id %d", id))
	}

	deleted, err := carshttp.DeleteBuyPayment(id)
	if err != nil {
		carsFail(err)
	}
	resynced, err := carshttp.ResyncBillPaid(current.IDBuyBill)
	if err != nil {
		carsFail(err)
	}

	ok(pair("action", "buy-payment-delete"), pair("id", id), pair("billId", current.IDBuyBill),
		pair("deleted", deleted), pair("billPayedResynced", resynced))
}
