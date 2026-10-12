// Temporary dev tool: dbmig inspect|backup|apply|verify
// Inspects the live `ai` schema, dumps a row-level backup, applies a
// migration SQL file, and verifies counts. Not meant to be committed.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dbmig inspect|backup <dir>|apply <file.sql>|verify <dir>")
		os.Exit(2)
	}
	if _, err := db.Open(); err != nil {
		fatal(err)
	}
	defer db.ClosePool()

	switch os.Args[1] {
	case "inspect":
		tables, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'ai' ORDER BY table_name")
		if err != nil {
			fatal(err)
		}
		for _, t := range tables {
			name := t["table_name"].(string)
			rows, err := db.Query("SHOW CREATE TABLE `" + name + "`")
			if err != nil {
				fatal(err)
			}
			cnt, err := db.QueryOne("SELECT COUNT(*) AS n FROM `" + name + "`")
			if err != nil {
				fatal(err)
			}
			fmt.Printf("===== %s (%v rows) =====\n%s\n", name, cnt["n"], rows[0]["Create Table"])
		}
	case "backup":
		dir := os.Args[2]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fatal(err)
		}
		tables, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'ai' ORDER BY table_name")
		if err != nil {
			fatal(err)
		}
		for _, t := range tables {
			name := t["table_name"].(string)
			rows, err := db.Query("SELECT * FROM `" + name + "`")
			if err != nil {
				fatal(err)
			}
			data, _ := json.MarshalIndent(rows, "", "  ")
			if err := os.WriteFile(filepath.Join(dir, name+".json"), data, 0o644); err != nil {
				fatal(err)
			}
			fmt.Printf("backed up %s: %d rows\n", name, len(rows))
		}
	case "apply":
		f, err := os.ReadFile(os.Args[2])
		if err != nil {
			fatal(err)
		}
		for i, stmt := range splitStatements(string(f)) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			n, err := db.Execute(stmt)
			if err != nil {
				fatal(fmt.Errorf("statement %d failed: %s\n%s", i+1, err, stmt))
			}
			fmt.Printf("[%02d] affected=%d\n%s\n", i+1, n, stmt)
		}
	case "check":
		for _, q := range []string{
			"SELECT * FROM benefits_config",
			"SELECT * FROM rate_config",
			"SELECT * FROM customs_config",
			"SELECT version FROM schema_migrations",
			"SELECT c.car_id, cc.color FROM car_colors cc JOIN cars c ON c.id = cc.car_ref ORDER BY cc.color",
			"SELECT c.car_id, co.option_name, co.price_usd FROM car_options co JOIN cars c ON c.id = co.car_ref ORDER BY co.id",
			"SELECT h.car_id, c.car_id AS ref_id, h.name, h.cost_usd FROM car_price_history h LEFT JOIN cars c ON c.id = h.car_ref ORDER BY h.id LIMIT 3",
			"SELECT f.id, lp.name AS loading, dp.name AS discharge, sl.name AS line, f.container_40ft_usd FROM freight f JOIN loading_ports lp ON lp.id = f.id_loading_port JOIN discharge_ports dp ON dp.id = f.id_discharge_port JOIN shipping_lines sl ON sl.id = f.id_shipping_line ORDER BY f.id",
			"SELECT config_key, value_json, effective_at FROM config ORDER BY config_key",
			"SELECT VERSION() AS v",
		} {
			rows, err := db.Query(q)
			if err != nil {
				fatal(err)
			}
			for _, r := range rows {
				for k, v := range r {
					fmt.Printf("%s=%v ", k, v)
				}
				fmt.Println()
			}
		}
	case "fixcustoms":
		cfg, err := db.QueryOne("SELECT currency, tvn_included, all_in_retail_benefit_usd, effective_at, created_at FROM customs_config ORDER BY effective_at DESC, id DESC LIMIT 1")
		if err != nil {
			fatal(err)
		}
		if cfg == nil {
			fatal(fmt.Errorf("no customs_config row"))
		}
		v := map[string]any{
			"currency":              cfg["currency"],
			"tvnIncluded":           fmt.Sprintf("%v", cfg["tvn_included"]) == "1",
			"allInRetailBenefitUsd": cfg["all_in_retail_benefit_usd"],
		}
		b, _ := json.Marshal(v)
		n, err := db.Execute("UPDATE config SET value_json = ? WHERE config_key = 'customs'", string(b))
		if err != nil {
			fatal(err)
		}
		fmt.Printf("customs config updated, affected=%d value=%s\n", n, string(b))
	case "verify":
		dir := os.Args[2]
		tables, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'ai' ORDER BY table_name")
		if err != nil {
			fatal(err)
		}
		for _, t := range tables {
			name := t["table_name"].(string)
			var want int
			if data, err := os.ReadFile(filepath.Join(dir, name+".json")); err == nil {
				var rows []json.RawMessage
				_ = json.Unmarshal(data, &rows)
				want = len(rows)
			}
			got, err := db.QueryOne("SELECT COUNT(*) AS n FROM `" + name + "`")
			if err != nil {
				fatal(err)
			}
			have := toInt(got["n"])
			mark := "ok"
			if want >= 0 && have != want {
				mark = "MISMATCH"
			}
			fmt.Printf("%s: have=%d backup=%d %s\n", name, have, want, mark)
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: dbmig inspect|backup <dir>|apply <file.sql>|verify <dir>")
		os.Exit(2)
	}
}

func toInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int64:
		return int(t)
	case uint64:
		return int(t)
	case int:
		return t
	case []byte:
		var n int
		fmt.Sscanf(string(t), "%d", &n)
		return n
	default:
		return -1
	}
}

func splitStatements(s string) []string {
	var out []string
	var cur strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}
		cur.WriteString(line)
		cur.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		out = append(out, cur.String())
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "fatal:", err)
	os.Exit(1)
}
