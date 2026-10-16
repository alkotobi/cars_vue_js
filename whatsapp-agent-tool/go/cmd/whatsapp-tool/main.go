// whatsapp-tool — the Go twin of the npm CLIs. One static binary, subcommands.
//
//	whatsapp-tool quote --car "COOLRAY BATTLE" --mode retail-dza [--port ALGIERS] [--line CMA] [--list] [--options]
//	whatsapp-tool freight [--port ALGIERS] [--line CMA]   # shipping rate per line to a port (no car needed)
//	whatsapp-tool faq  <search ".." | list [--all] | add --q .. --a .. | update --id ..>
//	whatsapp-tool stock [--car .. --color .. --vin .. --status .. --port .. --limit N --include-hidden] | creds | login
//	whatsapp-tool suppliers [name] [--name ..] [--limit N]   # supplier list (cars app)
//	whatsapp-tool suppliers add --name <name> [--contact ..] [--notes ..]   # create a supplier — writes
//	whatsapp-tool brands [brand] [--brand ..] [--limit N]    # read-only brand list (cars app)
//	whatsapp-tool brands add --name <brand> [--logo <path>]  # create a brand (optionally with logo) — writes
//	whatsapp-tool car-names [name] [--name ..] [--brand ..] [--limit N]   # read-only car-name list (cars app)
//	whatsapp-tool car-names add --name <model> --brand-id <id> (--big|--small) [--notes ..]   # create a car name — writes
//	whatsapp-tool inquiries [list | create | add-car | cars <id> | delete <id>]   # supplier quote requests (cars app)
//	whatsapp-tool buy-payments [--bill-id N | --supplier-id N | --id N] | add | edit | delete
//	whatsapp-tool todo <add|list|show|done|reopen|cancel|edit|move|assign|delete|remind|due|users>   # tasks for the logged-in user
//	whatsapp-tool identity [--reset]
//	whatsapp-tool profiles [list|mark-reply <name>|set-active <name> <true|false>]
//	whatsapp-tool pending [--state]   # trim + promote (+ group scan when browser is live)
//	whatsapp-tool triage [N] [msgs] [--no-read]   # batch unread triage -> data/triage.json
//	whatsapp-tool keepalive <start|status|stop>   # bring the shared Chrome up / report / kill
//	whatsapp-tool chat-poll [--once] [--interval N]   # cars-app chat AI (polls, generates, replies)
//	whatsapp-tool mcp              # stdio MCP server (Go twin of dist/index.js)
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"whatsapp-agent-tool/internal/carshttp"
	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db"
	"whatsapp-agent-tool/internal/db/faq"
	dbp "whatsapp-agent-tool/internal/db/pricing"
	"whatsapp-agent-tool/internal/groupauth"
	"whatsapp-agent-tool/internal/identity"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/keyringx"
	toolmcp "whatsapp-agent-tool/internal/mcp"
	"whatsapp-agent-tool/internal/profiles"
	"whatsapp-agent-tool/internal/quote"
	"whatsapp-agent-tool/internal/redact"
	"whatsapp-agent-tool/internal/wa"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	gateCommand(args[0], args[1:])
	switch args[0] {
	case "quote":
		cmdQuote(args[1:])
	case "freight":
		cmdFreight(args[1:])
	case "price":
		cmdPrice(args[1:])
	case "faq":
		cmdFaq(args[1:])
	case "procurement":
		cmdProcurement(args[1:])
	case "stock":
		cmdStock(args[1:])
	case "suppliers":
		cmdSuppliers(args[1:])
	case "brands":
		cmdBrands(args[1:])
	case "car-names":
		cmdCarNames(args[1:])
	case "inquiries":
		cmdInquiries(args[1:])
	case "clearance":
		cmdClearance(args[1:])
	case "buy-payments":
		cmdBuyPayments(args[1:])
	case "todo":
		cmdTodo(args[1:])
	case "group-perms":
		cmdGroupPerms(args[1:])
	case "identity":
		cmdIdentity(args[1:])
	case "profiles":
		cmdProfiles(args[1:])
	case "pending":
		cmdPending(wa.NewBrowser(), args[1:])
	case "triage":
		cmdTriage(wa.NewBrowser(), args[1:])
	case "keepalive":
		cmdKeepalive(wa.NewBrowser(), args[1:])
	case "chat-poll":
		cmdChatPoll(args[1:])
	case "mcp":
		cmdMCP()
	default:
		fmt.Fprintf(os.Stderr, "whatsapp-tool: unknown command %q\n", args[0])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: whatsapp-tool <quote|freight|price|faq|procurement|stock|suppliers|brands|car-names|inquiries|clearance|buy-payments|todo|group-perms|identity|profiles|pending|triage|keepalive|chat-poll|mcp> ...`)
}

// gateCommand refuses a bot-invoked command whose required capability the
// group lacks (see internal/groupauth). It is a no-op unless WBOT_PERMS is set,
// so the operator and the daemon are never constrained. Refusing the whole
// command here — before any DB or network work — is the primary enforcement;
// redaction in the print helpers covers the reads that ARE allowed.
func gateCommand(cmd string, rest []string) {
	if !groupauth.Restricted() {
		return
	}
	need := requiredPermission(cmd, rest)
	if need == "" || groupauth.Allow(need) {
		return
	}
	failErr("forbidden: this chat group is not allowed to use `"+cmd+"` ("+need+")",
		pair("code", "forbidden_permission"), pair("permission", need))
}

// requiredPermission maps a command (and its subcommands) to the single
// capability it needs, or "" when the command is unrestricted. Writes require
// can_give_edit_instruction; the two supplier-identity reads require
// can_c_suppliers_info; everything cost-bearing is handled by redaction instead
// of a hard refusal (a cost-restricted group still gets prices, not inputs).
func requiredPermission(cmd string, rest []string) string {
	sub, sub2 := "", ""
	if len(rest) > 0 {
		sub = rest[0]
	}
	if len(rest) > 1 {
		sub2 = rest[1]
	}
	in := func(names ...string) bool {
		for _, n := range names {
			if sub == n {
				return true
			}
		}
		return false
	}
	// inSub2 matches the second-level subcommand (e.g. `procurement quotations
	// create` → sub="quotations", sub2="create").
	inSub2 := func(names ...string) bool {
		for _, n := range names {
			if sub2 == n {
				return true
			}
		}
		return false
	}
	switch cmd {
	case "faq":
		if in("add", "update") {
			return groupauth.Edit
		}
	case "freight":
		// Freight is a cost basis — the container rate and the per-car share.
		return groupauth.Cost
	case "price":
		if strings.HasPrefix(sub, "set-") {
			return groupauth.Edit
		}
	case "procurement":
		switch sub {
		case "suppliers":
			if sub2 == "add" || sub2 == "update" {
				return groupauth.Edit
			}
			return groupauth.Suppliers
		case "quotations":
			if inSub2("create", "add-item", "set-price", "set-status", "delete") {
				return groupauth.Edit
			}
		case "documents":
			if inSub2("create", "link", "unlink") {
				return groupauth.Edit
			}
		case "payments":
			if inSub2("create", "edit", "delete", "link", "unlink", "set-swift") {
				return groupauth.Edit
			}
		}
	case "suppliers":
		if sub == "add" {
			return groupauth.Edit
		}
		return groupauth.Suppliers
	case "brands":
		if sub == "add" {
			return groupauth.Edit
		}
	case "car-names":
		if sub == "add" {
			return groupauth.Edit
		}
	case "inquiries":
		if in("create", "add-car", "delete") {
			return groupauth.Edit
		}
	case "clearance":
		if sub == "set" {
			return groupauth.Edit
		}
	case "buy-payments":
		if in("add", "edit", "delete") {
			return groupauth.Edit
		}
	case "todo":
		if in("add", "edit", "done", "reopen", "cancel", "move", "assign", "delete") {
			return groupauth.Edit
		}
		if sub == "remind" && inSub2("add", "rm", "sent") {
			return groupauth.Edit
		}
	}
	return ""
}

