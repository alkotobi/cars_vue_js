# WhatsApp Agent Tool (MCP)

An MCP server that lets an AI agent control WhatsApp Web via go-rod + real Google Chrome.

## ⚠️ Important
- WhatsApp's Terms of Service prohibit automated/bot usage on WhatsApp Web. Use only for personal, authorized use (your own account).
- Automated sending can get your number banned. Use with care.

## Requirements
- **Google Chrome installed** (mandatory). WhatsApp Web refuses bundled Chromium with
  "WhatsApp works with Google Chrome 100+". The tool launches real Chrome.
- **Go 1.27+** (to build the binary). The old Node/TypeScript source is gone — backed up at
  `backups/node-code-backup-*.tar.gz`.

## Install
```bash
cd whatsapp-agent-tool/go
go build -o bin/whatsapp-tool ./cmd/whatsapp-tool
```
The binary lands at `go/bin/whatsapp-tool`; the CLI examples below use it as `whatsapp-tool`.
Re-run the same command after any code change.

## Shared business data (MySQL)
Pricing and the client FAQ are **shared across the whole team** and live in a MySQL database
named **`ai`**, so everyone quotes the same numbers. Set it up once per team (one shared DB):

```bash
cp .env.example .env      # then fill in MySQL host/user/password (DB_* vars; gitignored)
# schema lives in db/schema.sql (apply with the mysql client if recreating the DB);
# the one-time JSON → MySQL seed was done with the old Node `db:seed` script.
```

From then on the database is the source of truth:

```bash
whatsapp-tool price list-cars
whatsapp-tool price set-rate --value 253.4            # cars / rate / freight / benefits / clearance
whatsapp-tool faq search "هل يوجد تخفيض"               # client FAQ lookup
whatsapp-tool quote --car "LIVAN MAN" --mode retail-dza     # never adds options by default
whatsapp-tool quote --car "LIVAN MAN" --mode retail-dza --options   # only when the client asked
```

Supplier quotations are shared in `ai` too — a **procurement system linked to the catalog** (each
car knows its supplier: walle, jean fat boy, lio; quotation lines can carry a `car_ref` to the
car they match, and every buy price can be stamped with its supplier/quotation — unrelated to the
cars-app suppliers/inquiries; a quotation is timestamped and its price differs by colour):

```bash
whatsapp-tool procurement suppliers                           # ai suppliers, name ASC
whatsapp-tool procurement suppliers add --name "abi" --contact "+213..."   # create one
whatsapp-tool procurement quotations                          # quotations, newest first
whatsapp-tool procurement quotations show 5                   # header + items + per-colour prices
whatsapp-tool procurement quotations create --supplier-id 3 --ref "QT-1"   # quoted_at defaults to now
whatsapp-tool procurement quotations add-item --quotation-id 5 --car "LIVAN MAN" --big --new --qty 4
whatsapp-tool procurement quotations set-price --item-id 9 --color "white" --price 7000
whatsapp-tool procurement quotations set-status 5 --status accepted
whatsapp-tool procurement quotations delete 5                 # header + items + prices (FK cascade)
```

The original JSON catalog has been **archived and removed** from the repo
(`backups/pricing-json-backup-*.tar.gz`, gitignored). **Per-machine state stays local and is not shared**: the WhatsApp session (`data/whatsapp-profile/`),
the question queue (`ai_questions.json` / `unanswered_chats.json`), `profiles.json` and the run
reports. See `DEVELOPMENT.md` "Multiple WhatsApp accounts".

## Car stock (the cars app — a separate system)

Live car stock lives in the **cars app**, not in the `ai` database. The tool reads it over the
app's own HTTP API as a **dedicated AI user** (created by an admin in the app). Its credentials
and the cached login token live in the **OS credential store** — Windows **Credential Manager**,
macOS **Keychain**, Linux **Secret Service** — never in `.env` or a file. One-time setup:

