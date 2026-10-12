// inquiries — the purchase-side inquiry system (supplier quote requests),
// mirroring the cars app's buy_bill / buy_details (tables inquiry /
// inquiry_details, migration 034). Reads and writes go through the same AI user
// and the query passthrough, gated on can_purchase_cars (modelsPermission) —
// the same permission the buy/models tabs use.
//
//	whatsapp-tool inquiries                                  # list all
//	whatsapp-tool inquiries list [--id N] [--supplier-id N] [--status S] [--ref X] [--limit N]
//	whatsapp-tool inquiries create [--supplier-id N] [--date ..] [--ref X] [--status S] [--notes ..]   # --date defaults to now
//	whatsapp-tool inquiries add-car --inquiry-id N --car-name-id M --color-id C --qty N --target-price P (--used|--new) [--year Y] [--month M] [--big] [--notes ..]
//	whatsapp-tool inquiries cars <id>                        # car lines of one inquiry
//	whatsapp-tool inquiries delete <id>                      # header + lines (FK cascade)
package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/carshttp"
)

// defaultInquiryStatus matches the inquiry.status column default, so a create
// that omits --status still reports the status that is actually stored.
const defaultInquiryStatus = "open"

func cmdInquiries(args []string) {
	if len(args) == 0 {
		inquiriesList(nil)
		return
	}
	switch args[0] {
	case "list":
		inquiriesList(args[1:])
	case "create":
		inquiriesCreate(args[1:])
	case "add-car":
		inquiriesAddCar(args[1:])
	case "cars":
		inquiriesCars(args[1:])
	case "delete":
		inquiriesDelete(args[1:])
	default:
		failErr(`usage: whatsapp-tool inquiries <list | create | add-car | cars <id> | delete <id>>`)
	}
}

// optInt reads an integer flag; ok is false when the flag is absent. A present
// but non-numeric value is a usage error (columns are ints server-side).
func optInt(args []string, name string) (int, bool) {
	v := strings.TrimSpace(flag(args, name))
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		failErr("--" + name + " must be an integer")
	}
	return n, true
}

// optFloat reads a float flag; ok is false when absent.
func optFloat(args []string, name string) (float64, bool) {
	v := strings.TrimSpace(flag(args, name))
	if v == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		failErr("--" + name + " must be a number")
	}
	return f, true
}

func inquiriesList(args []string) {
	filters := carshttp.InquiryFilters{}
	if n, ok := optInt(args, "id"); ok {
		filters.ID = n
	}
	if n, ok := optInt(args, "supplier-id"); ok {
		filters.SupplierID = n
	}
	if n, ok := optInt(args, "limit"); ok {
		filters.Limit = n
	}
	filters.Status = flag(args, "status")
	filters.Ref = flag(args, "ref")

	requireCarsPermission(modelsPermission)

	rows, err := carshttp.SearchInquiries(filters)
	if err != nil {
		carsFail(err)
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}
	ok(pair("action", "inquiries"), pair("filters", filters), pair("count", len(rows)), pair("inquiries", rows))
}

func inquiriesCreate(args []string) {
	var supplierID *int
	if n, ok := optInt(args, "supplier-id"); ok {
		supplierID = &n
	}
	// date_inquiry is required; default to "now" so the caller need not pass it.
	date := strings.TrimSpace(flag(args, "date"))
	if date == "" {
		date = time.Now().Format("2006-01-02 15:04:05")
	}
	ref := flag(args, "ref")
	status := flag(args, "status")
	if status == "" {
		status = defaultInquiryStatus
	}
	notes := flag(args, "notes")

	requireCarsPermission(modelsPermission)

	id, err := carshttp.CreateInquiry(supplierID, date, ref, status, notes)
	if err != nil {
		carsFail(err)
	}

	var supplierOut any
	if supplierID != nil {
		supplierOut = *supplierID
	}
	ok(pair("action", "inquiry-create"), pair("id", id), pair("supplierId", supplierOut), pair("date", date), pair("ref", ref), pair("status", status))
}

func inquiriesAddCar(args []string) {
	inquiryID, hasInquiry := optInt(args, "inquiry-id")
	if !hasInquiry {
		failErr("inquiries add-car requires --inquiry-id <id>")
	}
	carNameID, hasCarName := optInt(args, "car-name-id")
	if !hasCarName {
		failErr("inquiries add-car requires --car-name-id <id>")
	}

	// color, qty and target price are required by the schema (migration 035);
	// require them here too so the error is a clear usage line, not a DB error.
	colorID, hasColor := optInt(args, "color-id")
	if !hasColor {
		failErr("inquiries add-car requires --color-id <id>")
	}
	qty, hasQty := optInt(args, "qty")
	if !hasQty {
		failErr("inquiries add-car requires --qty <n>")
	}
	targetPrice, hasPrice := optFloat(args, "target-price")
	if !hasPrice {
		failErr("inquiries add-car requires --target-price <amount>")
	}
	// used/new is required too; make the caller choose rather than defaulting.
	used, hasUsed := boolChoice(args, "used", "new")
	if !hasUsed {
		failErr("inquiries add-car requires --used or --new (is the car used or new?)")
	}

	car := carshttp.InquiryCar{
		InquiryID:   inquiryID,
		CarNameID:   carNameID,
		ColorID:     &colorID,
		QTY:         &qty,
		TargetPrice: &targetPrice,
		IsUsedCar:   &used,
	}
	if n, ok := optInt(args, "year"); ok {
		car.Year = &n
	}
	if n, ok := optInt(args, "month"); ok {
		car.Month = &n
	}
	if has(args, "big") {
		b := true
		car.IsBigCar = &b
	}
	car.Notes = flag(args, "notes")

	requireCarsPermission(modelsPermission)

	id, err := carshttp.AddInquiryCar(car)
	if err != nil {
		carsFail(err)
	}
	ok(pair("action", "inquiry-car-add"), pair("id", id), pair("inquiryId", inquiryID), pair("carNameId", carNameID))
}

func inquiriesCars(args []string) {
	id, found := inquiryIDArg(args)
	if !found {
		failErr("inquiries cars requires an inquiry id: whatsapp-tool inquiries cars <id>")
	}

	requireCarsPermission(modelsPermission)

	rows, err := carshttp.SearchInquiryCars(carshttp.InquiryCarFilters{InquiryID: id})
	if err != nil {
		carsFail(err)
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}
	ok(pair("action", "inquiry-cars"), pair("inquiryId", id), pair("count", len(rows)), pair("cars", rows))
}

func inquiriesDelete(args []string) {
	id, found := inquiryIDArg(args)
	if !found {
		failErr("inquiries delete requires an inquiry id: whatsapp-tool inquiries delete <id>")
	}

	requireCarsPermission(modelsPermission)

	deleted, err := carshttp.DeleteInquiry(id)
	if err != nil {
		carsFail(err)
	}
	ok(pair("action", "inquiry-delete"), pair("id", id), pair("deleted", deleted))
}

// inquiryIDArg reads the inquiry id from --inquiry-id or the first positional.
func inquiryIDArg(args []string) (int, bool) {
	if n, ok := optInt(args, "inquiry-id"); ok {
		return n, true
	}
	pos := positionalArgs(args, map[string]bool{"--inquiry-id": true})
	if len(pos) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(pos[0]))
	if err != nil {
		return 0, false
	}
	return n, true
}