// cmdMCP runs the stdio MCP server (the Go twin of `node dist/index.js`). It
// owns stdout — diagnostics go to stderr only.
func cmdMCP() {
	if err := toolmcp.Start(wa.NewBrowser()); err != nil {
		fmt.Fprintln(os.Stderr, "[whatsapp-mcp] Fatal:", err)
		os.Exit(1)
	}
}

// flag parses --name value and --name=value. Returns "" when absent.
func flag(args []string, name string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--"+name && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(args[i], "--"+name+"=") {
			return args[i][len(name)+3:]
		}
	}
	return ""
}

func has(args []string, name string) bool {
	for _, a := range args {
		if a == "--"+name {
			return true
		}
	}
	return false
}

// boolChoice reads an explicit either/or flag pair (--used/--new, --big/--small).
// ok is false when neither was given, so callers can require an explicit choice
// instead of silently defaulting; passing both is a usage error. The first name
// maps to true.
func boolChoice(args []string, yes, no string) (value bool, ok bool) {
	y, n := has(args, yes), has(args, no)
	if y && n {
		failErr("--" + yes + " and --" + no + " are mutually exclusive")
	}
	if y {
		return true, true
	}
	if n {
		return false, true
	}
	return false, false
}

// positionalArgs returns the non-flag args, skipping the value of any
// "--name value" pair listed in valueFlags (so the value is not mistaken for a
// positional). "--name=value" form is a single token and skipped as a flag.
func positionalArgs(args []string, valueFlags map[string]bool) []string {
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			if valueFlags[arg] {
				i++
			}
			continue
		}
		positionals = append(positionals, arg)
	}
	return positionals
}

