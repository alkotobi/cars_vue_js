// Buy-payment write support: record a payment we make to a supplier against a
// purchase bill, mirroring BuyBillPaymentsView.vue savePayment() (INSERT INTO
// buy_payments) and its updateBillPaidAmount() resync of buy_bill.payed. This is
// a deliberate write (the AI user's role must grant can_pay_purchases), so it lives
// in its own file rather than folded into reads.
package carshttp

// BuildCreateBuyPaymentQuery returns the fixed INSERT for one supplier payment.
// amount, date_payment and swift_path are the app's obligatory fields (its
// validateForm refuses a blank amount/date/swift); id_user is the recording user
// (nil leaves it NULL when unknown); notes is optional.
func BuildCreateBuyPaymentQuery(billID int, datePayment string, amount float64, swiftPath, notes string, userID *int) (string, []any) {
	cols := []string{"id_buy_bill", "date_payment", "amount", "swift_path", "notes"}
	params := []any{billID, datePayment, amount, swiftPath, notes}
	if userID != nil {
		cols = append(cols, "id_user")
		params = append(params, *userID)
	}
	return insertQuery("buy_payments", cols), params
}

// CreateBuyPayment inserts one supplier payment and returns its new id.
func CreateBuyPayment(billID int, datePayment string, amount float64, swiftPath, notes string, userID *int) (int64, error) {
	sql, params := BuildCreateBuyPaymentQuery(billID, datePayment, amount, swiftPath, notes, userID)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.LastInsertID, nil
}

// BuildResyncBillPaidQuery returns the fixed UPDATE that recomputes one bill's
// payed total from its payments, mirroring updateBillPaidAmount().
func BuildResyncBillPaidQuery(billID int) (string, []any) {
	return "UPDATE buy_bill SET payed = (SELECT COALESCE(SUM(amount), 0) FROM buy_payments WHERE id_buy_bill = ?) WHERE id = ?",
		[]any{billID, billID}
}

// ResyncBillPaid recomputes buy_bill.payed for one bill after a payment change
// and returns the number of header rows updated (0 when the id did not exist).
func ResyncBillPaid(billID int) (int64, error) {
	sql, params := BuildResyncBillPaidQuery(billID)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.AffectedRows, nil
}

// BuildUpdateBuyPaymentQuery returns the fixed UPDATE for one payment, mirroring
// BuyBillPaymentsView.savePayment(true). amount and date are obligatory; the
// caller passes the swift path it decided on (a freshly uploaded file, or the
// existing one when the swift was not changed) and the notes to store.
func BuildUpdateBuyPaymentQuery(id int, amount float64, datePayment, swiftPath, notes string) (string, []any) {
	return "UPDATE buy_payments SET amount = ?, date_payment = ?, swift_path = ?, notes = ? WHERE id = ?",
		[]any{amount, datePayment, swiftPath, notes, id}
}

// UpdateBuyPayment applies an edit and returns the number of rows changed.
func UpdateBuyPayment(id int, amount float64, datePayment, swiftPath, notes string) (int64, error) {
	sql, params := BuildUpdateBuyPaymentQuery(id, amount, datePayment, swiftPath, notes)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.AffectedRows, nil
}

// BuildDeleteBuyPaymentQuery returns the fixed DELETE for one payment, mirroring
// BuyBillPaymentsView.handleDeletePayment() (which deletes by id then resyncs the
// bill's paid total).
func BuildDeleteBuyPaymentQuery(id int) (string, []any) {
	return "DELETE FROM buy_payments WHERE id = ?", []any{id}
}

// DeleteBuyPayment removes one payment and returns the number of rows deleted.
func DeleteBuyPayment(id int) (int64, error) {
	sql, params := BuildDeleteBuyPaymentQuery(id)
	result, err := ExecAsAI(sql, params)
	if err != nil {
		return 0, err
	}
	return result.AffectedRows, nil
}
