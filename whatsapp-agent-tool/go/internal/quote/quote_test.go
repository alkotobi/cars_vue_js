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
