// group-perms — manage the cars-app chat AI's per-group capabilities.
//
//	whatsapp-tool group-perms list [--tenant mig_26] [--all]
//	whatsapp-tool group-perms set (--group-name "<name>" | --group <id>) [--tenant mig_26] \
//	    [--can_give_edit_instruction=<t|f>] [--can_c_suppliers_info=<t|f>] [--can_c_cost=<t|f>]
//
// This is the operator's only way to grant a group the right to give edit
// instructions, see supplier info, or see cost. It is REFUSED whenever
// WBOT_PERMS is set (i.e. from inside the bot), so the sandboxed agent can never
// escalate its own group's permissions. Unknown/ambiguous group names are
// rejected rather than guessed.
package main

import (
	"strconv"
	"strings"

	"whatsapp-agent-tool/internal/carshttp"
	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db/groupperms"
	"whatsapp-agent-tool/internal/groupauth"
)

const groupPermsUsage = "usage: whatsapp-tool group-perms <list [--tenant T] [--all] | set (--group-name \"<name>\" | --group <id>) [--tenant T] " +
	"--can_give_edit_instruction=<true|false> --can_c_suppliers_info=<true|false> --can_c_cost=<true|false>>"

func cmdGroupPerms(args []string) {
	// Operator-only: a bot context (WBOT_PERMS set) must never change its own
	// capabilities. This is the anti-escalation guard.
	if groupauth.Restricted() {
		failErr("group-perms is operator-only and cannot be run from the bot")
	}
	if len(args) == 0 {
		failErr(groupPermsUsage)
	}
	switch args[0] {
	case "list":
		groupPermsList(args[1:])
	case "set":
		groupPermsSet(args[1:])
	default:
		failErr(groupPermsUsage)
	}
}

// groupPermsList prints every group of a tenant with its flags, resolving the
// active groups from the cars app so an unconfigured group shows as all-false.
func groupPermsList(args []string) {
	tenant := strings.TrimSpace(flag(args, "tenant"))
	if tenant == "" && !has(args, "all") {
		tenant = config.C.Cars.Tenant
	}

	rows, err := groupperms.List(tenant)
	if err != nil {
		failErr(err.Error())
	}
	if tenant == "" {
		// --all: only the configured rows (no single tenant to list groups for).
		ok(pair("action", "group-perms-list"), pair("tenant", "all"), pair("count", len(rows)), pair("groups", rows))
		return
	}

	groups, err := carshttp.AllGroups()
	if err != nil {
		carsFail(err)
	}
	byID := make(map[int64]groupperms.Perms, len(rows))
	for _, r := range rows {
		byID[r.GroupID] = r
	}

	type listRow struct {
		GroupID    int64  `json:"id_chat_group"`
		Name       string `json:"name"`
		Edit       bool   `json:"can_give_edit_instruction"`
		Suppliers  bool   `json:"can_c_suppliers_info"`
		Cost       bool   `json:"can_c_cost"`
		Configured bool   `json:"configured"`
	}
	out := make([]listRow, 0, len(groups))
	seen := make(map[int64]bool, len(groups))
	for _, g := range groups {
		seen[g.ID] = true
		p := byID[g.ID]
		_, configured := byID[g.ID]
		out = append(out, listRow{GroupID: g.ID, Name: g.Name, Edit: p.Edit, Suppliers: p.Suppliers, Cost: p.Cost, Configured: configured})
	}
	for _, p := range rows {
		if !seen[p.GroupID] {
			out = append(out, listRow{GroupID: p.GroupID, Name: "", Edit: p.Edit, Suppliers: p.Suppliers, Cost: p.Cost, Configured: true})
		}
	}
	ok(pair("action", "group-perms-list"), pair("tenant", tenant), pair("count", len(out)), pair("groups", out))
}

// groupPermsSet resolves the group and applies the given flags.
func groupPermsSet(args []string) {
	tenant := strings.TrimSpace(flag(args, "tenant"))
	if tenant == "" {
		tenant = config.C.Cars.Tenant
	}
	if tenant == "" {
		failErr("no tenant: set CARS_TENANT or pass --tenant")
	}

	groupID, err := resolveGroupID(args)
	if err != nil {
		failErr(err.Error())
	}

	patch := groupperms.Patch{}
	set := false
	if v, ok := boolFlag(args, groupauth.Edit); ok {
		patch.Edit, set = &v, true
	}
	if v, ok := boolFlag(args, groupauth.Suppliers); ok {
		patch.Suppliers, set = &v, true
	}
	if v, ok := boolFlag(args, groupauth.Cost); ok {
		patch.Cost, set = &v, true
	}
	if !set {
		failErr("no flags given: pass at least one of --" + groupauth.Edit + " / --" + groupauth.Suppliers + " / --" + groupauth.Cost)
	}

	p, err := groupperms.Set(tenant, groupID, patch)
	if err != nil {
		failErr(err.Error())
	}
	ok(pair("action", "group-perms-set"), pair("tenant", p.Tenant), pair("id_chat_group", p.GroupID),
		pair(groupauth.Edit, p.Edit), pair(groupauth.Suppliers, p.Suppliers), pair(groupauth.Cost, p.Cost))
}

// resolveGroupID returns the --group id, or resolves --group-name to a unique
// active group id (refusing an unknown or ambiguous name).
func resolveGroupID(args []string) (int64, error) {
	if raw := strings.TrimSpace(flag(args, "group")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			return 0, errInvalidGroupID
		}
		return id, nil
	}
	name := strings.TrimSpace(flag(args, "group-name"))
	if name == "" {
		return 0, errNoGroup
	}
	groups, err := carshttp.AllGroups()
	if err != nil {
		return 0, err
	}
	var matches []carshttp.ChatGroup
	for _, g := range groups {
		if strings.EqualFold(g.Name, name) {
			matches = append(matches, g)
		}
	}
	switch len(matches) {
	case 0:
		return 0, &apisErr{"no active group named \"" + name + "\" — list them with `whatsapp-tool group-perms list`"}
	case 1:
		return matches[0].ID, nil
	default:
		return 0, &apisErr{"the name \"" + name + "\" matches several groups; pass --group <id> instead"}
	}
}

// boolFlag parses --name=true|false|1|0. ok is false when the flag is absent.
func boolFlag(args []string, name string) (value bool, ok bool) {
	raw := strings.TrimSpace(flag(args, name))
	if raw == "" {
		return false, false
	}
	switch strings.ToLower(raw) {
	case "true", "1", "yes", "on":
		return true, true
	case "false", "0", "no", "off":
		return false, true
	}
	failErr("--" + name + " must be true or false")
	return false, false
}

type apisErr struct{ msg string }

func (e *apisErr) Error() string { return e.msg }

var errNoGroup = &apisErr{"set needs --group-name \"<name>\" or --group <id>"}
var errInvalidGroupID = &apisErr{"--group must be a positive integer id"}
