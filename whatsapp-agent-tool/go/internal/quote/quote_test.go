package quote

import (
	"strings"
	"testing"
	"time"

	dbp "whatsapp-agent-tool/internal/db/pricing"
)

func isoAgo(d time.Duration) string { return time.Now().Add(-d).UTC().Format(time.RFC3339) }
func isoSptrAgo(d time.Duration) *string {
	s := isoAgo(d)
	return &s
}
func dateIn(d time.Duration) string { return time.Now().Add(d).UTC().Format("2006-01-02") }

func freshAge() *string { return isoSptrAgo(24 * time.Hour) }

func mustNum(t *testing.T, v any) float64 {
	t.Helper()
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("want float64, got %T (%v)", v, v)
	}
	return f
}

func baseData() *Data {
	return &Data{
		Benefits: dbp.Benefits{WholesaleBenefitUsd: 300, RetailBenefitUsd: 700, Timestamp: freshAge()},
		Rate:     dbp.Rate{UsdToDza: 253.4, Timestamp: freshAge()},
		Custom:   dbp.Custom{AllInRetailBenefitUsd: 1000, UpdatedAt: isoSptrAgo(3 * 24 * time.Hour)},
		Cars: []dbp.Car{{
			CarId: "coolray-battle-white", Name: "COOLRAY BATTLE (WHITE)", Size: "small",
			Condition: "new", CostUsd: 11200, Colors: []string{"white"}, Timestamp: freshAge(),
		}},
	}
}

func runMain(t *testing.T, d *Data, a Args) *mainResult {
	t.Helper()
	m, ok := Run(d, a).Payload.(*mainResult)
	if !ok {
		t.Fatalf("want *mainResult, got %T", Run(d, a).Payload)
	}
	return m
}

// Stale rate/benefits/car/freight must NOT block: status stays OK, the price is
// still produced, and the age is reported in stale/warnings.
func TestStaleNeverBlocks(t *testing.T) {
	d := baseData()
	old := isoSptrAgo(10 * 24 * time.Hour)
	d.Rate.Timestamp = old
	d.Benefits.Timestamp = old
	d.Cars[0].Timestamp = old
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{{
		ShippingLine: "CMA", DischargePort: "ALGIERS / ANNABA (ALGERIA)",
		Container40ftUsd: 5196, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(10 * 24 * time.Hour),
	}}}

	m := runMain(t, d, Args{CarQuery: "battle white", Mode: "retail-dza", Port: "ALGIERS"})
	if m.Status != "OK" {
		t.Fatalf("status = %q, want OK", m.Status)
	}
	if m.Price.Dza == nil || m.Price.Usd == nil {
		t.Fatalf("price missing: %+v", m.Price)
	}
	if len(m.Stale) == 0 {
		t.Fatalf("expected stale entries, got none")
	}
	if len(m.Lines) != 1 || !m.Lines[0].Chosen {
		t.Fatalf("want 1 chosen line, got %+v", m.Lines)
	}
	if m.Action == "" {
		t.Fatalf("expected an action note for stale data")
	}
}

// When every applicable freight quote is expired we still quote the cheapest
// one and say so — no MISSING.
func TestAllExpiredUsesCheapestAndWarns(t *testing.T) {
	d := baseData()
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{
		{ShippingLine: "HMM", DischargePort: "ALGIERS / ANNABA (ALGERIA)", Container40ftUsd: 5500, ValidTo: dateIn(-24 * time.Hour), Timestamp: isoAgo(4 * 24 * time.Hour)},
		{ShippingLine: "HMM", DischargePort: "ALGIERS / ANNABA (ALGERIA)", Container40ftUsd: 5780, ValidTo: dateIn(-8 * 24 * time.Hour), Timestamp: isoAgo(9 * 24 * time.Hour)},
	}}

	m := runMain(t, d, Args{CarQuery: "battle white", Mode: "retail-dza", Port: "ALGIERS"})
	if m.Status != "OK" {
		t.Fatalf("status = %q, want OK", m.Status)
	}
	if len(m.Lines) != 2 || !m.Lines[0].Chosen || m.Lines[0].Container40ftUsd != 5500 {
		t.Fatalf("want cheapest HMM 5500 chosen, got %+v", m.Lines)
	}
	found := false
	for _, w := range m.Warnings {
		if strings.Contains(w, "expired") && strings.Contains(w, "cheapest") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected an all-expired warning, got %v", m.Warnings)
	}
}