// requireCarsPermission enforces the app's role->permission model for a
// read command: the AI user's role must grant perm. Skipped when a token was
// supplied via CARS_API_TOKEN — the operator explicitly trusted that token.
func requireCarsPermission(perm string) {
	if config.C.Cars.APIToken != "" {
		return
	}
	known, allowed, err := carshttp.HasPermission(perm)
	if err != nil {
		carsFail(err)
	}
	if known && !allowed {
		carsFail(&carshttp.ApiError{
			Code:    "missing_permission",
			Message: "the AI user's role is missing " + perm + " — grant it in the cars app RolePermissionsCrud, then re-run: whatsapp-tool stock creds --username <user>",
		})
	}
}

// printJSON prints with 2-space indent (quote + identity outputs — Node uses
// JSON.stringify(payload, null, 2) for those). In a restricted (bot) context the
// payload is first stripped of cost/supplier fields the group may not see.
func printJSON(v any) {
	if groupauth.Restricted() {
		v = redact.Apply(v, groupauth.Allow(groupauth.Suppliers), groupauth.Allow(groupauth.Cost))
	}
	data, err := jsonx.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// ordered is an insertion-ordered JSON object. Go maps sort keys, and Node
// preserves insertion order, so ordered objects must be used wherever the
// golden files carry Node key order.
type ordered struct {
	Pairs []orderedPair
}

type orderedPair struct {
	Key   string
	Value any
}

func (o *ordered) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, p := range o.Pairs {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := jsonx.Marshal(p.Key)
		if err != nil {
			return nil, err
		}
		v, err := jsonx.Marshal(p.Value)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

func pair(key string, value any) orderedPair { return orderedPair{Key: key, Value: value} }

// printCompact prints one compact JSON line (faq/stock/profiles outputs). In a
// restricted (bot) context the payload is first stripped of cost/supplier fields.
func printCompact(v any) {
	if groupauth.Restricted() {
		v = redact.Apply(v, groupauth.Allow(groupauth.Suppliers), groupauth.Allow(groupauth.Cost))
	}
	data, err := jsonx.Marshal(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// ok prints {status:"OK", ...} compact, preserving key order.
func ok(pairs ...orderedPair) {
	o := ordered{Pairs: make([]orderedPair, 0, len(pairs)+1)}
	o.Pairs = append(o.Pairs, pair("status", "OK"))
	o.Pairs = append(o.Pairs, pairs...)
	printCompact(&o)
}

// failErr prints {status:"ERROR", error, ...} compact (faq/stock shape), exit 1.
func failErr(error string, pairs ...orderedPair) {
	o := ordered{Pairs: make([]orderedPair, 0, len(pairs)+2)}
	o.Pairs = append(o.Pairs, pair("status", "ERROR"), pair("error", error))
	o.Pairs = append(o.Pairs, pairs...)
	printCompact(&o)
	os.Exit(1)
}

// fail prints a compact {status:"ERROR",error} line and exits (legacy map form).
func fail(status int, payload map[string]any) {
	data, err := jsonx.Marshal(payload)
	if err != nil {
		data = []byte(`{"status":"ERROR","error":"marshal failure"}`)
	}
	fmt.Println(string(data))
	os.Exit(status)
}

// ---------------------------------------------------------------- quote

type listCar struct {
	CarId       string          `json:"carId"`
	Name        string          `json:"name"`
	Size        string          `json:"size"`
	Condition   string          `json:"condition"`
	CostUsd     float64         `json:"costUsd"`
	SupplierId  any             `json:"supplierId"`
	LoadingPort any             `json:"loadingPort"`
	Colors      []string        `json:"colors"`
	Options     []dbp.CarOption `json:"options"`
}

// listOut mirrors quote.ts' LIST payload — a struct (not a map) so the pretty
// marshal keeps Node's key order.
type listOut struct {
	Status string    `json:"status"`
	Note   string    `json:"note"`
	Modes  []string  `json:"modes"`
	Cars   []listCar `json:"cars"`
}

func cmdQuote(args []string) {
	carQ := flag(args, "car")
	mode := flag(args, "mode")
	line := flag(args, "line")
	port := flag(args, "port")
	list := has(args, "list")
	withOptions := has(args, "options")

	// Only load data when the list branch needs it or a real quote runs.
	if list || carQ == "" {
		cars, err := dbp.LoadCars()
		if err != nil {
			fail(1, map[string]any{"status": "ERROR", "error": err.Error()})
		}
		if list && carQ != "" {
			printJSON(&quote.UsageResult{Status: "USAGE", Error: `Unknown flag for car "` + carQ + `"`, Modes: quote.Modes})
			return
		}
		lc := make([]listCar, 0, len(cars))
		for _, c := range cars {
			colors := c.Colors
			if colors == nil {
				colors = []string{}
			}
			opts := c.Options
			if opts == nil {
				opts = []dbp.CarOption{}
			}
			lc = append(lc, listCar{CarId: c.CarId, Name: c.Name, Size: c.Size, Condition: c.Condition,
				CostUsd: c.CostUsd, SupplierId: c.SupplierId, LoadingPort: c.LoadingPort, Colors: colors, Options: opts})
		}
		printJSON(&listOut{Status: "LIST", Note: "Catalog (latest entry per carId).", Modes: quote.Modes, Cars: lc})
		return
	}

	d, err := loadQuoteData()
	if err != nil {
		fail(1, map[string]any{"status": "ERROR", "error": err.Error()})
	}
	res := quote.Run(d, quote.Args{CarQuery: carQ, Mode: mode, Line: line, Port: port, List: false, IncludeOptions: withOptions})
	printJSON(res.Payload)
}

// cmdFreight answers a general shipping question ("how much is shipping to
// <port>") without needing a car. It reads the same freight table the quote
// calculator uses. Port defaults to every port when none is given (or when the
// name is the country, e.g. "Algeria").
func cmdFreight(args []string) {
	port := flag(args, "port")
	line := flag(args, "line")

	freight, err := dbp.LoadFreight()
	if err != nil {
		fail(1, map[string]any{"status": "ERROR", "error": err.Error()})
	}
	printJSON(quote.FreightList(&quote.Data{Freight: freight}, port, line))
}

func loadQuoteData() (*quote.Data, error) {
	benefits, err := dbp.LoadBenefits()
	if err != nil {
		return nil, err
	}
	rate, err := dbp.LoadRate()
	if err != nil {
		return nil, err
	}
	custom, err := dbp.LoadCustom()
	if err != nil {
		return nil, err
	}
	freight, err := dbp.LoadFreight()
	if err != nil {
		return nil, err
	}
	cars, err := dbp.LoadCars()
	if err != nil {
		return nil, err
	}
	options, err := dbp.LoadOptions()
	if err != nil {
		return nil, err
	}
	return &quote.Data{Benefits: benefits, Rate: rate, Custom: custom, Freight: freight, Cars: cars, Options: options}, nil
}

// ---------------------------------------------------------------- faq

func cmdFaq(args []string) {
	if len(args) == 0 {
		fail(1, map[string]any{"status": "ERROR", "error": `usage: whatsapp-tool faq <search "term" | list [--all] | add --q .. --a .. | update --id ..>`})
	}
	cmd := args[0]
	positionals := []string{}
	for _, a := range args[1:] {
		if !strings.HasPrefix(a, "--") {
			positionals = append(positionals, a)
		}
	}
	var err error
	switch cmd {
	case "search":
		{
			term := strings.Join(positionals, " ")
			if t := flag(args, "term"); t != "" {
				term = t
			}
			if term == "" {
				failErr(`search needs a term: whatsapp-tool faq search "..."`)
			}
			limit := 5
			if l := flag(args, "limit"); l != "" {
				fmt.Sscanf(l, "%d", &limit)
			}
			var matches []faq.Entry
			matches, err = faq.Search(term, limit)
			if err == nil {
				ok(pair("action", "search"), pair("term", term), pair("matches", matches))
			}
		}
	case "list":
		{
			var entries []faq.Entry
			entries, err = faq.List(has(args, "all"))
			if err == nil {
				ok(pair("action", "list"), pair("count", len(entries)), pair("entries", entries))
			}
		}
	case "add":
		{
			q := flag(args, "q")
			a := flag(args, "a")
			if q == "" || a == "" {
				failErr("add requires --q and --a")
			}
			var id int64
			id, err = faq.Add(q, a, flag(args, "context"))
			if err == nil {
				ok(pair("action", "add"), pair("id", id))
			}
		}
	case "update":
		{
			var id int
			fmt.Sscanf(flag(args, "id"), "%d", &id)
			if flag(args, "id") == "" {
				failErr("update requires --id")
			}
			patch := faq.UpdatePatch{}
			if q := flag(args, "q"); q != "" {
				patch.Question = &q
			}
			if a := flag(args, "a"); a != "" {
				patch.Answer = &a
			}
			if c := flag(args, "context"); c != "" {
				patch.Context = c
				patch.ContextSet = true
			}
			if act := flag(args, "active"); act != "" {
				patch.Active = act == "1" || act == "true"
				patch.ActiveSet = true
			}
			var changed bool
			changed, err = faq.Update(id, patch)
			if err == nil {
				var entry *faq.Entry
				entry, err = faq.Get(id)
				if err == nil {
					ok(pair("action", "update"), pair("id", id), pair("changed", changed), pair("entry", entry))
				}
			}
		}
	default:
		failErr(fmt.Sprintf(`usage: whatsapp-tool faq <search "term" | list [--all] | add --q .. --a .. | update --id ..>`))
	}
	if err != nil {
		db.ClosePool()
		failErr(err.Error())
	}
	db.ClosePool()
	os.Exit(0)
}

// ---------------------------------------------------------------- stock

// hintFor returns Node's stock.ts hint key/value. Only the used key is emitted.
func hintFor(err *carshttp.ApiError) (string, any) {
	switch err.Code {
	case "not_configured":
		return "hint", "Set CARS_BASE_URL and CARS_TENANT in .env."
	case "missing_credentials":
		return "hint", "Ask the user for the AI user's username/password the admin created, then run: whatsapp-tool stock creds --username <user>"
	case "invalid_credentials", "login_failed":
		return "hint", "The AI user's credentials were rejected — confirm them with the admin, then run: whatsapp-tool stock creds --username <user>"
	case "missing_permission":
		return "hint", "The AI user's role is missing a required permission — grant it in the cars app's RolePermissionsCrud (or give the AI a role that has it), then re-run: whatsapp-tool stock creds --username <user>"
	case "network_error":
		return "hint", "Could not reach the cars app — check CARS_BASE_URL and connectivity."
	default:
		return "", nil
	}
}

func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func readPasswordFromStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func cmdStock(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "creds":
			stockCreds(args[1:])
			return
		case "login":
			token, err := carshttp.ResolveToken()
			if err != nil {
				carsFail(err)
			}
			ok(pair("action", "login"), pair("tokenLength", len(token)))
			return
		}
	}

	statusFlag := flag(args, "status")
	if statusFlag != "" && statusFlag != "available" && statusFlag != "sold" && statusFlag != "any" {
		failErr("--status must be available | sold | any")
	}

	valueFlags := map[string]bool{"--status": true, "--car": true, "--color": true, "--vin": true, "--port": true, "--limit": true}
	positionals := positionalArgs(args, valueFlags)

	limit := 0
	if l := flag(args, "limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	carTerm := flag(args, "car")
	if carTerm == "" {
		carTerm = strings.Join(positionals, " ")
	}
	filters := carshttp.Filters{
		Car:           carTerm,
		Color:         flag(args, "color"),
		Vin:           flag(args, "vin"),
		Port:          flag(args, "port"),
		Status:        statusFlag,
		Limit:         limit,
		IncludeHidden: has(args, "include-hidden"),
	}
	if filters.Status == "" {
		filters.Status = "available"
	}

	// The stock-view permission (can_c_car_stock) mirrors CarStockTable.
	requireCarsPermission("can_c_car_stock")

	cars, err := carshttp.SearchStock(filters)
	if err != nil {
		carsFail(err)
	}

	// Column scope: mirror the cars app's permission-gated columns. A role
	// without e.g. can_c_buy_price does not get buy_price in the output. An
	// explicit CARS_API_TOKEN is trusted the same way as the stock guard.
	if config.C.Cars.APIToken == "" {
		permRaw, perr := keyringx.GetPermissions()
		if perr != nil {
			carsFail(perr)
		}
		var perms []string
		if permRaw != "" {
			if jerr := json.Unmarshal([]byte(permRaw), &perms); jerr != nil {
				carsFail(&carshttp.ApiError{Code: "bad_response", Message: "stored permission list is invalid"})
			}
		}
		cars = carshttp.ScrubRestrictedColumns(cars, perms)
	}

	if cars == nil {
		cars = []json.RawMessage{}
	}
	ok(pair("action", "search"), pair("filters", filters), pair("count", len(cars)), pair("cars", cars))
}

// ---------------------------------------------------------------- suppliers

// SuppliersView (cars app) is reachable only with can_c_suppliers: the
// suppliers tab is gated on it in CarsView.vue, and admins bypass the check.
const suppliersPermission = "can_c_suppliers"

// cmdSuppliers reads the cars app's suppliers table (fixed SELECT), or with the
// `add` subcommand creates a supplier. It shares the AI user's credentials/token
// with `stock`, so setup is the same.
func cmdSuppliers(args []string) {
	if len(args) > 0 && args[0] == "add" {
		cmdSupplierAdd(args[1:])
		return
	}

	positionals := positionalArgs(args, map[string]bool{"--name": true, "--limit": true})

	limit := 0
	if l := flag(args, "limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	name := flag(args, "name")
	if name == "" {
		name = strings.Join(positionals, " ")
	}
	filters := carshttp.SupplierFilters{Name: name, Limit: limit}

	requireCarsPermission(suppliersPermission)

	suppliers, err := carshttp.SearchSuppliers(filters)
	if err != nil {
		carsFail(err)
	}
	if suppliers == nil {
		suppliers = []json.RawMessage{}
	}
	ok(pair("action", "suppliers"), pair("filters", filters), pair("count", len(suppliers)), pair("suppliers", suppliers))
}

// cmdSupplierAdd creates a supplier in the cars app, mirroring
// SuppliersView.vue addSupplier(). Gated on can_c_suppliers, like the suppliers
// tab that hosts supplier management. This mutates the tenant's data, so it only
// runs when the operator invokes it explicitly. contact_info/notes are optional
// (stored NULL when omitted).
func cmdSupplierAdd(args []string) {
	positionals := positionalArgs(args, map[string]bool{"--name": true, "--contact": true, "--notes": true})
	name := strings.TrimSpace(flag(args, "name"))
	if name == "" {
		name = strings.TrimSpace(strings.Join(positionals, " "))
	}
	if name == "" {
		failErr("suppliers add requires a name: whatsapp-tool suppliers add --name <name> [--contact ..] [--notes ..]")
	}

	requireCarsPermission(suppliersPermission)

	id, err := carshttp.CreateSupplier(name, flag(args, "contact"), flag(args, "notes"))
	if err != nil {
		carsFail(err)
	}
	ok(pair("action", "supplier-add"), pair("id", id), pair("name", name))
}

// ---------------------------------------------------------------- brands

// modelsPermission gates the cars app's models tab (brands + cars_names) in
// CarsView.vue; admins bypass it.
const modelsPermission = "can_purchase_cars"

// cmdBrands reads the cars app's brands table (read-only, fixed SELECT),
// mirroring the BrandsModal / useBrands.js list. The `add` subcommand instead
// creates a brand (optionally uploading a logo) — the tool's one brand write.
func cmdBrands(args []string) {
	if len(args) > 0 && args[0] == "add" {
		cmdBrandAdd(args[1:])
		return
	}

	positionals := positionalArgs(args, map[string]bool{"--brand": true, "--limit": true})

	limit := 0
	if l := flag(args, "limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	brand := flag(args, "brand")
	if brand == "" {
		brand = strings.Join(positionals, " ")
	}
	filters := carshttp.BrandFilters{Brand: brand, Limit: limit}

	requireCarsPermission(modelsPermission)

	brands, err := carshttp.SearchBrands(filters)
	if err != nil {
		carsFail(err)
	}
	if brands == nil {
		brands = []json.RawMessage{}
	}
	ok(pair("action", "brands"), pair("filters", filters), pair("count", len(brands)), pair("brands", brands))
}

// cmdBrandAdd creates a brand in the cars app, mirroring useBrands.js addBrand():
// it optionally uploads a logo first, then INSERTs the row. Gated on
// can_purchase_cars, like the models tab that hosts brand management. Unlike the
// read commands this mutates the tenant's data, and only runs when the operator
// invokes it explicitly.
func cmdBrandAdd(args []string) {
	positionals := positionalArgs(args, map[string]bool{"--name": true, "--logo": true})
	name := strings.TrimSpace(flag(args, "name"))
	if name == "" {
		name = strings.TrimSpace(strings.Join(positionals, " "))
	}
	if name == "" {
		failErr("brands add requires a name: whatsapp-tool brands add --name <brand> [--logo <path>]")
	}

	requireCarsPermission(modelsPermission)

	var logoPath string
	if logo := flag(args, "logo"); logo != "" {
		var err error
		logoPath, err = carshttp.UploadFile(logo, "uploads")
		if err != nil {
			carsFail(err)
		}
	}

	id, err := carshttp.CreateBrand(name, logoPath)
	if err != nil {
		carsFail(err)
	}

	var logoOut any
	if logoPath != "" {
		logoOut = logoPath
	}
	ok(pair("action", "brand-add"), pair("id", id), pair("brand", name), pair("logoPath", logoOut))
}

// ---------------------------------------------------------------- car-names

// cmdCarNames reads the cars app's cars_names table (read-only, fixed SELECT),
// mirroring CarNamesView / useCarNames.js (joined to brands).
// cmdCarNames reads the cars app's cars_names table (fixed SELECT), mirroring
// CarNamesView.vue / useCarNames.js. The `add` subcommand instead creates a model
// (cmdCarNameAdd), a deliberate write gated on can_purchase_cars.
func cmdCarNames(args []string) {
	if len(args) > 0 && args[0] == "add" {
		cmdCarNameAdd(args[1:])
		return
	}

	positionals := positionalArgs(args, map[string]bool{"--name": true, "--brand": true, "--limit": true})

	limit := 0
	if l := flag(args, "limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	name := flag(args, "name")
	if name == "" {
		name = strings.Join(positionals, " ")
	}
	filters := carshttp.CarNameFilters{Name: name, Brand: flag(args, "brand"), Limit: limit}

	requireCarsPermission(modelsPermission)

	carNames, err := carshttp.SearchCarNames(filters)
	if err != nil {
		carsFail(err)
	}
	if carNames == nil {
		carNames = []json.RawMessage{}
	}
	ok(pair("action", "car-names"), pair("filters", filters), pair("count", len(carNames)), pair("car_names", carNames))
}

// carsFail prints the shared cars-app error shape {status,error,code[,hint]} and exits.
func carsFail(err error) {
	ae, ok := err.(*carshttp.ApiError)
	if !ok {
		failErr(err.Error())
	}
	order := []orderedPair{pair("error", ae.Message), pair("code", ae.Code)}
	if k, v := hintFor(ae); k != "" {
		order = append(order, pair(k, v))
	}
	o := ordered{Pairs: make([]orderedPair, 0, len(order)+1)}
	o.Pairs = append(o.Pairs, pair("status", "ERROR"))
	o.Pairs = append(o.Pairs, order...)
	printCompact(&o)
	os.Exit(1)
}

func stockCreds(args []string) {
	action := "set"
	if has(args, "status") {
		action = "status"
	} else if has(args, "remove") {
		action = "remove"
	}
	if action == "status" {
		u, e1 := keyringx.GetUsername()
		p, e2 := keyringx.GetPassword()
		t, e3 := keyringx.GetToken()
		r, e4 := keyringx.GetRole()
		permRaw, e5 := keyringx.GetPermissions()
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
			carsFail(firstErr(e1, e2, e3, e4, e5))
		}
		var perms []string
		if permRaw != "" {
			_ = json.Unmarshal([]byte(permRaw), &perms)
		}
		pairs := []orderedPair{
			pair("action", "creds-status"),
			pair("username", u),
			pair("role", r),
			pair("permissions", perms),
			pair("usernameStored", u != ""),
			pair("passwordStored", p != ""),
			pair("tokenStored", t != ""),
		}
		ok(pairs...)
		return
	}
	if action == "remove" {
		ru, e1 := keyringx.DeleteUsername()
		rp, e2 := keyringx.DeletePassword()
		rt, e3 := keyringx.DeleteToken()
		rr, e4 := keyringx.DeleteRole()
		rx, e5 := keyringx.DeletePermissions()
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
			carsFail(firstErr(e1, e2, e3, e4, e5))
		}
		ok(pair("action", "creds-remove"), pair("removed", ru || rp || rt || rr || rx))
		return
	}

	username := strings.TrimSpace(flag(args, "username"))
	if username == "" {
		failErr("creds requires --username <user>")
	}
	password := flag(args, "password")
	if password == "" {
		if isTTY() {
			failErr(`pass the password via --password or pipe it: echo "..." | whatsapp-tool stock creds --username <user>`)
		}
		var err error
		password, err = readPasswordFromStdin()
		if err != nil {
			failErr(err.Error())
		}
	}
	if password == "" {
		failErr("an empty password was provided")
	}
	token, err := carshttp.SaveAndVerifyCredentials(username, password)
	if err != nil {
		carsFail(err)
	}
	ok(pair("action", "creds"), pair("username", username), pair("verified", true), pair("tokenLength", len(token)))
}

// ---------------------------------------------------------------- identity

func cmdIdentity(args []string) {
	id, fresh := identity.Get(), false
	if has(args, "--reset") {
		id = identity.Reset()
		fresh = true
	} else {
		id, fresh = identity.Ensure()
	}
	out := struct {
		Name      string `json:"name"`
		Token     string `json:"token"`
		Profile   string `json:"profile"`
		CreatedAt string `json:"createdAt"`
		File      string `json:"file"`
		Fresh     bool   `json:"fresh"`
	}{Name: id.Name, Token: id.Token, Profile: id.Profile, CreatedAt: id.CreatedAt,
		File: identity.File(config.WAProfile(), config.ProfileIsDefault()), Fresh: fresh}
	data, _ := jsonx.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
	if fresh {
		fmt.Println("minted a new AI identity for this profile.")
	} else {
		fmt.Println("loaded the existing AI identity.")
	}
	os.Exit(0)
}

// ---------------------------------------------------------------- profiles

type profileRow struct {
	Id            int     `json:"id"`
	Name          string  `json:"name"`
	Number        string  `json:"number"`
	Description   string  `json:"description"`
	Active        bool    `json:"active"`
	LastAiReplyAt *string `json:"lastAiReplyAt"`
}

func cmdProfiles(args []string) {
	cmd := "list"
	if len(args) > 0 && args[0] != "" {
		cmd = args[0]
	}
	failCLI := func(msg string) {
		fmt.Fprintf(os.Stderr, "[profiles] %s\n", msg)
		fmt.Fprintf(os.Stderr, "[profiles] usage: whatsapp-tool profiles | profiles mark-reply -- <name> | profiles set-active -- <name> <true|false>\n")
		os.Exit(1)
	}
	switch cmd {
	case "list":
		pf := profiles.Load()
		rows := make([]profileRow, 0, len(pf.Profiles))
		for _, e := range pf.Profiles {
			rows = append(rows, profileRow{Id: e.Id, Name: e.Name, Number: e.Number, Description: e.Description, Active: e.Active, LastAiReplyAt: e.LastAiReplyAt})
		}
		o := ordered{Pairs: []orderedPair{
			pair("file", profiles.ProfilesPath()),
			pair("nextId", pf.NextId),
			pair("count", len(rows)),
			pair("profiles", rows),
		}}
		printCompact(&o)
	case "mark-reply":
		name := ""
		if len(args) > 1 {
			name = args[1]
		} else if f := flag(args, "name"); f != "" {
			name = f
		}
		if name == "" {
			failCLI("mark-reply needs a profile name")
		}
		e, err := profiles.MarkAiReply(name)
		if err != nil {
			failCLI(err.Error())
		}
		o := ordered{Pairs: []orderedPair{
			pair("id", e.Id),
			pair("name", e.Name),
			pair("lastAiReplyAt", e.LastAiReplyAt),
		}}
		printCompact(&o)
	case "set-active":
		if len(args) < 3 {
			failCLI("set-active needs <name> and <true|false>")
		}
		name := args[1]
		flagVal := args[2]
		if flagVal != "true" && flagVal != "false" {
			failCLI("set-active needs <name> and <true|false>")
		}
		e, err := profiles.SetActive(name, flagVal == "true")
		if err != nil {
			failCLI(err.Error())
		}
		o := ordered{Pairs: []orderedPair{
			pair("id", e.Id),
			pair("name", e.Name),
			pair("active", e.Active),
		}}
		printCompact(&o)
	default:
		failCLI(`unknown command "` + cmd + `"`)
	}
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
