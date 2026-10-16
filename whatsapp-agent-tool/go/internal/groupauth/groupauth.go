// Package groupauth answers one question: what is a chat group allowed to do?
//
// The cars-app bot never gets write access to the database directly — it drives
// the `whatsapp-tool` CLI, which does. Before spawning a reply, the bot sets the
// environment variable WBOT_PERMS to the comma-separated list of capabilities
// the group was granted (see internal/db/groupperms). Every CLI invocation made
// by the sandboxed agent inherits that variable, so the CLI can enforce the
// group's limits itself.
//
// CRITICAL: an ABSENT WBOT_PERMS means unrestricted — the operator typing
// `whatsapp-tool …` by hand, or any process that is not the bot. Only the bot
// sets it, and the agent's bash sandbox cannot forge the parent's environment.
// A PRESENT-but-empty value means fully restricted (all flags off).
package groupauth

import (
	"os"
	"strings"
)

// The three capability names. They are used verbatim as env values, JSON keys
// and CLI flags; never rename one without migrating the others.
const (
	Edit      = "can_give_edit_instruction"
	Suppliers = "can_c_suppliers_info"
	Cost      = "can_c_cost"
)

// EnvVar is the variable the bot sets (and the CLI reads) with the granted set.
const EnvVar = "WBOT_PERMS"

var (
	restricted bool
	granted    map[string]bool
)

func init() { loadValue() }

// loadValue (re)reads WBOT_PERMS into the package state.
func loadValue() {
	raw, ok := os.LookupEnv(EnvVar)
	load(raw, ok)
}

// load sets the package state from a raw value and whether the variable was
// present at all. Split out so tests can drive it directly.
func load(raw string, present bool) {
	restricted, granted = false, nil
	if !present {
		return // not the bot: unrestricted
	}
	restricted = true
	granted = make(map[string]bool)
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			granted[p] = true
		}
	}
}

// Restricted reports whether this process was started by the bot (WBOT_PERMS
// set, even if empty).
func Restricted() bool { return restricted }

// Allow reports whether the caller may use the named capability. Unrestricted
// callers are allowed everything; restricted callers only their grants.
func Allow(name string) bool {
	if !restricted {
		return true
	}
	return granted[name]
}

// Valid reports whether name is one of the three known capabilities.
func Valid(name string) bool {
	switch name {
	case Edit, Suppliers, Cost:
		return true
	}
	return false
}

// Names returns the known capability names in a stable order.
func Names() []string { return []string{Edit, Suppliers, Cost} }
