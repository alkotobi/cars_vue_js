// Package quote is the Go twin of src/quote.ts — the deterministic pricing
// calculator. The model never does §13 arithmetic itself; the CLI quotes these
// numbers verbatim. Statuses: OK / STALE / MISSING / AMBIGUOUS / USAGE / LIST.
package quote

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	dbp "whatsapp-agent-tool/internal/db/pricing"
	"whatsapp-agent-tool/internal/timex"
)

const maxAgeDays = 7.0          // §13 weekly validity
const freightBenefitUsd = 300.0 // per 40ft container, on top of "my cost"

const dayMs = 24.0 * 60 * 60 * 1000

var Modes = []string{
	"wholesale-fob",
	"wholesale-cfr",
	"retail-dza",
	"retail-usd",
	"retail-usd-fob",
	"allin-dza",
	"clearance",
}

func isMode(m string) bool {
	for _, x := range Modes {
		if x == m {
			return true
		}
	}
	return false
}

// Norm mirrors quote.ts norm(): uppercases, folds FIVAN→LIVAN, collapses
// non-alphanumerics to single spaces, trims.
func Norm(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "FIVAN", "LIVAN")
	// collapse [^A-Z0-9]+ to " "
	var b strings.Builder
	prevSpace := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteByte(c)
			prevSpace = false
		} else if !prevSpace {
			b.WriteByte(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

// Data bundles every input the calculator loads — mirrors the six loaders in
// src/quote.ts (Promise.all).
type Data struct {
	Benefits dbp.Benefits
	Rate     dbp.Rate
	Custom   dbp.Custom
	Freight  dbp.Freight
	Cars     []dbp.Car
	Options  []dbp.OptionDef
}

// Args is the parsed CLI invocation (--car / --mode / --line / --list).
type Args struct {
	CarQuery       string
	Mode           string // "" -> default retail-dza
	Line           string
	List           bool
	IncludeOptions bool // only when the client explicitly asks about options
}

type part struct {
	Label string   `json:"label"`
	Usd   *float64 `json:"usd,omitempty"`
	Dza   *float64 `json:"dza,omitempty"`
	Note  string   `json:"note,omitempty"`
}

type freighterUsed struct {
	ShippingLine          string  `json:"shippingLine"`
	LoadingPort           string  `json:"loadingPort"`
	DischargePort         string  `json:"dischargePort"`
	MyCostUsd             float64 `json:"myCostUsd"`
	FreightBenefitUsd     float64 `json:"freightBenefitUsd"`
	EffectiveContainerUsd float64 `json:"effectiveContainerUsd"`
	UnitsPerContainer     int     `json:"unitsPerContainer"`
	FreightPerUnitUsd     float64 `json:"freightPerUnitUsd"`
	Etd                   any     `json:"etd"`
	ValidTo               any     `json:"validTo"`
	Timestamp             string  `json:"timestamp"`
}

type clearanceUsed struct {
	CarId        any     `json:"carId"`
	Name         string  `json:"name"`
	ClearanceDza float64 `json:"clearanceDza"`
}

type inputPayload struct {
	WholesaleBenefitUsd   float64        `json:"wholesaleBenefitUsd"`
	RetailBenefitUsd      float64        `json:"retailBenefitUsd"`
	AllInRetailBenefitUsd *float64       `json:"allInRetailBenefitUsd,omitempty"`
	UsdToDza              float64        `json:"usdToDza"`
	RateTimestamp         any            `json:"rateTimestamp"`
	BenefitsTimestamp     any            `json:"benefitsTimestamp"`
	Freight               *freighterUsed `json:"freight,omitempty"`
	Clearance             *clearanceUsed `json:"clearance,omitempty"`
}

type staleEntry struct {
	File      string  `json:"file"`
	AgeDays   float64 `json:"ageDays"`
	Timestamp string  `json:"timestamp"`
}

type resolvedOption struct {
	Name     string `json:"name"`
	PriceUsd any    `json:"priceUsd"`
	Source   string `json:"source"`
}

type carInfo struct {
	CarId       string           `json:"carId"`
	Name        string           `json:"name"`
	Size        string           `json:"size"`
	Condition   string           `json:"condition"`
	CostUsd     float64          `json:"costUsd"`
	SupplierId  any              `json:"supplierId"`
	LoadingPort any              `json:"loadingPort"`
	Colors      []string         `json:"colors"`
	Options     []resolvedOption `json:"options"`
	Timestamp   any              `json:"timestamp"`
}

type pricePayload struct {
	Usd      any `json:"usd"`
	Dza      any `json:"dza"`
	DzaWords any `json:"dzaWords"`
}

type mainResult struct {
	Status    string       `json:"status"`
	Mode      string       `json:"mode"`
	Car       carInfo      `json:"car"`
	Inputs    inputPayload `json:"inputs"`
	Formula   string       `json:"formula"`
	Breakdown []part       `json:"breakdown"`
	Price     pricePayload `json:"price"`
	Stale     []staleEntry `json:"stale"`
	Warnings  []string     `json:"warnings"`
	Action    string       `json:"action,omitempty"`
}

type candidateCar struct {
	CarId   string  `json:"carId"`
	Name    string  `json:"name"`
	CostUsd float64 `json:"costUsd"`
}

type candidateClearance struct {
	CarId        any     `json:"carId"`
	Name         string  `json:"name"`
	ClearanceDza float64 `json:"clearanceDza"`
}

type freightValidity struct {
	Line    string `json:"line"`
	ValidTo string `json:"validTo"`
	Expired bool   `json:"expired"`
}

// ListResult / Usage / Missing / Ambiguous payloads.
type ListResult struct {
	Status string    `json:"status"`
	Note   string    `json:"note"`
	Modes  []string  `json:"modes"`
	Cars   []dbp.Car `json:"cars"`
}

type UsageResult struct {
	Status string   `json:"status"`
	Error  string   `json:"error"`
	Modes  []string `json:"modes"`
}

type MissingCarResult struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Hint   string `json:"hint"`
}

type AmbiguousCarResult struct {
	Status     string         `json:"status"`
	Error      string         `json:"error"`
	Candidates []candidateCar `json:"candidates"`
	Hint       string         `json:"hint"`
}

type MissingOptionResult struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Hint   string `json:"hint"`
}

