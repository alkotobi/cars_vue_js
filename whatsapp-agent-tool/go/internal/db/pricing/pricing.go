// Package pricing is the Go twin of src/db/pricing.ts — pricing catalog loaders
// returning exactly the shapes the old JSON files produced, so the calculator
// logic (internal/pricing) stays untouched.
package pricing

import (
	"encoding/json"

	"whatsapp-agent-tool/internal/db"
)

type CarOption struct {
	Name     string   `json:"name"`
	PriceUsd *float64 `json:"priceUsd,omitempty"`
}

type Car struct {
	CarId       string             `json:"carId"`
	Name        string             `json:"name"`
	Size        string             `json:"size"`      // "big" | "small"
	Condition   string             `json:"condition"` // "new" | "used"
	CostUsd     float64            `json:"costUsd"`
	SupplierId  any                `json:"supplierId"`  // number | null — who supplies the current price
	LoadingPort any                `json:"loadingPort"` // string | null
	Colors      []string           `json:"colors"`
	ColorCosts  map[string]float64 `json:"colorCosts,omitempty"` // per-colour buy cost override (USD)
	Options     []CarOption        `json:"options"`
	Notes       any                `json:"notes"`
	Timestamp   *string            `json:"timestamp"`
}

type OptionDef struct {
	Name     string  `json:"name"`
	PriceUsd float64 `json:"priceUsd"`
}

type Benefits struct {
	WholesaleBenefitUsd float64 `json:"wholesaleBenefitUsd"`
	RetailBenefitUsd    float64 `json:"retailBenefitUsd"`
	Timestamp           *string `json:"timestamp"`
}

type Rate struct {
	UsdToDza  float64 `json:"usdToDza"`
	Timestamp *string `json:"timestamp"`
}

type ClearanceEntry struct {
	CarId        any      `json:"carId"` // string | null
	Name         string   `json:"name"`
	ClearanceDza float64  `json:"clearanceDza"`
	Aliases      []string `json:"aliases"`
	Note         any      `json:"note"`
}

type Custom struct {
	AllInRetailBenefitUsd float64          `json:"allInRetailBenefitUsd"`
	UpdatedAt             *string          `json:"updatedAt"`
	TvnIncluded           bool             `json:"tvnIncluded"`
	Currency              string           `json:"currency"`
	Entries               []ClearanceEntry `json:"entries"`
}

type FreightQuote struct {
	LoadingPort      string  `json:"loadingPort"`
	DischargePort    string  `json:"dischargePort"`
	ShippingLine     string  `json:"shippingLine"`
	Container40ftUsd float64 `json:"container40ftUsd"`
	ValidFrom        any     `json:"validFrom,omitempty"`
	ValidTo          any     `json:"validTo,omitempty"`
	Etd              any     `json:"etd,omitempty"`
	FreeTimeDays     any     `json:"freeTimeDays,omitempty"`
	Timestamp        string  `json:"timestamp"`
	Notes            any     `json:"notes,omitempty"`
}

type Freight struct {
	Freight []FreightQuote `json:"freight"`
}

func num(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case int:
		return float64(t)
	case nil:
		return 0
	default:
		return 0
	}
}

// rawNull keeps NULL as nil but passes any non-nil value (including never
// collapsing "" — Node returns DB strings verbatim).
func rawNull(v any) any {
	if v == nil {
		return nil
	}
	return v
}

func s(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case nil:
		return ""
	}
	return ""
}

// carPriceAtISO normalises a price_at driver value to a UTC ISO string, or nil.
func carPriceAtISO(v any) any {
	if v == nil {
		return nil
	}
	if o := db.ToIso(v); o != nil {
		return *o
	}
	return v
}

func LoadCars() ([]Car, error) {
	cars, err := db.Query("SELECT id, car_id, name, size, `condition`, cost_usd, supplier_id, loading_port, notes, price_at FROM cars ORDER BY name")
	if err != nil {
		return nil, err
	}
	colors, err := db.Query("SELECT car_ref, color, cost_usd FROM car_colors ORDER BY color")
	if err != nil {
		return nil, err
	}
	opts, err := db.Query("SELECT car_ref, option_name, price_usd FROM car_options ORDER BY id")
	if err != nil {
		return nil, err
	}

	colorsByCar := map[int64][]string{}
	colorCostsByCar := map[int64]map[string]float64{}
	for _, r := range colors {
		ref := int64(num(r["car_ref"]))
		name := r["color"].(string)
		colorsByCar[ref] = append(colorsByCar[ref], name)
		if r["cost_usd"] != nil {
			if colorCostsByCar[ref] == nil {
				colorCostsByCar[ref] = map[string]float64{}
			}
			colorCostsByCar[ref][name] = num(r["cost_usd"])
		}
	}
	optionsByCar := map[int64][]CarOption{}
	for _, r := range opts {
		ref := int64(num(r["car_ref"]))
		opt := CarOption{Name: r["option_name"].(string)}
		if p, ok := r["price_usd"]; ok && p != nil {
			f := num(p)
			opt.PriceUsd = &f
		}
		optionsByCar[ref] = append(optionsByCar[ref], opt)
	}

	out := make([]Car, 0, len(cars))
	for _, c := range cars {
		ref := int64(num(c["id"]))
		out = append(out, Car{
			CarId:       c["car_id"].(string),
			Name:        c["name"].(string),
			Size:        c["size"].(string),
			Condition:   c["condition"].(string),
			CostUsd:     num(c["cost_usd"]),
			SupplierId:  rawNull(c["supplier_id"]),
			LoadingPort: rawNull(c["loading_port"]),
			Colors:      colorsByCar[ref],
			ColorCosts:  colorCostsByCar[ref],
			Options:     optionsByCar[ref],
			Notes:       rawNull(c["notes"]),
			Timestamp:   db.ToIso(c["price_at"]),
		})
	}
	return out, nil
}

