// car-names edit — UPDATE the notes field in cars_names. Mirrors the query
// passthrough as the AI user, gated on can_purchase_cars (modelsPermission).
package main

import (
	"fmt"

	"whatsapp-agent-tool/internal/carshttp"
)

func cmdCarNameEdit(args []string) {
	idStr := flag(args, "id")
	if idStr == "" {
		failErr("car-names edit requires --id <id>: whatsapp-tool car-names edit --id <id> [--notes ..]")
	}
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id == 0 {
		failErr("car-names edit: invalid --id (must be an integer)")
	}

	notes := flag(args, "notes")
	if notes == "" {
		failErr("car-names edit requires --notes <text>")
	}

	requireCarsPermission(modelsPermission)

	affected, err := carshttp.UpdateCarNameNotes(id, notes)
	if err != nil {
		carsFail(err)
	}
	if affected == 0 {
		failErr("car-names edit: no row with id " + idStr)
	}
	row, err := carshttp.GetCarName(id)
	if err != nil {
		carsFail(err)
	}
	ok(pair("action", "car-name-edit"), pair("id", id), pair("affectedRows", affected), pair("row", row))
}