type MissingFreightResult struct {
	Status  string            `json:"status"`
	Error   string            `json:"error"`
	ValidTo []freightValidity `json:"validTo"`
	Hint    string            `json:"hint"`
}

type MissingClearanceResult struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Note   any    `json:"note"`
	Hint   string `json:"hint"`
}

type AmbiguousClearanceResult struct {
	Status     string               `json:"status"`
	Error      string               `json:"error"`
	Candidates []candidateClearance `json:"candidates"`
}

// Result is the fully-formed payload to JSON-marshal.
type Result struct {
	Payload any
}

func jsRound(n float64, dec int) float64 {
	p := math.Pow(10, float64(dec))
	return math.Round(n*p) / p
}

func ageDays(iso *string) float64 {
	if iso == nil || *iso == "" {
		return math.Inf(1)
	}
	t, ok := timex.ParseIso(*iso)
	if !ok {
		return math.Inf(1)
	}
	return float64(time.Since(t).Milliseconds()) / dayMs
}

// daysUntil(dateOnly) — day-only string expiry, treated as end-of-day; missing -> +Inf.
func daysUntil(dateOnly any) float64 {
	s, ok := dateOnly.(string)
	if !ok || s == "" {
		return math.Inf(1)
	}
	t, err := time.Parse("2006-01-02T15:04:05Z07:00", s+"T23:59:59Z")
	if err != nil {
		return math.Inf(1)
	}
	return float64(t.Sub(time.Now()).Milliseconds()) / dayMs
}

func freightExpiry(f dbp.FreightQuote) any {
	if f.ValidTo != nil {
		return f.ValidTo
	}
	return f.Etd
}

