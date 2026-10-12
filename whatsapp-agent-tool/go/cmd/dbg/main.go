package main

import (
	"fmt"

	"whatsapp-agent-tool/internal/db"
)

func dump(q, label string) {
	rows, err := db.Query(q)
	if err != nil {
		fmt.Println(label, "ERR:", err)
		return
	}
	fmt.Println("==", label)
	for _, r := range rows {
		fmt.Printf("  %v\n", r)
	}
}

func main() {
	dump("SELECT id, name FROM loading_ports ORDER BY name", "loading_ports")
	dump("SELECT id, name FROM discharge_ports ORDER BY name", "discharge_ports")
	dump("SELECT id, name FROM shipping_lines ORDER BY name", "shipping_lines")
	dump("SELECT f.id, lp.name lp, dp.name dp, sl.name sl, f.container_40ft_usd, f.etd, f.valid_to, f.quoted_at FROM freight f JOIN loading_ports lp ON lp.id=f.id_loading_port JOIN discharge_ports dp ON dp.id=f.id_discharge_port JOIN shipping_lines sl ON sl.id=f.id_shipping_line ORDER BY f.quoted_at DESC LIMIT 15", "recent freight")
}
