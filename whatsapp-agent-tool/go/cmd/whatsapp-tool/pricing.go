// Pricing write CLI — the Go twin of src/pricing-cli.ts. The ONLY way to change
// pricing now that MySQL is the source of truth (PRICING.md rules: append-only
// history, time-series config). Every write runs in a transaction; stdout is one
// compact JSON line.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/db"
	dbp "whatsapp-agent-tool/internal/db/pricing"
	proc "whatsapp-agent-tool/internal/db/procurement"
)

func jsonArg(args []string, name string) map[string]any {
	raw := flag(args, name)
	if raw == "" {
		failErr("--" + name + " is required")
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		failErr("--" + name + " is not valid JSON: " + err.Error())
	}
	m, ok := v.(map[string]any)
	if !ok {
		failErr("--" + name + " must be a JSON object")
	}
	return m
}

func numFlag(args []string, name string) float64 {
	raw := flag(args, name)
	if raw == "" {
		failErr("--" + name + " is required")
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		failErr("--" + name + " must be a number")
	}
	return n
}

// sqlDt mirrors Node's sqlDt(): parse iso -> "YYYY-MM-DD HH:MM:SS" (UTC), NaN -> now.
func sqlDt(iso string) string {
	t, ok := parseIso(iso)
	if !ok {
		return nowSql()
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

func parseIso(iso string) (time.Time, bool) {
	if iso == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.Parse(layout, iso); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func nowSql() string {
	// "%Y-%m-%d %H:%M:%S" in UTC, like Node's new Date().toISOString().slice(0,19).
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

func cmdPrice(args []string) {
	if len(args) == 0 {
		failErr("usage: whatsapp-tool price <list-cars|set-car|set-option|set-rate|set-benefits|set-freight|set-clearance|set-customs-config>")
	}
	var result []orderedPair
	var err error
	switch args[0] {
	case "list-cars":
		result, err = priceListCars()
	case "check":
		result, err = priceCheck()
	case "set-car":
		result, err = priceSetCar(args[1:])
	case "set-option":
		result, err = priceSetOption(args[1:])
	case "set-rate":
		result, err = priceSetRate(args[1:])
	case "set-benefits":
		result, err = priceSetBenefits(args[1:])
	case "set-freight":
		result, err = priceSetFreight(args[1:])
	case "set-clearance":
		result, err = priceSetClearance(args[1:])
	case "set-customs-config":
		result, err = priceSetCustomsConfig(args[1:])
	default:
		failErr("unknown command \"" + args[0] + "\"")
	}
	db.ClosePool()
	if err != nil {
		failErr(err.Error())
	}
	ok(result...)
}

// --------------------------------------------------------------- list-cars

func priceListCars() ([]orderedPair, error) {
	rows, err := db.Query("SELECT car_id, name, size, `condition`, cost_usd, supplier_id FROM cars ORDER BY name")
	if err != nil {
		return nil, err
	}
	type row struct {
		CarId      string  `json:"carId"`
		Name       string  `json:"name"`
		Size       string  `json:"size"`
		Condition  string  `json:"condition"`
		CostUsd    float64 `json:"costUsd"`
		SupplierId any     `json:"supplierId"`
	}
	cars := make([]row, 0, len(rows))
	for _, r := range rows {
		cars = append(cars, row{CarId: str(r["car_id"]), Name: str(r["name"]), Size: str(r["size"]),
			Condition: str(r["condition"]), CostUsd: fn(r["cost_usd"]), SupplierId: r["supplier_id"]})
	}
	return []orderedPair{pair("action", "list-cars"), pair("cars", cars)}, nil
}

// --------------------------------------------------------------- check

func priceCheck() ([]orderedPair, error) {
	issues, err := dbp.CheckPriceInvariants()
	if err != nil {
		return nil, err
	}
	status := "ok"
	if len(issues) > 0 {
		status = "issues"
	}
	var list []map[string]any
	if len(issues) > 0 {
		list = make([]map[string]any, 0, len(issues))
		for _, it := range issues {
			list = append(list, map[string]any{
				"carId":             it.CarId,
				"severity":          it.Severity,
				"message":           it.Message,
				"carCostUsd":        it.CarCostUsd,
				"historyCostUsd":    it.HistoryCostUsd,
				"carPriceAt":        it.CarPriceAt,
				"historyPriceAt":    it.HistoryPriceAt,
				"carSupplierId":     it.CarSupplierId,
				"historySupplierId": it.HistorySupplier,
			})
		}
	}
	return []orderedPair{
		pair("action", "price-check"),
		pair("invariant", "price-cache"),
		pair("status", status),
		pair("issueCount", len(list)),
		pair("issues", list),
	}, nil
}

// --------------------------------------------------------------- set-car

func priceSetCar(args []string) ([]orderedPair, error) {
	c := jsonArg(args, "json")
	carId := str(c["carId"])
	name := str(c["name"])
	if carId == "" || name == "" {
		failErr("car.carId and car.name are required")
	}
	supplierName := flag(args, "supplier")
	size := "small"
	if c["size"] == "big" {
		size = "big"
	}
	condition := "new"
	if c["condition"] == "used" {
		condition = "used"
	}
	costUsd := fn(c["costUsd"])
	if costUsd != costUsd { // NaN
		failErr("car.costUsd must be a number")
	}
	// Effective price date: an explicit --price-at wins; otherwise the new
	// price is dated "now", EXCEPT when the cost is unchanged — then the price
	// keeps the date of the latest history row, so the cache never drifts from
	// history. (`cars.price_at` must equal the newest history `price_at`; see
	// `price check`.)
	explicitAt := sqlDt(str(c["priceAt"]))
	hasExplicit := str(c["priceAt"]) != ""

	// Supplier attribution: every price must know who it came from. Optional
	// but strongly advised (the "prices related with suppliers" rule).
	var supplierID any
	if supplierName != "" {
		id, found, err := proc.FindSupplierID(supplierName)
		if err != nil {
			return nil, err
		}
		if !found {
			failErr("no supplier named \"" + supplierName + "\" — create it first with `whatsapp-tool procurement suppliers add --name \"" + supplierName + "\"`")
		}
		supplierID = id
	}

	// Optional provenance: which quotation this price came from.
	var quotationID any
	if q := flag(args, "quotation"); q != "" {
		qid, err := strconv.ParseInt(q, 10, 64)
		if err != nil || qid < 1 {
			failErr("--quotation must be a positive id")
		}
		row, err := db.QueryOne("SELECT id FROM supplier_quotations WHERE id = ?", qid)
		if err != nil {
			return nil, err
		}
		if row == nil {
			failErr("no quotation with id " + q + " — list them with `whatsapp-tool procurement quotations`")
		}
		quotationID = qid
	}

	var out []orderedPair
	_, err := db.WithTransactionR(func(tx *sql.Tx) (any, error) {
		rows, err := db.TxQuery(tx, "SELECT id, cost_usd FROM cars WHERE car_id = ?", carId)
		if err != nil {
			return nil, err
		}
		var prevCost *float64
		var carRef int64
		if len(rows) > 0 {
			v := fn(rows[0]["cost_usd"])
			prevCost = &v
			carRef = int64(fn(rows[0]["id"]))
		}
		var loadingPort, notes any = nil, nil
		if v, ok := c["loadingPort"].(string); ok && v != "" {
			loadingPort = v
		}
		if v, ok := c["notes"].(string); ok && v != "" {
			notes = v
		}
		priceAt := explicitAt
		if !hasExplicit {
			priceAt = nowSql()
			if prevCost != nil && *prevCost == costUsd {
				hist, herr := db.TxQuery(tx, "SELECT price_at FROM car_price_history WHERE car_id = ? ORDER BY price_at DESC, id DESC LIMIT 1", carId)
				if herr != nil {
					return nil, herr
				}
				if len(hist) > 0 && hist[0]["price_at"] != nil {
					if iso := db.ToIso(hist[0]["price_at"]); iso != nil {
						priceAt = sqlDt(*iso)
					}
				}
			}
		}
		_, _, err = db.TxExec(tx,
			"INSERT INTO cars (car_id, name, size, `condition`, cost_usd, supplier_id, loading_port, notes, price_at) "+
				"VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE "+
				"name=VALUES(name), size=VALUES(size), `condition`=VALUES(`condition`), cost_usd=VALUES(cost_usd), "+
				"supplier_id=VALUES(supplier_id), loading_port=VALUES(loading_port), notes=VALUES(notes), price_at=VALUES(price_at)",
			carId, name, size, condition, costUsd, supplierID, loadingPort, notes, priceAt)
		if err != nil {
			return nil, err
		}
		if carRef == 0 {
			row, qerr := db.TxQuery(tx, "SELECT id FROM cars WHERE car_id = ?", carId)
			if qerr != nil {
				return nil, qerr
			}
			carRef = int64(fn(row[0]["id"]))
		}

		if colors, ok := c["colors"].([]any); ok {
			if _, _, err = db.TxExec(tx, "DELETE FROM car_colors WHERE car_ref = ?", carRef); err != nil {
				return nil, err
			}
			for _, col := range colors {
				// A colour is a bare name ("white") or an object with a
				// per-colour buy cost override ({"name":"white","costUsd":11300}).
				name := str(col)
				var colCost any = nil
				if m, ok := col.(map[string]any); ok {
					name = str(m["name"])
					if name == "" {
						name = str(m["color"])
					}
					if v, ok := m["costUsd"].(float64); ok {
						colCost = v
					}
				}
				if name == "" {
					continue
				}
				if _, _, err = db.TxExec(tx, "INSERT INTO car_colors (car_ref, color, cost_usd) VALUES (?,?,?)", carRef, name, colCost); err != nil {
					return nil, err
				}
			}
		}
		if opts, ok := c["options"].([]any); ok {
			if _, _, err = db.TxExec(tx, "DELETE FROM car_options WHERE car_ref = ?", carRef); err != nil {
				return nil, err
			}
			for _, o := range opts {
				om, ok := o.(map[string]any)
				if !ok {
					continue
				}
				var priceUsd any = nil
				if p, ok := om["priceUsd"].(float64); ok {
					priceUsd = p
				}
				if _, _, err = db.TxExec(tx, "INSERT INTO car_options (car_ref, option_name, price_usd) VALUES (?,?,?)",
					carRef, str(om["name"]), priceUsd); err != nil {
					return nil, err
				}
			}
		}

		changed := prevCost == nil || *prevCost != costUsd
		if changed {
			note := str(c["historyNote"])
			if note == "" {
				if prevCost == nil {
					note = "New car."
				} else {
					note = fmt.Sprintf("Buy price changed from %v.", *prevCost)
				}
			}
			if _, _, err = db.TxExec(tx,
				"INSERT INTO car_price_history (car_ref, car_id, name, cost_usd, price_at, note, supplier_id, quotation_id) VALUES (?,?,?,?,?,?,?,?)",
				carRef, carId, name, costUsd, priceAt, note, supplierID, quotationID); err != nil {
				return nil, err
			}
		}
		out = []orderedPair{pair("action", "set-car"), pair("carId", carId), pair("costUsd", costUsd), pair("historyAppended", changed)}
		if supplierID != nil {
			out = append(out, pair("supplierId", supplierID))
		}
		if quotationID != nil {
			out = append(out, pair("quotationId", quotationID))
		}
		return nil, nil
	})
	return out, err
}

// --------------------------------------------------------------- set-option

func priceSetOption(args []string) ([]orderedPair, error) {
	name := flag(args, "name")
	if name == "" {
		failErr("set-option requires --name")
	}
	price := numFlag(args, "price")
	err := db.WithTransaction(func(tx *sql.Tx) error {
		_, _, err := db.TxExec(tx,
			"INSERT INTO options (name, price_usd) VALUES (?,?) ON DUPLICATE KEY UPDATE price_usd=VALUES(price_usd)",
			name, price)
		return err
	})
	if err != nil {
		return nil, err
	}
	return []orderedPair{pair("action", "set-option"), pair("name", name), pair("priceUsd", price)}, nil
}

// --------------------------------------------------------------- set-rate

func priceSetRate(args []string) ([]orderedPair, error) {
	value := numFlag(args, "value")
	at := sqlDt(flag(args, "at"))
	err := db.WithTransaction(func(tx *sql.Tx) error {
		_, _, err := db.TxExec(tx, "INSERT INTO config (config_key, value_json, effective_at) VALUES ('rate', ?, ?)",
			mustJSON(map[string]any{"usdToDza": value}), at)
		return err
	})
	if err != nil {
		return nil, err
	}
	return []orderedPair{pair("action", "set-rate"), pair("usdToDza", value), pair("effectiveAt", at)}, nil
}

// --------------------------------------------------------------- set-benefits

func priceSetBenefits(args []string) ([]orderedPair, error) {
	wholesale := numFlag(args, "wholesale")
	retail := numFlag(args, "retail")
	at := sqlDt(flag(args, "at"))
	err := db.WithTransaction(func(tx *sql.Tx) error {
		_, _, err := db.TxExec(tx,
			"INSERT INTO config (config_key, value_json, effective_at) VALUES ('benefits', ?, ?)",
			mustJSON(map[string]any{"wholesaleBenefitUsd": wholesale, "retailBenefitUsd": retail}), at)
		return err
	})
	if err != nil {
		return nil, err
	}
	return []orderedPair{
		pair("action", "set-benefits"), pair("wholesaleBenefitUsd", wholesale),
		pair("retailBenefitUsd", retail), pair("effectiveAt", at),
	}, nil
}

// --------------------------------------------------------------- set-freight

func priceSetFreight(args []string) ([]orderedPair, error) {
	f := jsonArg(args, "json")
	for _, k := range []string{"loadingPort", "dischargePort", "shippingLine", "container40ftUsd"} {
		if _, present := f[k]; !present {
			failErr("freight." + k + " is required")
		}
	}
	at := sqlDt(flag(args, "at"))
	slices := func(v any) any {
		s, ok := v.(string)
		if !ok {
			return nil
		}
		if len(s) > 10 {
			s = s[:10]
		}
		return s
	}

	var id int64
	var err error
	err = db.WithTransaction(func(tx *sql.Tx) error {
		var freeTime any = nil
		if v, ok := f["freeTimeDays"].(float64); ok {
			freeTime = v
		}
		var notes any = nil
		if v, ok := f["notes"].(string); ok && v != "" {
			notes = v
		}
		lp, e := upsertLookup(tx, "loading_ports", str(f["loadingPort"]))
		if e != nil {
			return e
		}
		dp, e := upsertLookup(tx, "discharge_ports", str(f["dischargePort"]))
		if e != nil {
			return e
		}
		sl, e := upsertLookup(tx, "shipping_lines", str(f["shippingLine"]))
		if e != nil {
			return e
		}
		var eid int64
		_, eid, err = db.TxExec(tx,
			"INSERT INTO freight (id_loading_port, id_discharge_port, id_shipping_line, container_40ft_usd, valid_from, valid_to, etd, free_time_days, notes, quoted_at) "+
				"VALUES (?,?,?,?,?,?,?,?,?,?)",
			lp, dp, sl, fn(f["container40ftUsd"]),
			slices(f["validFrom"]), slices(f["validTo"]), slices(f["etd"]), freeTime, notes, at)
		id = eid
		return err
	})
	if err != nil {
		return nil, err
	}
	return []orderedPair{
		pair("action", "set-freight"), pair("id", id),
		pair("shippingLine", str(f["shippingLine"])), pair("quotedAt", at),
	}, nil
}

// --------------------------------------------------------------- set-clearance

func priceSetClearance(args []string) ([]orderedPair, error) {
	e := jsonArg(args, "json")
	name := str(e["name"])
	clearanceDza := fn(e["clearanceDza"])
	if name == "" || clearanceDza != clearanceDza {
		failErr("clearance.name and clearance.clearanceDza are required")
	}
	var carId any = nil
	if v, ok := e["carId"].(string); ok && v != "" {
		carId = v
	}
	effectiveAt := sqlDt(str(e["effectiveAt"]))
	var aliases []any
	if a, ok := e["aliases"].([]any); ok {
		aliases = a
	}
	var id int64
	action := "inserted"
	var err error
	_, err = db.WithTransactionR(func(tx *sql.Tx) (any, error) {
		var carRef any = nil
		if carId != nil {
			rows, qerr := db.TxQuery(tx, "SELECT id FROM cars WHERE car_id = ?", carId)
			if qerr != nil {
				return nil, qerr
			}
			if len(rows) > 0 {
				carRef = int64(fn(rows[0]["id"]))
			}
		}
		if carId != nil {
			rows, qerr := db.TxQuery(tx, "SELECT id FROM clearance WHERE car_id = ?", carId)
			if qerr != nil {
				return nil, qerr
			}
			if len(rows) > 0 {
				id = int64(fn(rows[0]["id"]))
				if _, _, werr := db.TxExec(tx,
					"UPDATE clearance SET name=?, clearance_dza=?, note=?, effective_at=?, car_ref=? WHERE id=?",
					name, clearanceDza, optStr(e["note"]), effectiveAt, carRef, id); werr != nil {
					return nil, werr
				}
				if _, _, werr := db.TxExec(tx, "DELETE FROM clearance_aliases WHERE clearance_id = ?", id); werr != nil {
					return nil, werr
				}
				action = "updated"
			}
		}
		if action == "inserted" {
			_, newID, werr := db.TxExec(tx,
				"INSERT INTO clearance (car_ref, car_id, name, clearance_dza, note, effective_at) VALUES (?,?,?,?,?,?)",
				carRef, carId, name, clearanceDza, optStr(e["note"]), effectiveAt)
			if werr != nil {
				return nil, werr
			}
			id = newID
		}
		for _, alias := range aliases {
			if _, _, werr := db.TxExec(tx, "INSERT INTO clearance_aliases (clearance_id, alias) VALUES (?,?)", id, str(alias)); werr != nil {
				return nil, werr
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return []orderedPair{
		pair("action", "set-clearance"), pair("result", action), pair("id", id),
		pair("carId", carId), pair("clearanceDza", clearanceDza),
	}, nil
}

// --------------------------------------------------------------- set-customs-config

func priceSetCustomsConfig(args []string) ([]orderedPair, error) {
	allIn := numFlag(args, "all-in")
	tvn := flag(args, "tvn") == "1" || flag(args, "tvn") == "true"
	at := nowSql()
	err := db.WithTransaction(func(tx *sql.Tx) error {
		_, _, err := db.TxExec(tx,
			"INSERT INTO config (config_key, value_json, effective_at) VALUES ('customs', ?, ?)",
			mustJSON(map[string]any{"currency": "DZA", "tvnIncluded": tvn, "allInRetailBenefitUsd": allIn}), at)
		return err
	})
	if err != nil {
		return nil, err
	}
	return []orderedPair{pair("action", "set-customs-config"), pair("allInRetailBenefitUsd", allIn), pair("tvnIncluded", tvn)}, nil
}

// --------------------------------------------------------------- helpers

// mustJSON serialises v for a config.value_json column.
func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// upsertLookup returns the id of the lookup row named `name`, inserting it if
// needed. `table` is a fixed literal from the caller (never user input).
func upsertLookup(tx *sql.Tx, table, name string) (int64, error) {
	_, id, err := db.TxExec(tx,
		"INSERT INTO "+table+" (name) VALUES (?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)", name)
	return id, err
}

// optStr mirrors `?? null`: empty/missing string scans to nil.
func optStr(v any) any {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return ""
}

func fn(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}

var _ = strings.TrimSpace
var _ = os.Stdout
