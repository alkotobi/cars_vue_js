// Freight listing for general "how much is shipping to X" questions. The
// calculator (Run) only exposes freight as part of a car quote; this file adds
// a deterministic read that answers the shipping question on its own — one row
// per applicable shipping line to a port, cheapest first, with the per-car
// share. No model arithmetic: every figure is computed here.
package quote

import (
	"sort"
	"strings"

	dbp "whatsapp-agent-tool/internal/db/pricing"
	"whatsapp-agent-tool/internal/timex"
)

// FreightRate is one shipping line's best quote to a discharge port.
type FreightRate struct {
	DischargePort      string  `json:"dischargePort"`
	ShippingLine       string  `json:"shippingLine"`
	LoadingPort        string  `json:"loadingPort"`
	Container40ftUsd   float64 `json:"container40ftUsd"`   // my cost, 40ft container
	ContainerChargeUsd float64 `json:"containerChargeUsd"` // cost + freight benefit (what the price carries)
	PerCarBigUsd       float64 `json:"perCarBigUsd"`       // charge / 3 (big car)
	PerCarSmallUsd     float64 `json:"perCarSmallUsd"`     // charge / 4 (small car)
	Etd                any     `json:"etd,omitempty"`
	ValidFrom          any     `json:"validFrom,omitempty"`
	ValidTo            any     `json:"validTo,omitempty"`
	FreeTimeDays       any     `json:"freeTimeDays,omitempty"`
	Expired            bool    `json:"expired"`
	Timestamp          string  `json:"timestamp"`
	Notes              any     `json:"notes,omitempty"`
}

// FreightListResult is the payload for `whatsapp-tool freight`.
type FreightListResult struct {
	Status string        `json:"status"`
	Note   string        `json:"note"`
	Port   string        `json:"port,omitempty"`
	Line   string        `json:"line,omitempty"`
	Count  int           `json:"count"`
	Rates  []FreightRate `json:"rates"`
	Ports  []string      `json:"ports,omitempty"` // every discharge port with a quote
	Hint   string        `json:"hint,omitempty"`
}

// normalizePortFilter folds the common country/spelling variants onto a port
// token: "Algeria" (the country) means "no port filter" (list them all) and
// "Alger" means "ALGIERS". Everything else is upper-cased to A–Z only.
func normalizePortFilter(p string) string {
	switch normPort(p) {
	case "", "ALGERIA", "ALGERIE":
		return ""
	case "ALGER":
		return "ALGIERS"
	}
	return normPort(p)
}

// portMatches reports whether a discharge-port name answers a filter token
// (substring either way, so "ALGIERS / ANNABA (ALGERIA)" matches "ALGIERS").
func portMatches(filter, discharge string) bool {
	if filter == "" {
		return true
	}
	d := normPort(discharge)
	return strings.Contains(d, filter) || strings.Contains(filter, d)
}

// preferFreight picks the better of two quotes for the same port+line: a valid
// quote beats an expired one, then the cheaper wins, then the newer.
func preferFreight(a, b dbp.FreightQuote) bool {
	av := daysUntil(freightExpiry(a)) >= 0
	bv := daysUntil(freightExpiry(b)) >= 0
	if av != bv {
		return av
	}
	if a.Container40ftUsd != b.Container40ftUsd {
		return a.Container40ftUsd < b.Container40ftUsd
	}
	ta, oka := timex.ParseIso(a.Timestamp)
	tb, okb := timex.ParseIso(b.Timestamp)
	if oka && okb {
		return ta.After(tb)
	}
	return false
}

// FreightList returns the best quote per shipping line for the requested port
// (or per port+line when no port is given), cheapest first. A requested port
// that matches nothing falls back to listing every port, so a country-level
// question ("shipping to Algeria") still gets an answer instead of a deferral.
func FreightList(d *Data, port, line string) *FreightListResult {
	filter := normalizePortFilter(port)
	lineQ := strings.TrimSpace(line)
	scoped := filter != "" // a specific port was asked for

	allPorts := distinctPorts(d.Freight.Freight)

	pick := func(pf string) map[string]dbp.FreightQuote {
		best := map[string]dbp.FreightQuote{}
		for _, f := range d.Freight.Freight {
			if !portMatches(pf, f.DischargePort) {
				continue
			}
			if lineQ != "" && !strings.EqualFold(strings.TrimSpace(f.ShippingLine), lineQ) {
				continue
			}
			// Scoped to one port: collapse to one row per shipping line (a line
			// is the unit that competes). Listing all ports: keep port+line.
			k := f.ShippingLine
			if !scoped {
				k = f.DischargePort + "\x00" + f.ShippingLine
			}
			if cur, ok := best[k]; !ok || preferFreight(f, cur) {
				best[k] = f
			}
		}
		return best
	}

	note := "Freight per 40ft container (ex-" + loadingLabel(d.Freight.Freight) + "). " +
		"perCar* = (container + 300 USD freight benefit) / units, where a big car is 3 per container and a small car is 4. " +
		"Cheapest line per port wins automatically; expired quotes are flagged."

	best := pick(filter)
	usedFilter := filter
	if len(best) == 0 && filter != "" {
		// Unknown port (typo, alias, or a country with no exact match): show all.
		scoped = false
		best = pick("")
		usedFilter = ""
		note = `No freight quote matched "` + port + `"; showing every port. ` + note
	}

	if len(best) == 0 {
		return &FreightListResult{
			Status: "MISSING",
			Note:   "No freight quotes in freight.",
			Ports:  allPorts,
			Hint:   "Ask the boss for a freight cost. Do not guess one.",
		}
	}

	rates := make([]FreightRate, 0, len(best))
	for _, f := range best {
		charge := jsRound(f.Container40ftUsd+freightBenefitUsd, 2)
		rates = append(rates, FreightRate{
			DischargePort:      f.DischargePort,
			ShippingLine:       f.ShippingLine,
			LoadingPort:        f.LoadingPort,
			Container40ftUsd:   f.Container40ftUsd,
			ContainerChargeUsd: charge,
			PerCarBigUsd:       jsRound(charge/3, 2),
			PerCarSmallUsd:     jsRound(charge/4, 2),
			Etd:                f.Etd,
			ValidFrom:          f.ValidFrom,
			ValidTo:            f.ValidTo,
			FreeTimeDays:       f.FreeTimeDays,
			Expired:            daysUntil(freightExpiry(f)) < 0,
			Timestamp:          f.Timestamp,
			Notes:              f.Notes,
		})
	}
	sort.SliceStable(rates, func(i, j int) bool {
		if !strings.EqualFold(rates[i].DischargePort, rates[j].DischargePort) {
			return rates[i].DischargePort < rates[j].DischargePort
		}
		return rates[i].Container40ftUsd < rates[j].Container40ftUsd
	})

	return &FreightListResult{
		Status: "OK",
		Note:   note,
		Port:   usedFilter,
		Line:   lineQ,
		Count:  len(rates),
		Rates:  rates,
		Ports:  allPorts,
	}
}

func distinctPorts(freight []dbp.FreightQuote) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range freight {
		if !seen[f.DischargePort] {
			seen[f.DischargePort] = true
			out = append(out, f.DischargePort)
		}
	}
	sort.Strings(out)
	return out
}

func loadingLabel(freight []dbp.FreightQuote) string {
	for _, f := range freight {
		if f.LoadingPort != "" {
			return f.LoadingPort
		}
	}
	return "China"
}
