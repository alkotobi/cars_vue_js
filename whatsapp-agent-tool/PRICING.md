# Pricing data — schemas, formulas, maintenance

**When to read this file:** only when (a) the boss gives you a new freight/rate/benefit/clearance
number to store, or (b) `whatsapp-tool quote` returns `MISSING` (or `warnings`/`stale` flag old or
expired data) and you need to know which table is involved. Day-to-day quoting does **not** need this
file — `whatsapp-tool quote` does the math (AGENTS.md §13). Stale/expired data is no longer a hard
stop: the calculator quotes the best available figure and flags it (2026-10-14).

## 0. Where the data lives (MySQL)

Pricing is **shared across the whole team** and now lives in a MySQL database named **`ai`**
(not in the JSON files any more). Connection is read from `.env` (`DB_HOST/DB_PORT/DB_USER/
DB_PASSWORD/DB_NAME`, default DB name `ai`) — see `.env.example`. The original JSON catalog has
been **archived and removed** from the repo: it is at `backups/pricing-json-backup-*.tar.gz`
(gitignored). The one-time JSON→MySQL seed was done with the old Node `db:seed` script (no Go
equivalent); the schema lives in `db/schema.sql` — apply it with the `mysql` client if the DB is
ever recreated. That file is the **final shape for a fresh database** (idempotent
`CREATE TABLE IF NOT EXISTS`); it does not ALTER an already-populated database. To upgrade a
**live, populated** `ai` database to the current shape while keeping every row, run
`db/migrate-002-live.sql` then `db/migrate-003-live.sql` once each (both idempotent; back up
first with the repo's `go/cmd/dbmig backup <dir>`, apply with
`go run ./cmd/dbmig apply db/migrate-002-live.sql` / `db/migrate-003-live.sql`). Migration 003
links prices to suppliers (seeds `walle` + `jean fat boy` and attributes the existing 32 cars).

Supplier quotations (`suppliers`, `supplier_quotations`, `supplier_quotation_items`,
`supplier_quotation_item_prices`) also live in `ai` as a **procurement system linked to the
catalog**: `cars.supplier_id` (who supplies the current price), `car_price_history.supplier_id` +
`quotation_id` (price provenance), and `supplier_quotation_items.car_ref` (a line linked to the
`cars` row it matches). See `whatsapp-tool procurement` — the chain answers "who gives us this
price?".

- **To change anything you MUST use the write CLI** (you can no longer hand-edit the files):

  ```
  whatsapp-tool price set-car --json '{"carId":"livan-man","name":"LIVAN MAN","size":"small","condition":"new","costUsd":7000,"loadingPort":"NANSHA","colors":["grey"],"options":[{"name":"spare tire"}],"notes":"102 grey"}' [--supplier "walle" --quotation 3]
  whatsapp-tool price set-option --name "spare tire" --price 150
  whatsapp-tool price set-rate --value 253.4
  whatsapp-tool price set-benefits --wholesale 300 --retail 700
  whatsapp-tool price set-freight --json '{"loadingPort":"ANY (CHINA)","dischargePort":"ALGIERS / ANNABA (ALGERIA)","shippingLine":"HMM","container40ftUsd":5500,"etd":"2026-10-13","validTo":"2026-10-13"}'
  whatsapp-tool price set-clearance --json '{"carId":"livan-man","name":"LIVAN MAN","clearanceDza":800000,"aliases":["livan manual"]}'
  whatsapp-tool price set-customs-config --all-in 1000 --tvn 0
  whatsapp-tool price list-cars
  ```

  `set-car` appends to `car_price_history` automatically when the buy price changes. Every user-
  given supplier price is recorded **as a quotation first** (create + `add-item --car-ref` +
  `set-price` per colour, then `set-car --supplier "<name>" --quotation <id>`), which writes
  `cars.supplier_id` and stamps the history row with `supplier_id` + `quotation_id`. Every quote
  still goes through `whatsapp-tool quote` (§13) exactly as before — only its data source moved.

## 1. Tables (schemas)

- `cars` (+ `car_colors`, `car_options`) — cars with USD cost (current price only). `car_id` is the
  unique business key; rows also carry a numeric surrogate `id` that the child tables reference.
  - Fields: `carId`, `name`, `size` (`"big"` = 3 units per 40ft, `"small"` = 4), `condition`
    (`"new"`/`"used"`), `costUsd`, `supplierId` (who supplies the current `costUsd`; current suppliers: walle (1), jean
fat boy (2), lio (7) — list with `procurement suppliers`), `loadingPort` (e.g. `NANSHA`, `SHANGHAI`), `colors`
    (array of available colours, e.g. `["grey"]`; `[]` = unknown), `options`
    (array of `{ name, priceUsd? }` — a car feature; omit `priceUsd` to use the default from
    `options`, set it to override per car (`0` = free/included); `[]` = none. Never invent a
    feature or a price — ask the boss), `timestamp`.
  - Keep only the last 2 months; always the most recent entry per car. Two buy prices for the
    same car → keep the **highest**.
  - **FIVAN = LIVAN.** Clients write `FIVAN X3 Pro` for our LIVAN. **MAN ≠ AUTO** — two
    different cars (LIVAN MAN 7,000 / LIVAN AUTO 7,850); match the transmission the client asks
    for ("أوتوماتيك/boîte automatique" → AUTO, "مانيال/manuelle" → MAN).
- `car_price_history` — **append-only** buy-price history
  (USD): `{ carId, name, costUsd, timestamp, note? }`. This is the time-stamped
  audit trail of every price a car has had. **Never edit or delete** past entries — append a new
  one every time a buy price changes. The latest entry per `carId` must equal the current value
  in `cars`.