```bash
# in .env (no secrets here)
CARS_BASE_URL=https://<cars-app-host>   # no trailing slash
CARS_TENANT=<tenant-folder/db-name>     # the URL path segment, e.g. merhab_cars

# seed the OS credential store with the AI user (logs in + caches the token)
whatsapp-tool stock creds --username <AI user> [--password <...>]
whatsapp-tool stock creds --status          # what's stored
```

```bash
whatsapp-tool stock                                   # available cars (limit 50)
whatsapp-tool stock "LIVAN"                           # match by car name
whatsapp-tool stock --car "LIVAN" --color white --limit 20
whatsapp-tool stock --vin "LW..." --status any        # any = available + sold
whatsapp-tool stock login                             # force a fresh login

whatsapp-tool suppliers                               # all suppliers, name ASC
whatsapp-tool suppliers "ALI"                         # match by supplier name
whatsapp-tool suppliers --name "ALI" --limit 20
whatsapp-tool suppliers add --name "abi"              # create a supplier

whatsapp-tool brands                                  # all brands, brand ASC
whatsapp-tool brands "TOYOTA"                         # match by brand name
whatsapp-tool brands add --name "TOYOTA"              # create a brand
whatsapp-tool brands add --name "TOYOTA" --logo logo.png   # create with logo (uploads first)
whatsapp-tool car-names                               # all car names, name ASC
whatsapp-tool car-names "GOLF" --brand "VW"           # match by car name and/or brand
whatsapp-tool car-names add --name "GS3 MID" --brand-id 17 --small   # create a model (--big/--small required)

whatsapp-tool inquiries                               # supplier quote requests, newest first
whatsapp-tool inquiries create --supplier-id 3 --ref "INQ-1"   # date defaults to now
whatsapp-tool inquiries add-car --inquiry-id 12 --car-name-id 4 --color-id 1 --qty 2 --new --target-price 12500
whatsapp-tool inquiries cars 12                       # the car lines of inquiry 12
whatsapp-tool inquiries delete 12                     # header + its lines (FK cascade)

whatsapp-tool clearance                               # per-car customs clearance (DZD, TVN excluded)
whatsapp-tool clearance --brand "GEELY"               # match by car name and/or brand
whatsapp-tool clearance set --car-name-id 39 --dza 950000 --small --new   # upsert (--big/--small, --used/--new required)

whatsapp-tool buy-payments --bill-id 1               # payments to the supplier of bill 1 (+ header)
whatsapp-tool buy-payments --supplier-id 3           # a supplier's payments across all bills
whatsapp-tool buy-payments add --bill-id 1 --amount 77000 --date "2026-10-11 12:00:00" --swift swift.pdf   # record a payment
whatsapp-tool buy-payments edit --id 3 --amount 60000 --date "2026-10-11 13:30:00"  # change one (options to keep swift/notes)
whatsapp-tool buy-payments delete --id 3             # remove one
```

## Tasks & reminders (`ai` database)

A **todo** is a task for the logged-in WhatsApp account (owner = `WA_PROFILE`); todos can nest
without limit as subtasks, carry a priority + optional deadline, and carry reminders. A parent is
`done` only when every child is done or cancelled (a new open child reopens it). Reminders have no
daemon — they surface in `whatsapp-tool todo due` and the `todosDue` block of the `pending` report.

```
whatsapp-tool todo add --title "call Sofia back" --priority high --due 2026-10-20 --requested-by "Sofia"
whatsapp-tool todo add --title "collect VIN" --parent 12            # a subtask
whatsapp-tool todo list                                             # open/urgent/due-soon first
whatsapp-tool todo list --overdue --tree                            # nested view
whatsapp-tool todo show 12                                          # todo + children + reminders + history
whatsapp-tool todo done 13                                          # parent auto-completes when all kids resolve
whatsapp-tool todo move 13 --parent 20                              # re-parent (cycle rejected)
whatsapp-tool todo remind add 12 --at "2026-10-20 09:30" --note "before noon"
whatsapp-tool todo due                                              # pending reminders now due
whatsapp-tool todo remind sent 4                                    # mark one sent after acting
whatsapp-tool todo delete 12                                        # removes the todo and its whole subtree
```


