// clearance — per-car customs clearance (DZD, TVN excluded), table clearance
// (migration 036), joined to its model and brand (cars_names / brands). Reads and
// the `set` write go through the same AI user and the query passthrough. No
// permission gate: clearance is free (available to any logged-in AI user).
//
//	whatsapp-tool clearance                                   # all rows
//	whatsapp-tool clearance [--id N] [--car-name-id N] [--name X] [--brand Y] [--limit N]
//	whatsapp-tool clearance set --car-name-id N --dza N (--big|--small) (--used|--new) [--notes ..]
package main

import (
	"encoding/json"
	"strings"

	"whatsapp-agent-tool/internal/carshttp"
)

func cmdClearance(args []string) {
	if len(args) > 0 && args[0] == "set" {
		cmdClearanceSet(args[1:])
		return
	}

	filters := carshttp.ClearanceFilters{}
	if n, ok := optInt(args, "id"); ok {
		filters.ID = n
	}
	if n, ok := optInt(args, "car-name-id"); ok {
		filters.CarNameID = n
	}
	if n, ok := optInt(args, "limit"); ok {
		filters.Limit = n
	}
	filters.Name = flag(args, "name")
	filters.Brand = flag(args, "brand")
	if filters.Name == "" {
		filters.Name = strings.Join(positionalArgs(args, map[string]bool{"--id": true, "--car-name-id": true, "--name": true, "--brand": true, "--limit": true}), " ")
	}

	rows, err := carshttp.SearchClearance(filters)
	if err != nil {
		carsFail(err)
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}
	ok(pair("action", "clearance"), pair("filters", filters), pair("count", len(rows)), pair("clearance", rows))
}

// cmdClearanceSet upserts one clearance row (INSERT ... ON DUPLICATE KEY UPDATE on
// the (id_car_name, is_big_car, is_used_car) unique key). The car name must exist
// (list ids with `whatsapp-tool car-names`) and big/small + used/new are part of
// the key, so they are required explicit choices — never guessed.
func cmdClearanceSet(args []string) {
	carNameID, hasCarName := optInt(args, "car-name-id")
	if !hasCarName {
		failErr("clearance set requires --car-name-id <id> (list ids with: whatsapp-tool car-names)")
	}
	dza, hasDza := optFloat(args, "dza")
	if !hasDza {
		failErr("clearance set requires --dza <amount> (customs clearance in DZD, TVN excluded)")
	}

	big, hasSize := boolChoice(args, "big", "small")
	if !hasSize {
		failErr("clearance set requires --big or --small (which variant does this clearance apply to?)")
	}
	used, hasUsed := boolChoice(args, "used", "new")
	if !hasUsed {
		failErr("clearance set requires --used or --new (which variant does this clearance apply to?)")
	}

	var notes *string
	if has(args, "notes") {
		n := flag(args, "notes")
		notes = &n
	}

	result, err := carshttp.SetClearance(carNameID, dza, big, used, notes)
	if err != nil {
		carsFail(err)
	}

	var notesOut any
	if notes != nil {
		notesOut = *notes
	}
	ok(pair("action", "clearance-set"), pair("carNameId", carNameID), pair("dza", dza),
		pair("isBigCar", big), pair("isUsedCar", used), pair("notes", notesOut),
		pair("inserted", result.LastInsertID), pair("affectedRows", result.AffectedRows))
}
