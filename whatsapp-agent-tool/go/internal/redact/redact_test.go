package redact

import (
	"encoding/json"
	"testing"
)

// quoteLike mimics the shape of a `quote` payload: cost fields, supplier, and
// the selling-side price that must survive redaction.
const quoteLike = `{
  "status": "OK",
  "mode": "retail-dza",
  "car": {"carId": "COOLRAY", "name": "COOLRAY BATTLE", "costUsd": 12000, "supplierId": 7, "colors": ["white"]},
  "inputs": {"retailBenefitUsd": 700, "freight": {"myCostUsd": 2500, "freightBenefitUsd": 300}},
  "formula": "costUsd + retailBenefitUsd",
  "breakdown": [{"label": "supplier cost (cars)", "usd": 12000}],
  "price": {"usd": 15200, "dza": 3363631.6, "dzaWords": "..."}
}`

func decode(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestApplyAllAllowedIsUnchanged(t *testing.T) {
	in := decode(t, quoteLike)
	got := Apply(in, true, true).(map[string]any)
	if _, ok := got["inputs"]; !ok {
		t.Fatal("cost allowed: inputs must remain")
	}
	if _, ok := got["car"].(map[string]any)["supplierId"]; !ok {
		t.Fatal("supplier allowed: supplierId must remain")
	}
}

func TestApplyHidesCostKeepsPrice(t *testing.T) {
	got := Apply(decode(t, quoteLike), true, false).(map[string]any)
	for _, k := range []string{"inputs", "formula", "breakdown"} {
		if _, ok := got[k]; ok {
			t.Fatalf("cost hidden: %q must be dropped", k)
		}
	}
	car := got["car"].(map[string]any)
	if _, ok := car["costUsd"]; ok {
		t.Fatal("cost hidden: car.costUsd must be dropped")
	}
	// Price MUST survive — the group still gets the final number.
	price, ok := got["price"].(map[string]any)
	if !ok || price["dza"] == nil || price["usd"] == nil {
		t.Fatalf("cost hidden: price must remain, got %v", got["price"])
	}
	if _, ok := car["supplierId"]; !ok {
		t.Fatal("supplier still allowed: supplierId must remain")
	}
}

func TestApplyHidesSupplierKeepsCost(t *testing.T) {
	got := Apply(decode(t, quoteLike), false, true).(map[string]any)
	car := got["car"].(map[string]any)
	if _, ok := car["supplierId"]; ok {
		t.Fatal("supplier hidden: car.supplierId must be dropped")
	}
	if _, ok := car["costUsd"]; !ok {
		t.Fatal("cost still allowed: car.costUsd must remain")
	}
	if _, ok := got["inputs"]; !ok {
		t.Fatal("cost still allowed: inputs must remain")
	}
}

func TestApplyNestedArraysAndMaps(t *testing.T) {
	in := map[string]any{
		"cars": []any{
			map[string]any{"name": "A", "buy_price": 10, "supplierName": "x"},
			map[string]any{"name": "B", "costUsd": 20},
		},
		"clearance_dza": 950000,
		"price_cell":    12000,
	}
	got := Apply(in, false, false).(map[string]any)
	cars := got["cars"].([]any)
	first := cars[0].(map[string]any)
	if _, ok := first["buy_price"]; ok {
		t.Fatal("nested buy_price must be dropped")
	}
	if _, ok := first["supplierName"]; ok {
		t.Fatal("nested supplierName must be dropped")
	}
	second := cars[1].(map[string]any)
	if _, ok := second["costUsd"]; ok {
		t.Fatal("nested costUsd must be dropped")
	}
	// Selling-side figures the operator wants visible are NOT in the denylist.
	if got["clearance_dza"] != 950000.0 {
		t.Fatal("clearance_dza must survive")
	}
	if got["price_cell"] != 12000.0 {
		t.Fatal("price_cell must survive")
	}
}

func TestApplyNonObject(t *testing.T) {
	if got := Apply([]any{"a", "b"}, false, false); got == nil {
		t.Fatal("non-object payload must not become nil")
	}
}