// lines[] carries one price per applicable shipping line; the cheapest is the
// headline price and is flagged chosen.
func TestPricePerShippingLine(t *testing.T) {
	d := baseData()
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{
		{ShippingLine: "CMA", DischargePort: "ALGIERS / ANNABA (ALGERIA)", Container40ftUsd: 5196, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour)},
		{ShippingLine: "MSK", DischargePort: "ALGIERS / ANNABA (ALGERIA)", Container40ftUsd: 4800, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour)},
	}}

	m := runMain(t, d, Args{CarQuery: "battle white", Mode: "retail-dza", Port: "ALGIERS"})
	if len(m.Lines) != 2 {
		t.Fatalf("want 2 lines, got %d", len(m.Lines))
	}
	if m.Lines[0].ShippingLine != "MSK" || !m.Lines[0].Chosen {
		t.Fatalf("want cheapest MSK chosen first, got %+v", m.Lines[0])
	}
	if m.Lines[1].Chosen {
		t.Fatalf("only one line may be chosen")
	}
	if m.Price.Dza == nil || m.Lines[0].Dza != m.Price.Dza {
		t.Fatalf("headline price must equal the chosen line: price=%v line=%v", m.Price.Dza, m.Lines[0].Dza)
	}
	if m.Lines[0].Dza == m.Lines[1].Dza {
		t.Fatalf("lines with different freight must differ in price")
	}
	if m.Action != "" {
		t.Fatalf("fresh data should carry no action note, got %q", m.Action)
	}
}

// A port with no freight quote is the one case that still returns MISSING.
func TestMissingFreightPort(t *testing.T) {
	d := baseData()
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{{
		ShippingLine: "CMA", DischargePort: "ALGIERS / ANNABA (ALGERIA)",
		Container40ftUsd: 5196, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour),
	}}}

	p := Run(d, Args{CarQuery: "battle white", Mode: "retail-dza", Port: "NOWHERE"}).Payload
	mf, ok := p.(*MissingFreightResult)
	if !ok {
		t.Fatalf("want *MissingFreightResult, got %T", p)
	}
	if mf.Status != "MISSING" {
		t.Fatalf("status = %q, want MISSING", mf.Status)
	}
}

// FOB modes carry no freight, so no per-line price array.
func TestFobHasNoLines(t *testing.T) {
	d := baseData()
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{{
		ShippingLine: "CMA", DischargePort: "ALGIERS / ANNABA (ALGERIA)",
		Container40ftUsd: 5196, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour),
	}}}

	m := runMain(t, d, Args{CarQuery: "battle white", Mode: "wholesale-fob"})
	if m.Status != "OK" || len(m.Lines) != 0 {
		t.Fatalf("FOB should be OK with no lines, got status=%q lines=%d", m.Status, len(m.Lines))
	}
}

// A car whose cars.notes carry the operator's block marker ("do not quote" /
// uppercase "BLOCKED") is refused outright — BLOCKED, no price computed.
func TestBlockedCarRefused(t *testing.T) {
	d := baseData()
	d.Cars[0].Notes = "KENT BLOCKED 2026-10-16 - not available, do not quote"

	p := Run(d, Args{CarQuery: "battle white", Mode: "retail-dza", Port: "ALGIERS"}).Payload
	br, ok := p.(*BlockedCarResult)
	if !ok {
		t.Fatalf("want *BlockedCarResult, got %T", p)
	}
	if br.Status != "BLOCKED" {
		t.Fatalf("status = %q, want BLOCKED", br.Status)
	}
}

// A lowercase "kent blocked … reverted to jean price" note is NOT a block —
// those cars were reverted to another supplier and stay quotable.
func TestRevertedNoteIsNotBlocked(t *testing.T) {
	d := baseData()
	d.Cars[0].Notes = "kent blocked 2026-10-16 - reverted to jean price"
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{{
		ShippingLine: "CMA", DischargePort: "ALGIERS / ANNABA (ALGERIA)",
		Container40ftUsd: 5196, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour),
	}}}

	m := runMain(t, d, Args{CarQuery: "battle white", Mode: "retail-dza", Port: "ALGIERS"})
	if m.Status != "OK" {
		t.Fatalf("status = %q, want OK (reverted cars stay quotable)", m.Status)
	}
}

