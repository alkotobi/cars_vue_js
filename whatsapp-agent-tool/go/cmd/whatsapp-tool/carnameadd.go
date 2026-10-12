// car-names add — create a model in the cars app, mirroring CarNamesView.vue /
// useCarNames.js addCarName() (INSERT INTO cars_names (car_name, notes,
// is_big_car, id_brand)). This mutates the tenant's data, so it only runs when
// invoked explicitly, and is gated on can_purchase_cars (modelsPermission) — the
// models tab that hosts model management. The brand must already exist.
package main

import (
	"strings"

	"whatsapp-agent-tool/internal/carshttp"
)

func cmdCarNameAdd(args []string) {
	positionals := positionalArgs(args, map[string]bool{"--name": true, "--brand-id": true, "--notes": true})
	name := strings.TrimSpace(flag(args, "name"))
	if name == "" {
		name = strings.TrimSpace(strings.Join(positionals, " "))
	}
	if name == "" {
		failErr("car-names add requires a name: whatsapp-tool car-names add --name <model> --brand-id <id> --big|--small [--notes ..]")
	}

	brandID, hasBrand := optInt(args, "brand-id")
	if !hasBrand {
		failErr("car-names add requires --brand-id <id> (list ids with: whatsapp-tool brands)")
	}

	// big/small is a required choice — the DB column has no default, and the
	// operator asked not to have it assumed.
	big, hasSize := boolChoice(args, "big", "small")
	if !hasSize {
		failErr("car-names add requires --big or --small (is the model a big car or a small car?)")
	}

	requireCarsPermission(modelsPermission)

	id, err := carshttp.CreateCarName(name, brandID, big, flag(args, "notes"))
	if err != nil {
		carsFail(err)
	}
	ok(pair("action", "car-name-add"), pair("id", id),
		pair("carName", strings.ToUpper(strings.TrimSpace(name))),
		pair("brandId", brandID), pair("isBigCar", big))
}