func LoadOptions() ([]OptionDef, error) {
	rows, err := db.Query("SELECT name, price_usd FROM options ORDER BY name")
	if err != nil {
		return nil, err
	}
	out := make([]OptionDef, 0, len(rows))
	for _, r := range rows {
		out = append(out, OptionDef{Name: r["name"].(string), PriceUsd: num(r["price_usd"])})
	}
	return out, nil
}

// InvariantIssue is one drift between the price cache (cars) and the
// append-only history (car_price_history).
type InvariantIssue struct {
	CarId           string `json:"carId"`
	Severity        string `json:"severity"` // "error" | "warning"
	Message         string `json:"message"`
	CarCostUsd      any    `json:"carCostUsd,omitempty"`
	HistoryCostUsd  any    `json:"historyCostUsd,omitempty"`
	CarPriceAt      any    `json:"carPriceAt,omitempty"`
	HistoryPriceAt  any    `json:"historyPriceAt,omitempty"`
	CarSupplierId   any    `json:"carSupplierId,omitempty"`
	HistorySupplier any    `json:"historySupplierId,omitempty"`
}

// CheckPriceInvariants compares cars' current values against the latest
// car_price_history row of each car. Cost_Usd must always match (error);
// supplier and price_at differences are reported as warnings (they should only
// deviate transiently while a price update is recorded).
func CheckPriceInvariants() ([]InvariantIssue, error) {
	rows, err := db.Query(
		"SELECT c.car_id, c.cost_usd AS car_cost, c.price_at AS car_price_at, c.supplier_id AS car_sup, " +
			"h.id AS hist_id, h.cost_usd AS hist_cost, h.price_at AS hist_price_at, h.supplier_id AS hist_sup " +
			"FROM cars c " +
			"JOIN car_price_history h ON h.id = (" +
			"  SELECT h2.id FROM car_price_history h2 WHERE h2.car_id = c.car_id " +
			"  ORDER BY h2.price_at DESC, h2.id DESC LIMIT 1)")
	if err != nil {
		return nil, err
	}
	orphanRows, err := db.Query(
		"SELECT car_id FROM cars c WHERE NOT EXISTS (SELECT 1 FROM car_price_history h WHERE h.car_id = c.car_id)")
	if err != nil {
		return nil, err
	}

	issues := make([]InvariantIssue, 0, len(rows)+len(orphanRows))
	for _, r := range rows {
		carId := s(r["car_id"])
		carCost, histCost := num(r["car_cost"]), num(r["hist_cost"])
		carSup, histSup := num(r["car_sup"]), num(r["hist_sup"])
		carAt := rawNull(carPriceAtISO(r["car_price_at"]))
		histAt := rawNull(carPriceAtISO(r["hist_price_at"]))
		base := InvariantIssue{
			CarId:           carId,
			CarCostUsd:      carCost,
			HistoryCostUsd:  histCost,
			CarSupplierId:   rawNull(r["car_sup"]),
			HistorySupplier: rawNull(r["hist_sup"]),
			CarPriceAt:      carAt,
			HistoryPriceAt:  histAt,
		}
		if carCost != histCost {
			base.Severity = "error"
			base.Message = "cars.cost_usd differs from the latest car_price_history row."
			issues = append(issues, base)
			continue
		}
		if carSup != histSup {
			base.Severity = "warning"
			base.Message = "cars.supplier_id differs from the latest history supplier. Will fix when a price change is recorded."
			issues = append(issues, base)
		}
		if carAt != histAt {
			base.Severity = "warning"
			base.Message = "cars.price_at differs from the latest history price_at."
			issues = append(issues, base)
		}
	}
	for _, r := range orphanRows {
		issues = append(issues, InvariantIssue{
			CarId: s(r["car_id"]), Severity: "warning",
			Message: "car has no car_price_history rows.",
		})
	}
	return issues, nil
}