// freightFor adds a single valid CMA quote to ALGIERS so retail-dza can run.
func withFreight(d *Data) *Data {
	d.Freight = dbp.Freight{Freight: []dbp.FreightQuote{{
		ShippingLine: "CMA", DischargePort: "ALGIERS / ANNABA (ALGERIA)",
		Container40ftUsd: 5196, ValidTo: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour),
	}}}
	return d
}

// colourCar is a car whose "white" colour costs more than the base ("black").
func colourCar() *Data {
	d := baseData()
	d.Cars = []dbp.Car{{
		CarId: "coolray-battle", Name: "COOLRAY BATTLE", Size: "small",
		Condition: "new", CostUsd: 11000,
		Colors:     []string{"black", "white"},
		ColorCosts: map[string]float64{"white": 11200},
		Timestamp:  freshAge(),
	}}
	return withFreight(d)
}

// Default (no --color) quotes the CHEAPEST colour — the base cost here, since
// only white has a (higher) override.
func TestDefaultColourIsCheapest(t *testing.T) {
	m := runMain(t, colourCar(), Args{CarQuery: "battle", Mode: "retail-dza", Port: "ALGIERS"})
	if m.Status != "OK" {
		t.Fatalf("status = %q, want OK", m.Status)
	}
	if m.Color != "black" {
		t.Fatalf("color = %q, want the cheapest (black)", m.Color)
	}
	if m.ColorIsCheapest == nil || !*m.ColorIsCheapest {
		t.Fatalf("colorIsCheapest = %v, want true for the default colour", m.ColorIsCheapest)
	}
	if m.Car.CostUsd != 11000 {
		t.Fatalf("car.costUsd = %v, want the base 11000", m.Car.CostUsd)
	}
	// retail-dza = (11000 + 700 + (5196+300)/4) × 253.4
	wantUsd := 11000.0 + 700 + (5196+300)/4.0
	if m.Price.Usd == nil || mustNum(t, m.Price.Usd) != wantUsd {
		t.Fatalf("usd = %v, want %v", m.Price.Usd, wantUsd)
	}
}

// --color white uses that colour's override and is flagged as not cheapest.
func TestExplicitColourUsesOverride(t *testing.T) {
	m := runMain(t, colourCar(), Args{CarQuery: "battle", Mode: "retail-dza", Port: "ALGIERS", Color: "White"})
	if m.Status != "OK" {
		t.Fatalf("status = %q, want OK", m.Status)
	}
	if m.Color != "white" {
		t.Fatalf("color = %q, want white (case-insensitive match)", m.Color)
	}
	if m.ColorIsCheapest == nil || *m.ColorIsCheapest {
		t.Fatalf("colorIsCheapest = %v, want false (white costs more)", m.ColorIsCheapest)
	}
	wantUsd := 11200.0 + 700 + (5196+300)/4.0
	if m.Price.Usd == nil || mustNum(t, m.Price.Usd) != wantUsd {
		t.Fatalf("usd = %v, want %v (white override)", m.Price.Usd, wantUsd)
	}
	if m.Car.ColorCosts["white"] != 11200 {
		t.Fatalf("car.colorCosts = %v, want white:11200", m.Car.ColorCosts)
	}
}

// A colour the car does not list is MISSING — never guess a colour price.
func TestUnknownColourMissing(t *testing.T) {
	p := Run(colourCar(), Args{CarQuery: "battle", Mode: "retail-dza", Port: "ALGIERS", Color: "red"}).Payload
	mc, ok := p.(*MissingColorResult)
	if !ok {
		t.Fatalf("want *MissingColorResult, got %T", p)
	}
	if mc.Status != "MISSING" || len(mc.Colors) != 2 {
		t.Fatalf("want MISSING with 2 colours, got status=%q colors=%v", mc.Status, mc.Colors)
	}
}