func str(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Run executes the calculator and returns the payload struct. The CLI marshals
// with indent=2, exit 0 — `status` is the signal.
func Run(d *Data, a Args) *Result {
	mode := a.Mode
	if mode == "" {
		mode = "retail-dza"
	}

	// ---- --list / usage gate (src/quote.ts order: list branch runs first)
	if a.List || a.CarQuery == "" {
		if a.CarQuery != "" {
			return &Result{Payload: &UsageResult{
				Status: "USAGE",
				Error:  `Unknown flag for car "` + a.CarQuery + `"`,
				Modes:  Modes,
			}}
		}
		return &Result{Payload: &ListResult{Status: "LIST", Note: "Catalog (latest entry per carId).", Modes: Modes, Cars: d.Cars}}
	}

	if !isMode(mode) {
		return &Result{Payload: &UsageResult{Status: "USAGE", Error: `Unknown mode "` + mode + `"`, Modes: Modes}}
	}

	var stale []staleEntry
	warnings := []string{}

	usesRate := mode == "retail-dza" || mode == "allin-dza"
	usesBenefits := mode != "clearance"

	checkFresh := func(file string, ts *string, enforce bool) {
		d := ageDays(ts)
		if d > maxAgeDays {
			if enforce {
				t := "unknown"
				if ts != nil {
					t = *ts
				}
				stale = append(stale, staleEntry{File: file, AgeDays: jsRound(d, 1), Timestamp: t})
			} else {
				warnings = append(warnings, file+" last updated "+fmtInt(math.Round(d))+" days ago (advisory only).")
			}
		}
	}

	if usesBenefits {
		checkFresh("config (benefits)", d.Benefits.Timestamp, true)
	}
	if usesRate {
		checkFresh("config (rate)", d.Rate.Timestamp, true)
	}
	checkFresh("clearance", d.Custom.UpdatedAt, false)

	// ---- car match
	q := Norm(a.CarQuery)
	catalog := d.Cars
	var matches []dbp.Car
	for _, c := range catalog {
		id := Norm(c.CarId)
		name := Norm(c.Name)
		if id == q || name == q || strings.Contains(name, q) || strings.Contains(id, q) || strings.Contains(q, name) {
			matches = append(matches, c)
		}
	}

	if len(matches) == 0 {
		return &Result{Payload: &MissingCarResult{
			Status: "MISSING",
			Error:  `No car in cars matches "` + a.CarQuery + `" (FIVAN is read as LIVAN).`,
			Hint:   "Run with --list to see the catalog. Never quote a car that is not in it.",
		}}
	}

	matchedCars := matches
	if len(matches) > 1 {
		exact := matches[:0]
		for _, c := range matches {
			id := Norm(c.CarId)
			name := Norm(c.Name)
			if id == q || name == q {
				exact = append(exact, c)
			}
		}
		if len(exact) == 1 {
			matchedCars = exact
		}
	}
	if len(matchedCars) > 1 {
		cands := make([]candidateCar, 0, len(matchedCars))
		for _, c := range matchedCars {
			cands = append(cands, candidateCar{CarId: c.CarId, Name: c.Name, CostUsd: c.CostUsd})
		}
		return &Result{Payload: &AmbiguousCarResult{
			Status:     "AMBIGUOUS",
			Error:      `"` + a.CarQuery + `" matches ` + fmtInt(float64(len(matchedCars))) + " cars.",
			Candidates: cands,
			Hint:       "MAN and AUTO are different cars — re-run with the exact name.",
		}}
	}

	car := matchedCars[0]
	checkFresh("cars", car.Timestamp, true)

	// ---- options
	// Rule: options are NEVER priced by default — they are added to the quote
	// only when the client explicitly asks about them (--options). This mirrors
	// the operating rule: never quote a spec the client has not asked for.
	var resolvedOptions []resolvedOption
	optionsUsd := 0.0
	if a.IncludeOptions && mode != "clearance" {
		defaultOptionPrice := map[string]float64{}
		for _, o := range d.Options {
			defaultOptionPrice[Norm(o.Name)] = o.PriceUsd
		}
		var missingOptions []string
		for _, opt := range car.Options {
			var fromCar *float64
			if opt.PriceUsd != nil {
				fromCar = opt.PriceUsd
			}
			fromDefault, hasDefault := defaultOptionPrice[Norm(opt.Name)]
			var price *float64
			if fromCar != nil {
				price = fromCar
			} else if hasDefault {
				price = &fromDefault
			}
			if price == nil {
				missingOptions = append(missingOptions, opt.Name)
				resolvedOptions = append(resolvedOptions, resolvedOption{Name: opt.Name, PriceUsd: nil, Source: "missing"})
			} else {
				optionsUsd += *price
				src := "default"
				if fromCar != nil {
					src = "car"
				}
				resolvedOptions = append(resolvedOptions, resolvedOption{Name: opt.Name, PriceUsd: *price, Source: src})
			}
		}
		if len(missingOptions) > 0 {
			return &Result{Payload: &MissingOptionResult{
				Status: "MISSING",
				Error:  car.Name + " has option(s) with no price: " + strings.Join(missingOptions, ", ") + ".",
				Hint:   "Add a default in options or a per-car priceUsd in cars. Never guess an option price.",
			}}
		}
	}

	// ---- freight
	unitsPerContainer := 4
	if car.Size == "big" {
		unitsPerContainer = 3
	}
	needFreight := mode != "wholesale-fob" && mode != "retail-usd-fob" && mode != "clearance"

	var freightPerUnitUsd *float64
	var freightUsed *freighterUsed

	if needFreight {
		var badValidity []dbp.FreightQuote
		for _, f := range d.Freight.Freight {
			vts, ok1 := f.ValidTo.(string)
			ets, ok2 := f.Etd.(string)
			if ok1 && ok2 {
				t1, e1 := time.Parse("2006-01-02T15:04:05Z07:00", vts+"T23:59:59Z")
				t2, e2 := time.Parse("2006-01-02T15:04:05Z07:00", ets+"T23:59:59Z")
				if e1 == nil && e2 == nil && t1.After(t2) {
					badValidity = append(badValidity, f)
				}
			}
		}
		if len(badValidity) > 0 {
			var parts []string
			for _, f := range badValidity {
				parts = append(parts, f.ShippingLine+" (validTo "+str(f.ValidTo)+" > etd "+str(f.Etd)+")")
			}
			warnings = append(warnings, "freight: validTo must be <= etd — "+strings.Join(parts, ", ")+".")
		}

		var candidates []dbp.FreightQuote
		var expired []dbp.FreightQuote
		for _, f := range d.Freight.Freight {
			if daysUntil(freightExpiry(f)) >= 0 {
				candidates = append(candidates, f)
			} else {
				expired = append(expired, f)
			}
		}
		if len(expired) > 0 {
			parts := make([]string, 0, len(expired))
			for _, e := range expired {
				exp := freightExpiry(e)
				expS := "?"
				if exp != nil {
					expS = str(exp)
				}
				parts = append(parts, e.ShippingLine+" (validTo "+expS+")")
			}
			warnings = append(warnings, "Ignored expired freight quotes: "+strings.Join(parts, ", ")+".")
		}

		if a.Line != "" {
			wanted := strings.ToUpper(a.Line)
			kept := candidates[:0]
			for _, f := range candidates {
				if strings.ToUpper(f.ShippingLine) == wanted {
					kept = append(kept, f)
				}
			}
			candidates = kept
		}

		if len(candidates) == 0 {
			var list []freightValidity
			for _, f := range d.Freight.Freight {
				exp := freightExpiry(f)
				expS := ""
				if exp != nil {
					expS = str(exp)
				}
				list = append(list, freightValidity{Line: f.ShippingLine, ValidTo: expS, Expired: daysUntil(freightExpiry(f)) < 0})
			}
			msg := "No unexpired freight quote in freight."
			if a.Line != "" {
				msg = `No unexpired freight quote for line "` + a.Line + `".`
			}
			return &Result{Payload: &MissingFreightResult{
				Status:  "MISSING",
				Error:   msg,
				ValidTo: list,
				Hint:    "Ask the boss for a current freight cost. Do not guess one.",
			}}
		}

		sort.SliceStable(candidates, func(i, j int) bool {
			ti, oki := timex.ParseIso(candidates[i].Timestamp)
			tj, okj := timex.ParseIso(candidates[j].Timestamp)
			if !oki || !okj {
				return false
			}
			return ti.After(tj)
		})
		chosen := candidates[0]
		if len(candidates) > 1 {
			var others []string
			for _, c := range candidates[1:] {
				others = append(others, c.ShippingLine+" "+formatNum(c.Container40ftUsd))
			}
			warnings = append(warnings,
				fmtInt(float64(len(candidates)))+" valid freight quotes; used the most recent ("+chosen.ShippingLine+", "+chosen.Timestamp+"). "+
					"Others: "+strings.Join(others, ", ")+". "+
					"Pass --line to pick one.")
		}

		checkFresh("freight", &chosen.Timestamp, true)
		containerUsd := chosen.Container40ftUsd + freightBenefitUsd
		fpu := containerUsd / float64(unitsPerContainer)
		freightPerUnitUsd = &fpu
		freightUsed = &freighterUsed{
			ShippingLine:          chosen.ShippingLine,
			LoadingPort:           chosen.LoadingPort,
			DischargePort:         chosen.DischargePort,
			MyCostUsd:             chosen.Container40ftUsd,
			FreightBenefitUsd:     freightBenefitUsd,
			EffectiveContainerUsd: containerUsd,
			UnitsPerContainer:     unitsPerContainer,
			FreightPerUnitUsd:     jsRound(fpu, 2),
			Etd:                   chosen.Etd,
			ValidTo:               chosen.ValidTo,
			Timestamp:             chosen.Timestamp,
		}
	}

	// ---- clearance (allin / clearance)
	var clearanceEntry *dbp.ClearanceEntry
	var pool []dbp.ClearanceEntry
	if mode == "allin-dza" || mode == "clearance" {
		var direct, byAlias []dbp.ClearanceEntry
		for _, e := range d.Custom.Entries {
			if e.CarId == car.CarId {
				direct = append(direct, e)
			}
		}
		for _, e := range d.Custom.Entries {
			for _, al := range e.Aliases {
				na := Norm(al)
				if na == q || strings.Contains(na, q) || strings.Contains(q, na) {
					byAlias = append(byAlias, e)
					break
				}
			}
		}
		pool = direct
		if len(pool) == 0 {
			pool = byAlias
		}
		if len(pool) == 0 {
			return &Result{Payload: &MissingClearanceResult{
				Status: "MISSING",
				Error:  "clearance has no clearance entry for " + car.CarId + ".",
				Hint:   "carId: null entries are unmatched — ask the boss. Never invent a clearance figure.",
			}}
		}
		if len(pool) > 1 {
			cands := make([]candidateClearance, 0, len(pool))
			for _, e := range pool {
				cands = append(cands, candidateClearance{CarId: e.CarId, Name: e.Name, ClearanceDza: e.ClearanceDza})
			}
			return &Result{Payload: &AmbiguousClearanceResult{
				Status:     "AMBIGUOUS",
				Error:      "clearance matched " + fmtInt(float64(len(pool))) + " entries for " + car.CarId + ".",
				Candidates: cands,
			}}
		}
		clearanceEntry = &pool[0]
		if clearanceEntry.CarId == nil {
			return &Result{Payload: &MissingClearanceResult{
				Status: "MISSING",
				Error:  `clearance entry "` + clearanceEntry.Name + `" is unmatched (carId: null).`,
				Note:   clearanceEntry.Note,
				Hint:   "Ask the boss which catalog car this is before quoting.",
			}}
		}
	}

	// ---- compute
	var parts []part
	wholesaleBenefit := d.Benefits.WholesaleBenefitUsd
	retailBenefit := d.Benefits.RetailBenefitUsd
	allInBenefit := d.Custom.AllInRetailBenefitUsd
	rate := d.Rate.UsdToDza

	addUsd := func(label string, usd float64, note string) {
		u := jsRound(usd, 2)
		p := part{Label: label, Usd: &u}
		if note != "" {
			p.Note = note
		}
		parts = append(parts, p)
	}
	addDza := func(label string, dza float64, note string) {
		z := jsRound(dza, 0)
		p := part{Label: label, Dza: &z}
		if note != "" {
			p.Note = note
		}
		parts = append(parts, p)
	}

	costNote := car.CarId
	addUsd("supplier cost (cars)", car.CostUsd, costNote)

	formula := ""
	priceUsd := math.NaN()
	priceDza := math.NaN()

	freightNote := 0.0
	if freightPerUnitUsd != nil {
		freightNote = jsRound(*freightPerUnitUsd, 2)
	}
	const OPTIONS_LABEL = "car options (cars / options)"
	addOptions := func() {
		if optionsUsd != 0 {
			var names []string
			for _, o := range resolvedOptions {
				names = append(names, o.Name)
			}
			addUsd(OPTIONS_LABEL, optionsUsd, strings.Join(names, ", "))
		}
	}

	freightLine := ""
	if freightUsed != nil {
		freightLine = freightUsed.ShippingLine
	}

	switch mode {
	case "wholesale-fob":
		addUsd("wholesale benefit (config)", wholesaleBenefit, "")
		addOptions()
		priceUsd = car.CostUsd + wholesaleBenefit + optionsUsd
		formula = "costUsd + wholesaleBenefitUsd + optionsUsd   (FOB: goods at loading port, NO sea freight)"
	case "wholesale-cfr":
		addUsd("wholesale benefit (config)", wholesaleBenefit, "")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = car.CostUsd + wholesaleBenefit + freightNote + optionsUsd
		formula = "costUsd + wholesaleBenefitUsd + freightPerUnitUsd + optionsUsd   (CFR / landed-style)"
	case "retail-dza":
		addUsd("retail benefit (config)", retailBenefit, "")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = car.CostUsd + retailBenefit + freightNote + optionsUsd
		addUsd("subtotal (USD)", priceUsd, "")
		addDza("× rate "+formatUsd(rate)+" (config)", priceUsd*rate, "")
		priceDza = priceUsd * rate
		formula = "(costUsd + retailBenefitUsd + freightPerUnitUsd + optionsUsd) × usdToDza"
	case "retail-usd":
		addUsd("retail benefit (config)", retailBenefit, "")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = car.CostUsd + retailBenefit + freightNote + optionsUsd
		formula = "costUsd + retailBenefitUsd + freightPerUnitUsd + optionsUsd   (only when the client explicitly asks USD)"
	case "retail-usd-fob":
		addUsd("retail benefit (config)", retailBenefit, "")
		addOptions()
		priceUsd = car.CostUsd + retailBenefit + optionsUsd
		formula = "costUsd + retailBenefitUsd + optionsUsd   (FOB retail, no freight — only if explicitly asked)"
	case "allin-dza":
		addUsd("all-in benefit (clearance)", allInBenefit, "not the usual 700 retail benefit")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = car.CostUsd + allInBenefit + freightNote + optionsUsd
		addUsd("subtotal (USD)", priceUsd, "")
		converted := priceUsd * rate
		addDza("× rate "+formatUsd(rate)+" (config)", converted, "")
		addDza("customs clearance (clearanceDza)", clearanceEntry.ClearanceDza, clearanceEntry.Name)
		priceDza = converted + clearanceEntry.ClearanceDza
		formula = "(costUsd + allInRetailBenefitUsd + freightPerUnitUsd + optionsUsd) × usdToDza + clearanceDza   (clearance added AFTER conversion)"
	case "clearance":
		addDza("customs clearance (clearanceDza)", clearanceEntry.ClearanceDza, clearanceEntry.Name)
		priceDza = clearanceEntry.ClearanceDza
		formula = "clearanceDza as stored in clearance"
	}

	if mode == "wholesale-fob" || mode == "wholesale-cfr" {
		warnings = append(warnings, "Wholesale = full containers only, USD only, never DZA.")
	}
	if mode == "retail-dza" && !math.IsNaN(priceDza) {
		if d.Custom.TvnIncluded {
			warnings = append(warnings, "TVN is NOT included (see clearance).")
		} else {
			warnings = append(warnings, "TVN is NOT included (clearance tvnIncluded: false).")
		}
	}
	if mode == "allin-dza" {
		warnings = append(warnings, "TVN is NOT included — say so if the client asks.")
	}

	status := "OK"
	if len(stale) > 0 {
		status = "STALE"
	}

	colors := car.Colors
	if colors == nil {
		colors = []string{}
	}
	opts := resolvedOptions
	if opts == nil {
		opts = []resolvedOption{}
	}
	staleList := stale
	if staleList == nil {
		staleList = []staleEntry{}
	}

	res := &mainResult{
		Status: status,
		Mode:   mode,
		Car: carInfo{
			CarId:       car.CarId,
			Name:        car.Name,
			Size:        car.Size,
			Condition:   car.Condition,
			CostUsd:     car.CostUsd,
			SupplierId:  car.SupplierId,
			LoadingPort: car.LoadingPort,
			Colors:      colors,
			Options:     opts,
			Timestamp:   car.Timestamp,
		},
		Inputs: inputPayload{
			WholesaleBenefitUsd: wholesaleBenefit,
			RetailBenefitUsd:    retailBenefit,
			UsdToDza:            rate,
			RateTimestamp:       d.Rate.Timestamp,
			BenefitsTimestamp:   d.Benefits.Timestamp,
			Freight:             freightUsed,
			Clearance:           clearanceUsedFor(clearanceEntry),
		},
		Formula:   formula,
		Breakdown: parts,
		Price: pricePayload{
			Usd:      nullableNum(priceUsd),
			Dza:      nullableNum(priceDza),
			DzaWords: dzaWords(priceDza),
		},
		Stale:    staleList,
		Warnings: warnings,
	}
	if mode == "allin-dza" {
		a := allInBenefit
		res.Inputs.AllInRetailBenefitUsd = &a
	}
	if status == "STALE" {
		res.Action = "Do NOT quote these numbers. Ask the boss (سأطلب من المدير وأعود إليك + AI line) and log it in ai_questions.json."
	}
	return &Result{Payload: res}
}

func clearanceUsedFor(e *dbp.ClearanceEntry) *clearanceUsed {
	if e == nil {
		return nil
	}
	return &clearanceUsed{CarId: e.CarId, Name: e.Name, ClearanceDza: e.ClearanceDza}
}

// nullableNum mirrors Node's `priceUsd != null ? round(...) : null`.
func nullableNum(n float64) any {
	if math.IsNaN(n) {
		return nil
	}
	return jsRound(n, 2)
}

func dzaWords(priceDza float64) any {
	if math.IsNaN(priceDza) {
		return nil
	}
	dz := jsRound(priceDza, 0)
	cent := dz * 100
	mil1 := jsRound(cent/1e6, 1)
	mil2 := jsRound(dz/1e6, 2)
	return formatNum(dz) + " DZD = " + formatNum(cent) + " centimes ≈ " + formatNum(mil1) + " مليون centimes (" + formatNum(mil2) + "M DZD)"
}

func fmtInt(f float64) string { return formatNum(f) }

func formatUsd(n float64) string { return formatNum(n) }

func formatNum(n float64) string {
	// Shortest round-trip repr (JS number toString) but avoid sci notation for
	// these magnitudes; big strings need plain decimal.
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "0"
	}
	switch {
	case n == math.Trunc(n) && math.Abs(n) < 1e21:
		return int64Str(n)
	default:
		return fmt.Sprintf("%v", n)
	}
}

func int64Str(n float64) string {
	i := int64(n)
	s := ""
	neg := i < 0
	u := uint64(i)
	if neg {
		u = uint64(-i)
	}
	if u == 0 {
		return "0"
	}
	for u > 0 {
		s = string(rune('0'+u%10)) + s
		u /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
