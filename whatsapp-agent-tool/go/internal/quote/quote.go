// Package quote is the Go twin of src/quote.ts — the deterministic pricing
// calculator. The model never does §13 arithmetic itself; the CLI quotes these
// numbers verbatim. Statuses: OK / MISSING / AMBIGUOUS / BLOCKED / USAGE / LIST.
//
// Stale or expired inputs are NOT a hard stop: the calculator always quotes the
// best available figure (cheapest applicable shipping line) and records the age
// in `stale` / `warnings` / per-line `expired` so the reply can mention it
// (operating rule, 2026-10-14).
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

// Args is the parsed CLI invocation (--car / --mode / --line / --port / --color / --list).
type Args struct {
	CarQuery       string
	Mode           string // "" -> default retail-dza
	Line           string
	Port           string // requested discharge port; the cheapest valid line to it wins
	Color          string // requested colour; "" -> the cheapest colour
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

// linePrice is one shipping line's version of the quote: the freight (and thus
// the price) differs per line. The cheapest applicable line is the headline
// price; every applicable line gets an entry (expired ones flagged, not hidden).
type linePrice struct {
	ShippingLine      string  `json:"shippingLine"`
	LoadingPort       string  `json:"loadingPort,omitempty"`
	DischargePort     string  `json:"dischargePort,omitempty"`
	Container40ftUsd  float64 `json:"container40ftUsd"`
	FreightPerUnitUsd float64 `json:"freightPerUnitUsd"`
	ValidTo           any     `json:"validTo"`
	Etd               any     `json:"etd"`
	QuotedAt          string  `json:"quotedAt,omitempty"`
	Expired           bool    `json:"expired"`
	Stale             bool    `json:"stale"`
	Chosen            bool    `json:"chosen"`
	Usd               any     `json:"usd,omitempty"`
	Dza               any     `json:"dza,omitempty"`
	DzaWords          any     `json:"dzaWords,omitempty"`
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
	CarId       string             `json:"carId"`
	Name        string             `json:"name"`
	Size        string             `json:"size"`
	Condition   string             `json:"condition"`
	CostUsd     float64            `json:"costUsd"`
	SupplierId  any                `json:"supplierId"`
	LoadingPort any                `json:"loadingPort"`
	Colors      []string           `json:"colors"`
	ColorCosts  map[string]float64 `json:"colorCosts,omitempty"` // per-colour buy cost override (USD)
	Options     []resolvedOption   `json:"options"`
	Timestamp   any                `json:"timestamp"`
}

type pricePayload struct {
	Usd      any `json:"usd"`
	Dza      any `json:"dza"`
	DzaWords any `json:"dzaWords"`
}

type mainResult struct {
	Status    string       `json:"status"`
	Mode      string       `json:"mode"`
	Color           string       `json:"color,omitempty"`           // colour this price is for
	ColorIsCheapest *bool        `json:"colorIsCheapest,omitempty"` // true when Color is the cheapest colour
	Car       carInfo      `json:"car"`
	Inputs    inputPayload `json:"inputs"`
	Formula   string       `json:"formula"`
	Breakdown []part       `json:"breakdown"`
	Price     pricePayload `json:"price"`
	Lines     []linePrice  `json:"lines,omitempty"`
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

// MissingColorResult is returned when --color names a colour the car does not
// list. Never guess a colour price — offer the cheapest or a listed colour.
type MissingColorResult struct {
	Status string   `json:"status"`
	Error  string   `json:"error"`
	Colors []string `json:"colors"`
	Hint   string   `json:"hint"`
}

type AmbiguousCarResult struct {
	Status     string         `json:"status"`
	Error      string         `json:"error"`
	Candidates []candidateCar `json:"candidates"`
	Hint       string         `json:"hint"`
}

// BlockedCarResult refuses a quote for a car whose cars.notes carry the
// operator's block marker ("do not quote" / "BLOCKED"). Blocking happens
// before any price is computed — a blocked price must never reach a client.
type BlockedCarResult struct {
	Status string `json:"status"`
	Car    string `json:"car"`
	Notes  string `json:"notes"`
	Error  string `json:"error"`
	Action string `json:"action"`
}

// colorCost returns a car's buy cost for a colour: the per-colour override in
// car_colors when present, else the car's base cost.
func colorCost(car dbp.Car, color string) float64 {
	if c, ok := car.ColorCosts[color]; ok {
		return c
	}
	return car.CostUsd
}

// IsBlockedNote reports whether a car's cars.notes mark it as blocked. The
// operator stamps blocked cars with "do not quote" and/or an uppercase
// "BLOCKED" marker (e.g. "KENT BLOCKED 2026-10-16 - not available, do not
// quote"). A lowercase "kent blocked … reverted to jean price" note is NOT a
// block — those cars were reverted to another supplier and stay quotable.
func IsBlockedNote(notes any) bool {
	s, ok := notes.(string)
	if !ok || s == "" {
		return false
	}
	if strings.Contains(strings.ToLower(s), "do not quote") {
		return true
	}
	return strings.Contains(s, "BLOCKED")
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

// normPort upper-cases and strips everything but A–Z so "ALGIERS / ANNABA
// (ALGERIA)" and "algiers" compare by substring.
func normPort(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r)
		}
	}
	return b.String()
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

	// Data older than the weekly window (rate / benefits / cars / freight) is no
	// longer a hard stop — we still quote the best available figure and record
	// the age so the reply can mention it (operating rule, 2026-10-14).
	checkFresh := func(file string, ts *string, _ bool) {
		d := ageDays(ts)
		if d > maxAgeDays {
			t := "unknown"
			if ts != nil {
				t = *ts
			}
			stale = append(stale, staleEntry{File: file, AgeDays: jsRound(d, 1), Timestamp: t})
			warnings = append(warnings, file+" is "+fmtInt(math.Round(d))+" days old (past the 7-day window) — quote it but say it may have changed.")
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

	// ---- blocked check (STRICT) — before any price is computed. A car whose
	// cars.notes carry the operator's block marker ("do not quote" / uppercase
	// "BLOCKED") is refused outright: its price must never be quoted or shared.
	if note, ok := car.Notes.(string); ok && IsBlockedNote(car.Notes) {
		return &Result{Payload: &BlockedCarResult{
			Status: "BLOCKED",
			Car:    car.Name,
			Notes:  note,
			Error:  car.Name + " is blocked by the operator — its price must NOT be quoted or shared with anyone.",
			Action: "Never quote a blocked car. Offer a non-blocked alternative from the catalog, or say the car is not available.",
		}}
	}

	checkFresh("cars", car.Timestamp, true)

	// ---- colour: pick the requested colour, else the cheapest one.
	// cars.cost_usd is the base cost; a colour may override it (car_colors.cost_usd).
	// Default (fast answer, client-reply.md §3.2): quote the CHEAPEST colour;
	// --color <name> quotes a specific one.
	baseCost := car.CostUsd
	effectiveCost := baseCost
	color := ""
	var colorIsCheapest *bool
	if len(car.Colors) > 0 {
		cheapest, cheapestCost := car.Colors[0], colorCost(car, car.Colors[0])
		for _, c := range car.Colors[1:] {
			if cc := colorCost(car, c); cc < cheapestCost {
				cheapest, cheapestCost = c, cc
			}
		}
		if a.Color == "" {
			t := true
			color, effectiveCost, colorIsCheapest = cheapest, cheapestCost, &t
		} else {
			found := ""
			for _, c := range car.Colors {
				if Norm(c) == Norm(a.Color) {
					found = c
					break
				}
			}
			if found == "" {
				return &Result{Payload: &MissingColorResult{
					Status: "MISSING",
					Error:  car.Name + ` has no colour "` + a.Color + `".`,
					Colors: car.Colors,
					Hint:   "Quote the cheapest colour by default; a colour not listed here has no price — never invent one.",
				}}
			}
			color, effectiveCost = found, colorCost(car, found)
			c := effectiveCost == cheapestCost
			colorIsCheapest = &c
		}
	} else if a.Color != "" {
		// The car lists no colours; accept the requested one at the base cost.
		color = a.Color
	}

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
	var freightApplicable []dbp.FreightQuote
	var chosenLine, chosenTs string

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

		// Destination port defaults to ALGIERS when none is named (operating
		// rule: quote it and say so in the reply). A named port always wins.
		portFilter := a.Port
		if portFilter == "" {
			portFilter = "ALGIERS"
			warnings = append(warnings, "No destination port given; defaulted to ALGIERS.")
		}

		// Keep every quote matching the requested line (if any) and port —
		// expired ones included, so each shipping line still gets a price.
		for _, f := range d.Freight.Freight {
			if a.Line != "" && !strings.EqualFold(strings.TrimSpace(f.ShippingLine), strings.TrimSpace(a.Line)) {
				continue
			}
			if portFilter != "" && !strings.Contains(normPort(f.DischargePort), normPort(portFilter)) {
				continue
			}
			freightApplicable = append(freightApplicable, f)
		}

		if len(freightApplicable) == 0 {
			var list []freightValidity
			for _, f := range d.Freight.Freight {
				exp := freightExpiry(f)
				expS := ""
				if exp != nil {
					expS = str(exp)
				}
				list = append(list, freightValidity{Line: f.ShippingLine, ValidTo: expS, Expired: daysUntil(freightExpiry(f)) < 0})
			}
			msg := "No freight quote in freight."
			if a.Line != "" {
				msg = `No freight quote for line "` + a.Line + `".`
			} else if portFilter != "" {
				msg = `No freight quote to port "` + portFilter + `".`
			}
			return &Result{Payload: &MissingFreightResult{
				Status:  "MISSING",
				Error:   msg,
				ValidTo: list,
				Hint:    "Ask the boss for a freight cost for this port. Do not guess one.",
			}}
		}

		// Sort cheapest-first; a newer quote only breaks a price tie.
		sortFreight := func(list []dbp.FreightQuote) {
			sort.SliceStable(list, func(i, j int) bool {
				if list[i].Container40ftUsd != list[j].Container40ftUsd {
					return list[i].Container40ftUsd < list[j].Container40ftUsd
				}
				ti, oki := timex.ParseIso(list[i].Timestamp)
				tj, okj := timex.ParseIso(list[j].Timestamp)
				if !oki || !okj {
					return false
				}
				return ti.After(tj)
			})
		}

		var valid, expiredQ []dbp.FreightQuote
		for _, f := range freightApplicable {
			if daysUntil(freightExpiry(f)) >= 0 {
				valid = append(valid, f)
			} else {
				expiredQ = append(expiredQ, f)
			}
		}
		if len(expiredQ) > 0 {
			parts := make([]string, 0, len(expiredQ))
			for _, e := range expiredQ {
				exp := freightExpiry(e)
				expS := "?"
				if exp != nil {
					expS = str(exp)
				}
				parts = append(parts, e.ShippingLine+" (validTo "+expS+")")
			}
			warnings = append(warnings, "Expired freight quote(s) for "+portFilter+": "+strings.Join(parts, ", ")+".")
		}

		// Default policy: the cheapest line wins (never ask the client / boss
		// which line). If every applicable line is already expired we still use
		// the cheapest one and say so — never block the quote (§13, 2026-10-14).
		allExpired := len(valid) == 0
		pool := freightApplicable
		if !allExpired {
			pool = valid
		}
		sortFreight(pool)
		chosen := pool[0]
		if allExpired {
			exp := freightExpiry(chosen)
			expS := "unknown"
			if exp != nil {
				expS = str(exp)
			}
			warnings = append(warnings, "All "+portFilter+" freight quotes are expired; using the cheapest anyway ("+
				chosen.ShippingLine+", validTo "+expS+") — the freight may have changed.")
		} else if len(pool) > 1 {
			var others []string
			for _, c := range pool[1:] {
				others = append(others, c.ShippingLine+" "+formatNum(c.Container40ftUsd))
			}
			warnings = append(warnings,
				fmtInt(float64(len(pool)))+" applicable freight quotes; used the cheapest ("+chosen.ShippingLine+", "+formatNum(chosen.Container40ftUsd)+"). "+
					"Others: "+strings.Join(others, ", ")+". See `lines` for a price per line; pass --line to override.")
		}

		chosenLine, chosenTs = chosen.ShippingLine, chosen.Timestamp
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

	// computePrice is the single source of truth for the price at a given
	// per-unit freight cost — reused for the headline price and for every
	// shipping line's price.
	computePrice := func(freightNote float64) (usd, dza float64) {
		switch mode {
		case "wholesale-fob":
			return effectiveCost + wholesaleBenefit + optionsUsd, math.NaN()
		case "wholesale-cfr":
			return effectiveCost + wholesaleBenefit + freightNote + optionsUsd, math.NaN()
		case "retail-dza":
			u := effectiveCost + retailBenefit + freightNote + optionsUsd
			return u, u * rate
		case "retail-usd":
			return effectiveCost + retailBenefit + freightNote + optionsUsd, math.NaN()
		case "retail-usd-fob":
			return effectiveCost + retailBenefit + optionsUsd, math.NaN()
		case "allin-dza":
			u := effectiveCost + allInBenefit + freightNote + optionsUsd
			z := u * rate
			if clearanceEntry != nil {
				z += clearanceEntry.ClearanceDza
			}
			return u, z
		case "clearance":
			if clearanceEntry != nil {
				return math.NaN(), clearanceEntry.ClearanceDza
			}
			return math.NaN(), math.NaN()
		}
		return math.NaN(), math.NaN()
	}

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
	if effectiveCost != baseCost {
		costNote = car.CarId + " · " + color
	}
	addUsd("supplier cost (cars)", effectiveCost, costNote)

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
		priceUsd = effectiveCost + wholesaleBenefit + optionsUsd
		formula = "costUsd + wholesaleBenefitUsd + optionsUsd   (FOB: goods at loading port, NO sea freight)"
	case "wholesale-cfr":
		addUsd("wholesale benefit (config)", wholesaleBenefit, "")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = effectiveCost + wholesaleBenefit + freightNote + optionsUsd
		formula = "costUsd + wholesaleBenefitUsd + freightPerUnitUsd + optionsUsd   (CFR / landed-style)"
	case "retail-dza":
		addUsd("retail benefit (config)", retailBenefit, "")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = effectiveCost + retailBenefit + freightNote + optionsUsd
		addUsd("subtotal (USD)", priceUsd, "")
		addDza("× rate "+formatUsd(rate)+" (config)", priceUsd*rate, "")
		priceDza = priceUsd * rate
		formula = "(costUsd + retailBenefitUsd + freightPerUnitUsd + optionsUsd) × usdToDza"
	case "retail-usd":
		addUsd("retail benefit (config)", retailBenefit, "")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = effectiveCost + retailBenefit + freightNote + optionsUsd
		formula = "costUsd + retailBenefitUsd + freightPerUnitUsd + optionsUsd   (only when the client explicitly asks USD)"
	case "retail-usd-fob":
		addUsd("retail benefit (config)", retailBenefit, "")
		addOptions()
		priceUsd = effectiveCost + retailBenefit + optionsUsd
		formula = "costUsd + retailBenefitUsd + optionsUsd   (FOB retail, no freight — only if explicitly asked)"
	case "allin-dza":
		addUsd("all-in benefit (clearance)", allInBenefit, "not the usual 700 retail benefit")
		addUsd("freight per unit ((my cost + 300) / units)", freightNote, freightLine)
		addOptions()
		priceUsd = effectiveCost + allInBenefit + freightNote + optionsUsd
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

	priceUsd, priceDza = computePrice(freightNote)

	// One price per applicable shipping line. The cheapest line is the
	// headline `price` above; the others are shown so the operator can compare
	// lines (and see which are expired / how old the quote is).
	var lines []linePrice
	if needFreight && len(freightApplicable) > 0 {
		sorted := append([]dbp.FreightQuote(nil), freightApplicable...)
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].Container40ftUsd != sorted[j].Container40ftUsd {
				return sorted[i].Container40ftUsd < sorted[j].Container40ftUsd
			}
			ti, oki := timex.ParseIso(sorted[i].Timestamp)
			tj, okj := timex.ParseIso(sorted[j].Timestamp)
			if !oki || !okj {
				return false
			}
			return ti.After(tj)
		})
		for _, q := range sorted {
			cu := q.Container40ftUsd + freightBenefitUsd
			qfpu := jsRound(cu/float64(unitsPerContainer), 2)
			u, z := computePrice(qfpu)
			lp := linePrice{
				ShippingLine:      q.ShippingLine,
				LoadingPort:       q.LoadingPort,
				DischargePort:     q.DischargePort,
				Container40ftUsd:  q.Container40ftUsd,
				FreightPerUnitUsd: qfpu,
				ValidTo:           q.ValidTo,
				Etd:               q.Etd,
				QuotedAt:          q.Timestamp,
				Expired:           daysUntil(freightExpiry(q)) < 0,
				Stale:             ageDays(&q.Timestamp) > maxAgeDays,
				Chosen:            q.ShippingLine == chosenLine && q.Timestamp == chosenTs,
				Usd:               nullableNum(u),
				Dza:               nullableNum(z),
			}
			if !math.IsNaN(z) {
				lp.DzaWords = dzaWords(z)
			}
			lines = append(lines, lp)
		}
	}

	if mode == "wholesale-fob" || mode == "wholesale-cfr" {
		warnings = append(warnings, "Wholesale = full containers only, USD only, never DZA.")
	}
	// The TVN note belongs ONLY to a price that already carries customs
	// clearance (all-in): a price without clearance excludes TVN as well, so
	// repeating the line there is pure noise (client-reply.md §3.6).
	if mode == "allin-dza" {
		if d.Custom.TvnIncluded {
			warnings = append(warnings, "TVN IS included (customs tvnIncluded: true).")
		} else {
			warnings = append(warnings, "TVN is NOT included — the price does include clearance; say so if the client asks.")
		}
	}

	// Staleness / expiry is never a hard stop (operating rule, 2026-10-14): we
	// quote the best available figure and flag it. `stale` lists what is old;
	// the caller should mention it but still give the number.
	status := "OK"

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
		Status:          status,
		Mode:            mode,
		Color:           color,
		ColorIsCheapest: colorIsCheapest,
		Car: carInfo{
			CarId:       car.CarId,
			Name:        car.Name,
			Size:        car.Size,
			Condition:   car.Condition,
			CostUsd:     car.CostUsd,
			SupplierId:  car.SupplierId,
			LoadingPort: car.LoadingPort,
			Colors:      colors,
			ColorCosts:  car.ColorCosts,
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
		Lines:    lines,
		Stale:    staleList,
		Warnings: warnings,
	}
	if mode == "allin-dza" {
		a := allInBenefit
		res.Inputs.AllInRetailBenefitUsd = &a
	}
	if len(stale) > 0 || hasExpiredLine(lines) {
		res.Action = "Quote the price, but mention the data is old/expired (may have changed) and offer to confirm with the team."
	}
	return &Result{Payload: res}
}

func clearanceUsedFor(e *dbp.ClearanceEntry) *clearanceUsed {
	if e == nil {
		return nil
	}
	return &clearanceUsed{CarId: e.CarId, Name: e.Name, ClearanceDza: e.ClearanceDza}
}

func hasExpiredLine(lines []linePrice) bool {
	for _, l := range lines {
		if l.Expired {
			return true
		}
	}
	return false
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
