package carshttp

import (
	"strings"
	"testing"
)

func TestBuildStockQueryDefault(t *testing.T) {
	sql, params := BuildStockQuery(Filters{Status: "available"})
	if !strings.Contains(sql, "cs.hidden = 0") {
		t.Error("default query should hide hidden rows")
	}
	if !strings.Contains(sql, "cs.status = 'Available'") {
		t.Error("default query should filter available")
	}
	if strings.Contains(sql, "cs.sell_date >= ?") || strings.Contains(sql, "DATE_FORMAT(cs.sell_date") {
		t.Error("no date filter without date flags")
	}
	if len(params) != 0 {
		t.Errorf("unexpected params %v", params)
	}
}

func TestBuildStockQuerySoldThisMonth(t *testing.T) {
	sql, params := BuildStockQuery(Filters{Status: "sold", SoldThisMonth: true})
	if !strings.Contains(sql, "DATE_FORMAT(cs.sell_date, '%Y-%m') = DATE_FORMAT(CURDATE(), '%Y-%m')") {
		t.Errorf("missing this-month condition:\n%s", sql)
	}
	if len(params) != 0 {
		t.Errorf("this-month takes no params, got %v", params)
	}
	// sell_date must be projected by the subquery (outer scope has only cs.*).
	if !strings.Contains(sql, "cs.sell_date,") {
		t.Errorf("sell_date not projected:\n%s", sql)
	}
	if !strings.Contains(sql, "sb.date_sell AS sell_date") {
		t.Errorf("subquery does not alias sb.date_sell:\n%s", sql)
	}
}

func TestBuildStockQuerySinceUntil(t *testing.T) {
	sql, params := BuildStockQuery(Filters{Status: "sold", SoldSince: "2026-10-01", SoldUntil: "2026-10-31"})
	if !strings.Contains(sql, "cs.sell_date >= ?") || !strings.Contains(sql, "cs.sell_date <= ?") {
		t.Errorf("missing since/until conditions:\n%s", sql)
	}
	if len(params) != 2 || params[0] != "2026-10-01" || params[1] != "2026-10-31" {
		t.Errorf("params = %v", params)
	}
}

func TestLikeEscapes(t *testing.T) {
	if got := Like("50%_x"); got != `%50\%\_x%` {
		t.Errorf("Like = %q", got)
	}
}
