package quote

import (
	"testing"
	"time"

	dbp "whatsapp-agent-tool/internal/db/pricing"
)

func freightData() *Data {
	return &Data{Freight: dbp.Freight{Freight: []dbp.FreightQuote{
		{ShippingLine: "CMA", DischargePort: "ALGIERS", LoadingPort: "NANSHA",
			Container40ftUsd: 5900, Etd: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour)},
		{ShippingLine: "COSCO", DischargePort: "ALGIERS", LoadingPort: "NANSHA",
			Container40ftUsd: 6100, Etd: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour)},
		{ShippingLine: "AKKON", DischargePort: "ANNABA", LoadingPort: "NANSHA",
			Container40ftUsd: 6600, Etd: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour)},
		{ShippingLine: "HMM", DischargePort: "ALGIERS / ANNABA (ALGERIA)", LoadingPort: "ANY (CHINA)",
			Container40ftUsd: 5500, ValidTo: dateIn(-2 * 24 * time.Hour), Timestamp: isoAgo(4 * 24 * time.Hour)},
	}}}
}

func TestFreightPortScopedCheapestPerLine(t *testing.T) {
	r := FreightList(freightData(), "ALGIERS", "")
	if r.Status != "OK" {
		t.Fatalf("status = %q", r.Status)
	}
	// CMA (plain, valid) and HMM (combined berth, expired) both serve ALGIERS.
	byLine := map[string]FreightRate{}
	for _, x := range r.Rates {
		byLine[x.ShippingLine] = x
	}
	if _, ok := byLine["CMA"]; !ok {
		t.Fatalf("CMA must serve ALGIERS, got %+v", r.Rates)
	}
	if _, ok := byLine["COSCO"]; !ok {
		t.Fatalf("COSCO must serve ALGIERS, got %+v", r.Rates)
	}
	// AKKON only serves ANNABA.
	if _, ok := byLine["AKKON"]; ok {
		t.Fatalf("AKKON must not appear for ALGIERS")
	}
	// The cheapest line is first: CMA 5900.
	if r.Rates[0].ShippingLine != "CMA" || r.Rates[0].Container40ftUsd != 5900 {
		t.Fatalf("cheapest first = %+v", r.Rates[0])
	}
	// Per-car share carries the +300 benefit: (5900+300)/3 and /4.
	cma := byLine["CMA"]
	if cma.ContainerChargeUsd != 6200 || cma.PerCarBigUsd != 2066.67 || cma.PerCarSmallUsd != 1550 {
		t.Fatalf("per-car math wrong: %+v", cma)
	}
	// The expired combined-berth HMM row is flagged, not hidden.
	if !byLine["HMM"].Expired {
		t.Fatalf("expired HMM must be flagged expired")
	}
}

func TestFreightCountryMeansAllPorts(t *testing.T) {
	all := FreightList(freightData(), "", "")
	country := FreightList(freightData(), "Algeria", "")
	if all.Status != "OK" || country.Status != "OK" {
		t.Fatalf("status all=%q country=%q", all.Status, country.Status)
	}
	if country.Port != "" {
		t.Fatalf("country must clear the port filter, got %q", country.Port)
	}
	if country.Count != all.Count {
		t.Fatalf("country count %d != all count %d", country.Count, all.Count)
	}
}

func TestFreightAlgerAlias(t *testing.T) {
	r := FreightList(freightData(), "ALGER", "")
	if r.Port != "ALGIERS" {
		t.Fatalf("ALGER should fold to ALGIERS, got %q", r.Port)
	}
	for _, x := range r.Rates {
		if x.ShippingLine == "AKKON" {
			t.Fatalf("ANNABA-only line leaked into an ALGIERS query")
		}
	}
}

func TestFreightUnknownPortFallsBackToListAll(t *testing.T) {
	r := FreightList(freightData(), "XYZ", "")
	if r.Status != "OK" {
		t.Fatalf("unknown port must not be MISSING, got %q", r.Status)
	}
	if r.Count != 4 {
		t.Fatalf("fallback should list every port+line, got %d", r.Count)
	}
}

func TestFreightValidBeatsExpiredForSameLine(t *testing.T) {
	d := &Data{Freight: dbp.Freight{Freight: []dbp.FreightQuote{
		{ShippingLine: "CMA", DischargePort: "ALGIERS", Container40ftUsd: 5000,
			ValidTo: dateIn(-1 * 24 * time.Hour), Timestamp: isoAgo(3 * 24 * time.Hour)},
		{ShippingLine: "CMA", DischargePort: "ALGIERS", Container40ftUsd: 5900,
			Etd: dateIn(5 * 24 * time.Hour), Timestamp: isoAgo(24 * time.Hour)},
	}}}
	r := FreightList(d, "ALGIERS", "")
	if r.Count != 1 {
		t.Fatalf("same line must collapse to one row, got %d", r.Count)
	}
	if r.Rates[0].Container40ftUsd != 5900 || r.Rates[0].Expired {
		t.Fatalf("valid quote must beat the cheaper expired one: %+v", r.Rates[0])
	}
}

func TestFreightNoQuotesIsMissing(t *testing.T) {
	r := FreightList(&Data{}, "", "")
	if r.Status != "MISSING" {
		t.Fatalf("empty freight must be MISSING, got %q", r.Status)
	}
}