- `options` — default price per car option in **USD**:
  `{ name, priceUsd }`. Effective price = car override (`cars`
  `options[].priceUsd`) ?? this default. A car referencing an option with no price anywhere →
  `MISSING`, never guessed. Added to the car price in **every** mode except `clearance` (flat
  USD add, **no** benefit markup; converted to DZD together with the rest).
- `config` (key `benefits`) — per-car benefit in USD (single source, stored nowhere
  else): `wholesaleBenefitUsd`, `retailBenefitUsd`, `timestamp` (time series; the newest
  `effective_at` wins). Currently wholesale **300**, retail **700**.
- `freight` — 40ft container freight **my cost** in USD: `loadingPort`,
  `dischargePort`, `shippingLine` (stored as ids into the `loading_ports` / `discharge_ports` /
  `shipping_lines` lookups), `container40ftUsd`, `timestamp`, `etd` (vessel estimated time
  of departure), `validTo`, `freeTimeDays`. A quote is valid until **`validTo`, which is always
  `<= etd`**; if `validTo` is omitted, `etd` is used as the expiry. Always **+300 USD freight
  benefit per container** on top of my cost. Keep 2 months; when several valid quotes exist the
  calculator uses the **cheapest** one **to the requested discharge port** (`quote --port …`,
  default **ALGIERS** when the client names none; a
  newer quote only breaks a price tie) — the shipping line is never asked, `--line` is an
  operator-only override.
- `config` (key `rate`) — USD → DZD: `usdToDza`, `timestamp` (time series; newest `effective_at` wins).
- `clearance` (+ `clearance_aliases`) and `config` (key `customs`) — customs clearance per car in **DZA**: `currency` (`"DZA"`),
  `tvnIncluded` (always `false`), `allInRetailBenefitUsd` (currently **1000**), `updatedAt`,
  entries `{ carId | null, name, clearanceDza, aliases?, note?, timestamp }`.
  Lookup by `carId`/`aliases` (case-insensitive). `carId: null` = unmatched → ask the boss.
  Costs are stable: the 7-day rule applies only when new numbers arrive; trim to 2 months.

## 2. Maintenance rules

1. **Update on new input** — use the `whatsapp-tool price …` commands from §0: freight →
   `set-freight` (+300 applied in calculations); USD/DZD → `set-rate`; benefits → `set-benefits`
   (a new row in the time series); clearance → `set-clearance`; **car buy price → `set-car`
   (which appends the new price to `car_price_history` automatically)**.
2. **Trim to 2 months** — old rows can be deleted; the loaders only ever use the latest/valid ones.
3. **Most recent valid entry only** (by `effective_at`/`timestamp`); freight also must not be past `validTo`, and among valid quotes the **cheapest line to the requested port** wins (newest breaks a tie).
4. **Weekly validity** — data older than **7 days** is not valid. Stale → ask the boss, never
   guess. If the boss says "continue using last week's data" → re-issue the last valid values with
   the matching `set-*` command (fresh timestamp), then proceed.
5. **Ask, never guess** — a missing number → ask the boss; after the answer, run the `set-*`
   command first, then reply.
6. **Cache ↔ history invariant** — `cars.cost_usd`/`price_at`/`supplier_id` must match the latest
   `car_price_history` row of the same car (the cache never drifts from history). Verify on any
   suspicion or before quoting stale-looking data with `whatsapp-tool price check` (errors = cost
   mismatch; warnings = `price_at`/`supplier_id` deviation, e.g. just after a `set-car`).

## 3. Formulas (what `whatsapp-tool quote` computes — use to sanity-check its output)

- `optionsUsd` = sum of the selected car's resolved option prices (`car_options` override ??
  `options` default). **Only counted when the quote runs with `--options`** — by default the
  calculator **excludes options** and a quote never includes them unless the client explicitly
  asked (AGENTS.md §13). Added flat to every mode except `clearance`, before conversion, with
  **no** benefit markup.
- `freightPerUnitUsd = (container40ftUsd + 300) / units` — units = 3 if `size: "big"`, else 4.
- **Wholesale** (USD only, full containers only):
  - FOB = `costUsd + wholesaleBenefitUsd + optionsUsd` (goods at loading port — **no** freight)
  - CFR = `costUsd + wholesaleBenefitUsd + freightPerUnitUsd + optionsUsd`
  - Never mix them, never quote wholesale in DZA (the 12,674 vs 11,300 mistake, 2026-10-07).
- **Retail** (default DZA, landed):
  `(costUsd + retailBenefitUsd + freightPerUnitUsd + optionsUsd) × usdToDza`.
  USD only if the client explicitly asks (`retail-usd`); FOB retail drops `freightPerUnitUsd`.
- **All-in retail** ("يشمل كل الرسوم" / "compris tous les frais"): benefit = **1000**
  (`allInRetailBenefitUsd`), not 700 —
  `allInDza = (costUsd + 1000 + freightPerUnitUsd + optionsUsd) × usdToDza + clearanceDza`
  (clearance added **after** conversion). Wholesale never includes Algerian fees.
- **TVN** is never included (`tvnIncluded: false`) — say so if asked.
- **Customs clearance question** → `clearance` table `clearanceDza` (DZA); no match/`carId: null` →
  ask the boss.
- Centimes: **1 DZD = 100 centimes**; `248 مليون` = 248M centimes = 2.48M DZD. Always show the
  arithmetic.