Later calls reuse the stored token (no re-login until the server rejects it). If the store is
empty, the command prints `status: ERROR, code: missing_credentials` — the agent then asks you
for the AI user's details. `CARS_API_TOKEN` in `.env` is an optional pin that skips the login.

The AI runs as its own **app user with a restricted role**. At every fresh login the command caches
the role and its permission names the app returns (from the `role_permissions`/`permissions`
tables), and `stock` refuses to run unless the role still grants `can_c_car_stock` (the app's
stock-view permission; `code: missing_permission`). Grant it in the app's RolePermissionsCrud,
then re-run `whatsapp-tool stock creds --username <AI user>`. The `admin` account should never be
used for this — it carries every permission.

Those same cached permissions **scope the columns** (the app gates them client-side in
`CarStockTable.vue`): a gated price column is dropped from the output unless the role has its
permission — `price_cell` ← `can_c_fob_price`, `freight` ← `can_c_freight`, `rate`/`cfr_da` ←
`can_c_cfr_prices`, `buy_price` ← `can_c_buy_price`. So the AI sees exactly the columns its role
allows, and only what it is permitted to.

`whatsapp-tool suppliers` reads the same app's **suppliers** table (fixed SELECT), sharing the AI
user's stored credentials/token; `whatsapp-tool suppliers add --name <name> [--contact ..]
[--notes ..]` creates one (a write). It mirrors the app's suppliers tab, which is gated on
`can_c_suppliers`, so both the read and the write refuse to run without that permission
(`code: missing_permission`).

`whatsapp-tool brands` and `whatsapp-tool car-names` read the app's **brands** and **cars_names**
tables (read-only, fixed SELECT), sharing the same credentials. They mirror the app's models tab,
gated on `can_purchase_cars` (`code: missing_permission` without it).

`whatsapp-tool inquiries` drives the **inquiry** system (supplier quote requests), the purchase-side
twin of buy bills — tables `inquiry` / `inquiry_details`, mirroring `buy_bill` / `buy_details`
(migrations `034`/`035`). There is no app screen for it yet. `list`/`cars` are read-only; `create`,
`add-car` and `delete` write (deleting a header removes its lines via the foreign key's
`ON DELETE CASCADE`). `add-car` requires `--color-id`, `--qty`, `--target-price` and `--used|--new`
(required columns); `create` fills `date_inquiry` with now. Gated on `can_purchase_cars` like the
models/buy tabs.

`whatsapp-tool clearance` reads the app's **clearance** table (per-car customs clearance in DZD,
TVN excluded; migration `036`), one row per model/variant joined to its model and brand. Its `set`
subcommand upserts a row (`INSERT ... ON DUPLICATE KEY UPDATE` on the
`(id_car_name, is_big_car, is_used_car)` unique key); `--car-name-id`, `--dza` and `--big|--small` +
`--used|--new` are required (big/small and used/new are part of the key). No permission gate —
clearance is free. This is the cars-app per-model clearance, not the `ai`-database pricing
`--mode clearance` (PRICING.md).

`whatsapp-tool buy-payments` reads the app's **buy_payments** table (payments we make to a
supplier against a purchase bill), joined to `buy_bill` / `suppliers` / `users`, mirroring
`BuyBillPaymentsView.vue`. `--bill-id N` returns the bill header (with `calculated_amount` = Σ
amount×QTY over `buy_details`, and `total_paid`) plus its payments; `--supplier-id N` returns that
supplier's payments across all bills. Gated on `can_pay_purchases` (payments to suppliers are
buy-cost figures). The writes all resync `buy_bill.payed` (Σ of that bill's payments) and so mutate
money data: `add` inserts a payment (`INSERT INTO buy_payments`, requiring `--bill-id`, `--amount`,
`--date` and the `--swift` document, uploaded to `buy_payment_swifts` first), `edit` updates one in
place (`UPDATE buy_payments`, `--id`/`--amount`/`--date` required; `--swift` optional, omitted keeps
the stored document), and `delete` removes one (`DELETE FROM buy_payments WHERE id = ?`).

**The writes.** Six commands change cars-app data: `whatsapp-tool brands add` creates a brand,
`whatsapp-tool car-names add` creates a model (`INSERT INTO cars_names (car_name, notes, is_big_car,
id_brand)`, name trimmed/upcased, `--brand-id` and `--big|--small` required), `whatsapp-tool
suppliers add` creates a supplier (`INSERT INTO suppliers (name, contact_info, notes)`, empty
fields stored `NULL`, gated on `can_c_suppliers`), and `whatsapp-tool inquiries
create/add-car/delete` create and remove inquiries (`add-car` requires `--color-id`, `--qty`,
`--target-price` and `--used|--new`; `create` defaults `date_inquiry` to now), `whatsapp-tool
clearance set` upserts a model's customs clearance (free — no permission gate), and `whatsapp-tool
buy-payments add/edit/delete` record, change or remove a supplier payment (`add` requires
`--bill-id`, `--amount`, `--date`, `--swift`; `edit` requires `--id`, `--amount`, `--date`; `delete`
requires `--id`) then recompute `buy_bill.payed`. Only the payment writes touch money (the bill's
paid total); none touches stock. `brands add` mirrors
`useBrands.js` `addBrand()`: with `--logo <path>` it uploads the file to the tenant's
`api/upload.php` first (multipart, `base_directory=files`, `destination_folder=uploads`) and stores
the returned relative path as `logo_path`; then it runs the fixed `INSERT INTO brands (brand[,
logo_path])`. `brands add`, `car-names add` and `inquiries` are gated on `can_purchase_cars` like
the models/buy tabs; `suppliers add` on `can_c_suppliers`; `clearance set` is ungated;
`buy-payments` (read + `add`/`edit`/`delete`) on `can_pay_purchases`. Each — every other cars command being
read-only — only runs when you invoke it explicitly. The AI user's role must grant the matching
permission; the `admin` account still should not be used.


## Link your account (one time, needs your phone)
```bash
WA_HEADLESS=0 whatsapp-tool keepalive start   # opens a visible Chrome window
```
A visible Chrome window opens with a QR code. On your phone:
**WhatsApp → Linked devices → Link a device** → scan.
If a 2-step verification PIN is requested, type it in that Chrome window.
(`Ctrl+C` here does not stop Chrome; use `whatsapp-tool keepalive stop`.)

The session is stored in `data/whatsapp-profile/` (a real Chrome user-data dir, so IndexedDB
auth survives restarts). Close the window with `Ctrl+C` when done — the session persists.

To force a fresh QR: stop the server and delete `data/whatsapp-profile/`.

## Verify it works
```bash
WA_SMOKE=1 go test -run TestSmokeLive ./internal/wa -v   # (from the go/ dir) attach + login state
```

## Run the MCP server
```bash
whatsapp-tool mcp      # stdio MCP server (the Go binary)
```
Headless by default. Set `WA_HEADLESS=0` to run with a visible Chrome window.

## Keep one browser for everything (the default)
Chrome is started **detached** and stays alive between tasks and even between MCP server
restarts. The first call that finds no browser pays the boot (~10–25 s); every later call
attaches to the same Chrome in well under a second.

```bash
whatsapp-tool keepalive start    # optional: start Chrome now, verify login, then idle
whatsapp-tool keepalive status   # is Chrome up? logged in? which port?
whatsapp-tool keepalive stop     # kill the shared Chrome (do this before linking)
```
`whatsapp_session_info` reports the same thing from inside a task.

| Env | Effect |
|---|---|
| `WA_PROFILE` | run a **second account**: own Chrome profile, CDP port, client queue (see below) |
| `WA_CDP_PORT` | pin the debug port (default: first free port in 9222–9252) |
| `WA_HEADLESS=0` | visible Chrome window |
| `WA_TRANSPORT=launch` | legacy in-process Chrome, closes when the server exits |
| `WA_CHROME_PATH` | explicit Chrome binary |

A logged-in WhatsApp profile sits in `data/whatsapp-profile/` while the browser is up — anyone
with access to this Mac can read the chats, and `whatsapp-tool keepalive stop` ends the session.

## Multiple WhatsApp accounts
One MCP server = one WhatsApp account, selected at **launch time by `WA_PROFILE`** — never by the
model. The tool names stay `whatsapp_*`; whichever account you started the session with is the
only one the session can drive (accounts use different Chrome profiles, ports and locks, so they
never conflict — `data/<profile>/` + `data/<profile>-cdp.json`; unset `WA_PROFILE` keeps the
classic `data/whatsapp-profile/` + `data/whatsapp-cdp.json` exactly as before).

`opencode.json` maps the caller's env into the server (already in this repo):

```json
{
  "mcp": {
    "whatsapp": {
      "type": "local",
      "command": ["/abs/path/go/bin/whatsapp-tool", "mcp"],
      "enabled": true,
      "environment": { "WA_PROFILE": "{env:WA_PROFILE}" }
    }
  }
}
```

Pick the account from the shell — one account per `opencode run`:

```bash
opencode run --auto "process the pending report"        # the default account
WA_PROFILE=business opencode run --auto "reply to him"  # the business account
```

In the interactive TUI the session uses whatever `WA_PROFILE` your shell had when you launched
opencode (default = main account). The model's own `whatsapp-tool pending` / `whatsapp-tool quote`
inherit the same variable, so reports always match the active account.

Per-account lifecycle:
```bash
WA_PROFILE=business WA_HEADLESS=0 whatsapp-tool keepalive start   # visible QR window for the 2nd account
WA_PROFILE=business whatsapp-tool keepalive start                 # warm its Chrome
WA_PROFILE=business whatsapp-tool keepalive status
WA_PROFILE=business whatsapp-tool keepalive stop
WA_PROFILE=business whatsapp-tool pending                         # that account's queue -> data/business-pending.json
```

### Registering accounts (`profiles.json`)
`profiles.json` is the account registry — one entry per linked account (`id` incremental, `name`
= the WA_PROFILE slug, `number`, `description`, `active`, `linkedAt`, `lastAiReplyAt`). New accounts
are linked **manually** (the old Node `connect` tool is gone and must not be run — see the strict
rules at the bottom). Registry commands:

```bash
whatsapp-tool profiles                          # list accounts (id / name / number / active / lastAiReplyAt)
whatsapp-tool profiles mark-reply business      # stamp "last time AI replied" = now
whatsapp-tool profiles set-active business false   # toggle the active flag
```

Everything per-account is namespaced: the Chrome profile, the CDP state file, the client queue
(`ai_questions.json`/`unanswered_chats.json` live in `data/<profile>-questions.json` /
`data/<profile>-chats.json` for a custom profile), and the run reports. What stays **shared**
across accounts: the pricing catalog + FAQ (shared MySQL `ai` database) and this repo's guides —
you are one business. Pin
`WA_CDP_PORT` per account (e.g. add it to the `environment` above) so attach never finds a
neighbour's Chrome.

## MCP config (opencode / Claude Desktop / Cline)
```json
{
  "mcpServers": {
    "whatsapp": {
      "command": "/Users/merhab/dev/browser_automation/whatsapp-agent-tool/go/bin/whatsapp-tool",
      "args": ["mcp"]
    }
  }
}
```

## Tools
- `whatsapp_is_logged_in` – check auth
- `whatsapp_get_chats` (limit, `mode: "screen" | "newest" | "unread"`, timeoutMs) – list chats.
  `screen` = the ~15 rows already rendered (fast, default); `newest` = walk the sidebar to collect
  the last N; `unread` = WhatsApp's own unread filter, also returns `unreadTotal`
- `whatsapp_open_chat` (nameOrPhone) – search and open a conversation
- `whatsapp_send_message` (text) – send to the currently open chat
- `whatsapp_get_last_messages` (limit) – read recent messages of the open chat
- `whatsapp_mark_as_read` – mark the open chat as read
- `whatsapp_session_info` – which browser/transport is in use (no boot)
- `whatsapp_identity` – the assistant's code-generated name (stable per account)

WhatsApp virtualises the sidebar, so anything beyond the first screen needs
`get_chats` with `mode: "newest"` or `"unread"`.

Workflow: `whatsapp_open_chat` → `whatsapp_send_message` / `whatsapp_get_last_messages`.

## Commands (prefer these over long tool loops)
```bash
whatsapp-tool keepalive start              # start Chrome once for the day
whatsapp-tool pending                      # pending scan: trim + promote + scan "ai questions" -> one report
whatsapp-tool triage 10                    # batch unread triage -> data/triage.json (one file, not 30 calls)
whatsapp-tool quote --car "COOLRAY BATTLE" --mode retail-dza   # pricing calculator (options excluded unless --options)
whatsapp-tool price set-rate --value 253.4 # MySQL: write pricing (set-car/rate/freight/benefits/clearance)
whatsapp-tool faq search "..."             # MySQL: look up a client FAQ
whatsapp-tool web fetch "<url>"            # read a page from the chat-AI source allowlist (`web list` prints it)
whatsapp-tool procurement quotations       # MySQL: supplier quotations per colour (see above)
whatsapp-tool stock "LIVAN"                # cars app: read-only stock lookup (see above)
whatsapp-tool identity                     # the assistant's code-generated name
whatsapp-tool profiles                     # list registered accounts
```
(`db:schema`/`db:seed` and `reply-all` were Node-only and are gone; the DB is already seeded.)

## Notes
- Selectors use WhatsApp's `data-testid` attributes (`chat-list`, `list-item-*`, `msg-container`,
  `conversation-compose-box-input`). If WhatsApp redesigns the UI, these may need updating.
- Non-Latin (e.g. Arabic) UI is handled: bidi marks are stripped and time formats include
  local words like `أمس` / `اليوم`. Message direction is detected from bubble geometry, because the
  Arabic (RTL) UI mirrors WhatsApp's `tail-out` marker.
- Only one instance can use `data/whatsapp-profile/` at a time (Chrome profile lock). Stop the
  shared Chrome (`whatsapp-tool keepalive stop`) before linking; the reverse is fine — the MCP
  server simply attaches to the link window's Chrome when it has a debug port.
- The shared Chrome is bound to one debug port chosen at start-up; we never attach to a browser
  another program started, so a foreign Chrome on 9222 is skipped, not hijacked.

## Talking about the chats
These conversations are **Algerian Derja** with clients of our car-export business in China.
- Dialect glossary + business context: [`ALGERIAN_DERJA.md`](./ALGERIAN_DERJA.md)
  (history: [`ALGERIAN_DERJA_ARCHIVE.md`](./ALGERIAN_DERJA_ARCHIVE.md))
- Standard client replies (derja): [`CANNED_REPLIES.md`](./CANNED_REPLIES.md)
- Agent operating rules: [`AGENTS.md`](./AGENTS.md)
- Pricing data schemas + formulas: [`PRICING.md`](./PRICING.md) · state-file schemas:
  [`STATE.md`](./STATE.md)
- Building / changing the code: [`DEVELOPMENT.md`](./DEVELOPMENT.md)
- **Prices are in centimes** (1 DZD = 100 centimes) unless the message says `da` / `دج`.
  `248 مليون` = 2.48M DZD.


**Strict rule: never delete any profile (from profiles.json, data/* for that profile, CDP state) without explicit confirmation from the user.**

**Strict rule: never register/link accounts programmatically. Do not run any account-linking/`connect` command. The user connects WhatsApp manually.**

**STRICTEST RULE: DO NOT delete, remove, drop, or modify profiles.json entries, profile folders (data/<name>/), CDP files, or any account-related state WITHOUT EXPLICIT USER CONFIRMATION. Do not run any account-linking/connect command.**