func LoadBenefits() (Benefits, error) {
	row, err := db.QueryOne("SELECT value_json, effective_at FROM config WHERE config_key = 'benefits' ORDER BY effective_at DESC, id DESC LIMIT 1")
	if err != nil {
		return Benefits{}, err
	}
	if row == nil {
		return Benefits{}, nil
	}
	var v struct {
		WholesaleBenefitUsd float64 `json:"wholesaleBenefitUsd"`
		RetailBenefitUsd    float64 `json:"retailBenefitUsd"`
	}
	if s, ok := row["value_json"].(string); ok && s != "" {
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return Benefits{}, err
		}
	}
	return Benefits{
		WholesaleBenefitUsd: v.WholesaleBenefitUsd,
		RetailBenefitUsd:    v.RetailBenefitUsd,
		Timestamp:           db.ToIso(row["effective_at"]),
	}, nil
}

func LoadRate() (Rate, error) {
	row, err := db.QueryOne("SELECT value_json, effective_at FROM config WHERE config_key = 'rate' ORDER BY effective_at DESC, id DESC LIMIT 1")
	if err != nil {
		return Rate{}, err
	}
	if row == nil {
		return Rate{}, nil
	}
	var v struct {
		UsdToDza float64 `json:"usdToDza"`
	}
	if s, ok := row["value_json"].(string); ok && s != "" {
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return Rate{}, err
		}
	}
	return Rate{
		UsdToDza:  v.UsdToDza,
		Timestamp: db.ToIso(row["effective_at"]),
	}, nil
}

func LoadCustom() (Custom, error) {
	cfg, err := db.QueryOne("SELECT value_json, effective_at FROM config WHERE config_key = 'customs' ORDER BY effective_at DESC, id DESC LIMIT 1")
	if err != nil {
		return Custom{}, err
	}
	entries, err := db.Query("SELECT id, car_id, name, clearance_dza, note, effective_at FROM clearance ORDER BY name")
	if err != nil {
		return Custom{}, err
	}
	aliases, err := db.Query("SELECT clearance_id, alias FROM clearance_aliases ORDER BY alias")
	if err != nil {
		return Custom{}, err
	}

	aliasesByID := map[int64][]string{}
	for _, a := range aliases {
		id := int64(num(a["clearance_id"]))
		aliasesByID[id] = append(aliasesByID[id], a["alias"].(string))
	}

	out := Custom{Currency: "DZA", Entries: []ClearanceEntry{}}
	if cfg != nil {
		var v struct {
			Currency              string  `json:"currency"`
			TvnIncluded           bool    `json:"tvnIncluded"`
			AllInRetailBenefitUsd float64 `json:"allInRetailBenefitUsd"`
		}
		if s, ok := cfg["value_json"].(string); ok && s != "" {
			if err := json.Unmarshal([]byte(s), &v); err != nil {
				return Custom{}, err
			}
		}
		out.AllInRetailBenefitUsd = v.AllInRetailBenefitUsd
		out.TvnIncluded = v.TvnIncluded
		if v.Currency != "" {
			out.Currency = v.Currency
		}
		out.UpdatedAt = db.ToIso(cfg["effective_at"])
	}
	for _, e := range entries {
		entry := ClearanceEntry{
			CarId:        rawNull(e["car_id"]),
			Name:         e["name"].(string),
			ClearanceDza: num(e["clearance_dza"]),
			Aliases:      aliasesByID[int64(num(e["id"]))],
			Note:         rawNull(e["note"]),
		}
		out.Entries = append(out.Entries, entry)
	}
	return out, nil
}

func LoadFreight() (Freight, error) {
	rows, err := db.Query("SELECT lp.name AS loading_port, dp.name AS discharge_port, sl.name AS shipping_line, " +
		"f.container_40ft_usd, f.valid_from, f.valid_to, f.etd, f.free_time_days, f.notes, f.quoted_at " +
		"FROM freight f " +
		"JOIN loading_ports lp ON f.id_loading_port = lp.id " +
		"JOIN discharge_ports dp ON f.id_discharge_port = dp.id " +
		"JOIN shipping_lines sl ON f.id_shipping_line = sl.id " +
		"ORDER BY f.quoted_at")
	if err != nil {
		return Freight{}, err
	}
	out := Freight{Freight: []FreightQuote{}}
	for _, r := range rows {
		ts := db.ToIso(r["quoted_at"])
		tsStr := ""
		if ts != nil {
			tsStr = *ts
		} else if s, ok := r["quoted_at"].(string); ok {
			tsStr = s
		}
		q := FreightQuote{
			LoadingPort:      r["loading_port"].(string),
			DischargePort:    r["discharge_port"].(string),
			ShippingLine:     r["shipping_line"].(string),
			Container40ftUsd: num(r["container_40ft_usd"]),
			ValidFrom:        rawNull(r["valid_from"]),
			ValidTo:          rawNull(r["valid_to"]),
			Etd:              rawNull(r["etd"]),
			FreeTimeDays:     r["free_time_days"],
			Timestamp:        tsStr,
			Notes:            rawNull(r["notes"]),
		}
		out.Freight = append(out.Freight, q)
	}
	return out, nil
}
