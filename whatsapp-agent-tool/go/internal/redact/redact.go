// Package redact removes cost- and supplier-bearing fields from CLI output when
// the calling group lacks the matching permission.
//
// This is a defence-in-depth layer, not the only guard: the write/read gate in
// main.go refuses whole commands a group may not use, and the CLI computes the
// final price from cost regardless (a cost-restricted group still gets the
// price — only the inputs behind it are hidden). Redaction is fail-closed: when
// in doubt a sensitive-looking key is dropped, because leaking a supplier's name
// or our cost to a client is worse than an over-terse answer.
//
// The walk is generic (over decoded JSON) so every command's output is covered
// without threading the flags through each handler. It only runs for restricted
// callers, so the byte-for-byte output of operator/daemon runs is untouched.
package redact

import "encoding/json"

// costKeys are dropped unless can_c_cost is granted. They cover the ai DB's own
// cost fields (quote inputs/breakdown, procurement prices, payments) and the
// cars app's buy-side columns. Selling-side figures the operator explicitly
// wants visible (clearance_dza, price_cell, freight, rate, cfr_da) are NOT here.
var costKeys = map[string]bool{
	// quote calculator internals + cost
	"costUsd": true, "carCostUsd": true, "historyCostUsd": true, "colorCosts": true,
	"wholesaleBenefitUsd": true, "retailBenefitUsd": true, "allInRetailBenefitUsd": true,
	"myCostUsd": true, "freightBenefitUsd": true, "effectiveContainerUsd": true, "freightPerUnitUsd": true,
	"inputs": true, "formula": true, "breakdown": true,
	// ai pricing catalog / history
	"buy_price": true, "buyPrice": true,
	// procurement money
	"unitPrice": true, "target_price": true, "targetPrice": true,
	"amount": true, "amountInWords": true,
	"owed": true, "paid": true, "payed": true, "outstanding": true,
	"stored_amount": true, "calculated_amount": true, "total_paid": true,
	"total": true,
	// supplier payment proof / bank details
	"swift": true, "swiftPath": true, "swift_path": true, "swiftRef": true, "swift_ref": true, "uetr": true,
	"senderBic": true, "sender_bic": true, "receiverBic": true, "receiver_bic": true,
	"beneficiaryName": true, "beneficiary_name": true, "beneficiaryAccount": true, "beneficiary_account": true,
	"file_url": true, "fileUrl": true,
}

// supplierKeys are dropped unless can_c_suppliers_info is granted.
var supplierKeys = map[string]bool{
	"supplierId": true, "supplier_id": true, "id_supplier": true,
	"carSupplierId": true, "historySupplierId": true, "historySupplier": true,
	"supplierName": true, "supplier_name": true, "supplier": true,
	"contact_info": true, "contactInfo": true,
}

// Apply returns v with cost and/or supplier keys removed recursively.
// allowSupplier/allowCost true means the caller MAY see those fields (no change).
// The value is round-tripped through JSON, so key order may differ from the
// original — acceptable because restricted output is consumed by the model, not
// compared byte-for-byte to golden files.
func Apply(v any, allowSupplier, allowCost bool) any {
	if allowSupplier && allowCost {
		return v
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return v
	}
	return walk(decoded, allowSupplier, allowCost)
}

func walk(v any, allowSupplier, allowCost bool) any {
	switch t := v.(type) {
	case map[string]any:
		for k := range t {
			if !allowCost && costKeys[k] {
				delete(t, k)
				continue
			}
			if !allowSupplier && supplierKeys[k] {
				delete(t, k)
				continue
			}
			t[k] = walk(t[k], allowSupplier, allowCost)
		}
		return t
	case []any:
		for i := range t {
			t[i] = walk(t[i], allowSupplier, allowCost)
		}
		return t
	default:
		return v
	}
}
