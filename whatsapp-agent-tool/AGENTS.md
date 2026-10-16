# AGENTS.md — Operating the WhatsApp tools (read every session)

> **Personal context (user):** the user's wife's name is **Sofia** (user-stated 2026-10-10).

Day-to-day manual. Read it before using the `whatsapp_*` tools, replying to a client,
translating a chat or handling pricing/tracking questions. To **change the code**, read
[`DEVELOPMENT.md`](./DEVELOPMENT.md) — this file deliberately excludes that to keep the token
cost of running low.

> **Our chats are Algerian Derja (الجزائرية), not MSA.** We are **car dealers / exporters based
> in China**; every chat is with an **Algerian client**. See §10 for when to load
> [`ALGERIAN_DERJA.md`](./ALGERIAN_DERJA.md) and [`CANNED_REPLIES.md`](./CANNED_REPLIES.md).

---

## 0. Hard rules (do not break these)

1. **This is the user's real WhatsApp.** Never send test/dummy messages to real contacts. The
   only safe send target is the user's own self-chat ("Message yourself").
2. **Never delete or edit** anything in the user's chats unless explicitly asked.
3. **WhatsApp ToS forbids automation.** Keep volume low, human-like pacing (built in), never run
   bulk/mass messaging. Ban risk is real.
4. **One process at a time.** `data/whatsapp-profile/` is a Chrome profile with a lock file. If a
   `keepalive`/link Chrome window is already using the profile, the MCP server cannot start a
   second one. Close it first.
5. **Never log message content to stdout.** The MCP server speaks JSON-RPC on stdout; logging goes
   to stderr. CLI commands write payloads to data files and print summaries only.
6. **Never commit** `data/whatsapp-profile/` (session auth) — already gitignored.
7. **Messages are Algerian Derja, prices are centimes.** Never report a translated price without
   the arithmetic **when the arithmetic is asked for** (default is price-only, §0.14), never guess
   a dialect word (§10).

8. **Always use Arabic numerals (0123456789) when answering in Arabic.** Never use Eastern Arabic /
   Indic numerals (٠١٢٣٤٥٦٧٨٩) in any Arabic or Derja reply — this includes prices, quantities,
   years, phone numbers, model codes, percentages, everything. Example: write "210,3" not
   "٢١٠,٣", write "2026" not "٢٠٢٦".
9. **AI disclosure line — append to every outgoing client message:**
   - Clear-Arabic answers (FAQ, holding reply, voice/call reply, pricing answers):
     `هذا الرد مولد بالذكاء الاصطناعي`
   - Replies built from a `CANNED_REPLIES.md` template: that file's own line
     `ملاحظة: هاد الرد تولّد بالذكاء الاصطناعي (AI).` (its style rule 5)
    - Group questions (§14): Arabic `هذا الرد مولد بالذكاء الاصطناعي` /
      English `This is an AI-generated message` / Chinese `这是一条由AI生成的消息`

10. **Before any write/INSERT, ask the user for every obligatory field you don't already have,
   and wait for the answer. Never guess or accept a hidden default. Ask first, post second**
   (full list in "Strict rules" at the end, and §15b/§15c/§15d).

11. **Search order — always the `ai` database first, then the cars app.** When looking anything up,
   query the shared **`ai`** DB first (`whatsapp-tool faq` / `quote` / `price` / `procurement`),
   and only then the **cars app** (`whatsapp-tool stock` / `suppliers` / `brands` / `car-names` /
   `inquiries` / `clearance` / `buy-payments`). `ai` is the source of truth for pricing,
   procurement (suppliers, quotations, documents) and FAQs; the cars app is the operational/stock
   system. Never answer from the cars app when the same question is answered by `ai`.

12. **Default data source is the `ai` database — the cars-app tenant `mig_26` is opt-in.**
   Answer every lookup from the shared **`ai`** DB. Do **not** query the cars app
   (`whatsapp-tool stock` / `suppliers` / `brands` / `car-names` / `inquiries` / `clearance` /
   `buy-payments`, which read the `mig_26` tenant) unless the user explicitly says **"find in
   mig_26"** (or otherwise names mig_26) for that request. Never fall back to `mig_26` on your own.

13. **Never guess a price. Ask the client for every detail needed to give an exact price before
   quoting.** A price is only exact when the inputs that drive the formula are known — the
   **color** (price can differ per color) and the **price basis** (**CFR** landed vs **FOB** vs
   **TTC**/all-in, retail vs wholesale). If either is missing from the client's request, ask
   (clear Arabic) and wait — do not quote a default, do not fold unknown costs in. Then run
   `whatsapp-tool quote` with those exact inputs. Full flow in §13.
   **Destination port defaults to ALGIERS.** When the client does not name a port, quote with
   `--port ALGIERS` and **say so in the reply** (e.g. "واصلة ميناء الجزائر") — never assume silently.
   A port the client names always wins.
   **Shipping line is NEVER asked.** For the client's port the calculator auto-picks the
   **cheapest valid freight quote** (`--port <ALGIERS|ANNABA|ORAN|…>`; a newer quote only breaks a
   price tie). `--line` exists only as a manual override for the operator. Never ask the client or
   the boss which line — just run `quote` with `--port` and no `--line`.

14. **By default give the price only — do not show how it was calculated.** Quote the final
   `price.usd` / `price.dza` / `price.dzaWords` as plain numbers. Show the arithmetic / breakdown
   **only when the user or client explicitly asks** ("how did you get that", "show the
   calculation"). This overrides the "always show the arithmetic" wording elsewhere (§0.7, §13)
   for the default case.

15. **When quoting a client, always round the price UP.** Take the calculator's final figure and
   round **up** (never down, never to the nearest) to the nearest **10,000 DZD** (and to the
   nearest **100 USD** when quoting USD). E.g. 3,269,810 DZD → **3,270,000 DZD**; 12,903.75 USD →
   **12,910 USD**. This applies **only to what the client is told** — the exact calculator value
   is what I report to the user and what any breakdown shows (so a "how did you get that" answer
   still shows the raw number, with the rounded client-facing price stated alongside). Never round
   the stored data or the calculator inputs.

---

## 1. How the tools behave

> **The CLI is now Go.** Every `npm run <x>` below is replaced by the Go binary:
> `npm run <x> -- args` → `whatsapp-tool <x> args` (the binary lives at `go/bin/whatsapp-tool`;
> from the repo root that is `go/bin/whatsapp-tool <x> args`). The old Node/TS source is gone
> (backed up at `backups/node-code-backup-*.tar.gz`).

One chat is open at a time, shared across all `whatsapp_*` calls: `open_chat` → then
`send_message` / `get_last_messages` / `mark_as_read`. Chrome stays alive between sessions (fast
attach ~0.5 s). If a call feels slow, first call `whatsapp_session_info` — it does **not** boot
Chrome. Run `whatsapp-tool keepalive start` once at the start of a long working day so the browser
is warm. `whatsapp_get_chats` has `mode: "screen" | "newest" | "unread"` (screen = the ~15 rendered
rows, newest/unread = scroll the sidebar). Tool specs live in `DEVELOPMENT.md` §4.

**Your identity.** The tool mints a unique, irreplicable name for the assistant **by code**
(`crypto`), never by the model, and persists it per account. Read it with `whatsapp_identity`
(also embedded in `whatsapp_session_info`) or `whatsapp-tool identity` — it is stable across
sessions, so treat it as your own name and never invent a different one.

**Second account?** One `whatsapp` MCP server, account chosen by the shell, never by the model:
`opencode.json` injects `{env:WA_PROFILE}`, so `WA_PROFILE=business opencode run --auto "…"` makes
this session drive **only** the business account (tool names stay `whatsapp_*`; there is no
`whatsapp_b_*`). The account registry lives in `profiles.json` (id/name/number/description/active/
linkedAt/lastAiReplyAt): list it with `whatsapp-tool profiles`, stamp a reply with
`whatsapp-tool profiles mark-reply <name>`, flip it with `whatsapp-tool profiles set-active <name>
<true|false>`. New accounts are linked **manually** by the user (never programmatically — see the
strict rules). Every account keeps its **own** Chrome profile, own state queue (`pending`/`triage`
reports, `ai_questions.json`) — pricing/guides stay shared. Prefix the same `whatsapp-tool ...`
commands with the profile: `WA_PROFILE=business whatsapp-tool pending` →
`data/business-pending.json`. The `profile` field of `whatsapp_session_info` tells you which
account you are on. See README "Multiple WhatsApp accounts" / DEVELOPMENT.md §10.

**Three commands replace long tool loops — prefer them:**

- `whatsapp-tool pending` — pending scan (trim + promote + scan the `ai questions` group) and prints
  one report with everything needed to act. **Run first, every turn** (§14b).
- `whatsapp-tool triage [N] [msgs] [--no-read]` — opens the N newest unread chats, reads their last
  messages (marks read unless `--no-read`), writes `data/triage.json`, prints a one-line summary.
- `whatsapp-tool quote --car "<name>" --mode <mode>` — the **only** allowed way to compute a price
  (§13).
- `whatsapp-tool freight [--port <PORT>] [--line <LINE>]` — shipping rates to a port (no car
  needed); answers a general "how much is shipping to X" question (§13b).

---

## 5. Task playbooks

**Start of turn (mandatory)** — `whatsapp-tool pending` FIRST, then act on its report: send every
`toSendToGroup` question (format in `STATE.md` §3), then answer each `dataReady` chat one by one
(§14b). Only after that, handle new incoming messages.

**Triage unread — the fast way**
1. `whatsapp-tool triage 10` → read `data/triage.json` (one Read, not 30 tool calls).
2. Summarize per chat (sender direction + time + gist). Never auto-reply without being asked.

**Triage unread — manual** (only when a chat-by-chat pass is needed anyway)
1. `whatsapp_get_chats` with `mode: "unread"`, `limit: 30`.
2. For each: `whatsapp_open_chat(name)` → `whatsapp_get_last_messages(10)` →
   `whatsapp_mark_as_read`. Summarize; never auto-reply without being asked.

**Daily digest of unread** — `whatsapp-tool triage 30 5 --no-read` → read `data/triage.json` →
one summary. (Manual fallback: `get_chats(mode:"unread")`, then open/read per chat.) Skip
marking as read unless asked — it mutates unread state.

**"Read the last 5 messages from <chat>"** — `open_chat(name)` → `get_last_messages(5)`.
Report with sender direction and time.

**"Send X to <chat>"** — `open_chat(name)` → **verify the name it matched** (`get_chats` first
if ambiguous) → `send_message(X)`. Confirm to the user what you sent and to whom.

**Start-number chat** — `open_chat("+2136...")`. `false` = not a WhatsApp contact of this account.

---

## 10. Working with the Algerian Derja (mandatory)

Every chat is **Algerian Derja** in Arabic script, mixed with transliterated car names, French
parts vocabulary and English trim terms. Standard Modern Arabic is not enough — literal MSA
translation produces wrong meanings.

**When to load the glossary:** read [`ALGERIAN_DERJA.md`](./ALGERIAN_DERJA.md) before translating
prices, models, negotiation, order details or any commercially important line, and
[`CANNED_REPLIES.md`](./CANNED_REPLIES.md) before composing any reply. For quick routine triage
("who wrote what") summarize straight from the messages and load the glossary only when a word
actually carries meaning you don't know.

**Before answering any "what does this chat say / what is it about" request:**

1. Read `ALGERIAN_DERJA.md` first if the chat contains commercially important lines.
2. Read the messages; for every meaningful word **not** in the glossary, note it as an open
   question.
3. Answer with: literal line-by-line translation + a short **business** summary (what is being
   bought/sold, year, price in **centimes**, what is still unanswered).
4. End with a short "words I don't know yet" list and **ask the user to explain them**.
5. When the user explains, add it to `ALGERIAN_DERJA.md` (§3 confirmed / §5 inferred) **and** a
   change-log line in [`ALGERIAN_DERJA_ARCHIVE.md`](./ALGERIAN_DERJA_ARCHIVE.md). That is the
   whole point of the file: it gets better every session.

**Hard rules for translation**

- **Never invent a meaning.** Give the literal reading, mark it as a guess, and ask.
- **Money = centimes** unless "da"/"دج"/"دينار" appears. `248 مليون` = **2.48M DZD**, not 248M
  DZD. Always show the arithmetic. See `ALGERIAN_DERJA.md` §2.
- Keep the original Arabic next to the translation for anything commercially important (prices,
  years, model names).
- Distinguish **inferred** (your deduction) from **confirmed** (user explained).
- Politeness/insults are common in derja ("يا أخي", "خويا"). Report tone factually — don't
  soften or inflate it.

**Hard rules for replying**

- Reply in **Arabic script** (clear, friendly Arabic — not stiff formal MSA, not broken derja),
  with **1–3 light emoji**, warm but professional. Never dry.
- **Never sound like we're refusing to sell.** Frame the inspection policy as a service promise
  ("we want you to buy with total confidence"), then invite the client warmly.
- Use a template from `CANNED_REPLIES.md` and keep its structure; append the AI disclosure line
   (§0.10).
- Never quote a price, spec, availability or payment term the user has not stated — use the
   file's `<...>` placeholders and have the user fill them.
   - Confirm before sending, unless the user has explicitly authorized auto-answering.

---

## 11. Client FAQ (auto-reply)

Repeated client questions live in the shared MySQL database **`ai`**, table **`faq`** (was
`faq.md`). Search it with `whatsapp-tool faq search "<question>"` (returns JSON matches); add an
entry with `whatsapp-tool faq add --q "<Q>" --a "<A>" [--context "<note>"]`.

1. **Check first** — `whatsapp-tool faq search "<question>"`; exact or very similar match?
2. **If found** — reply with the saved answer in **clear Arabic** (not derja), natural and to the
   point, + the AI disclosure line (§0.10; clients write Arabic, so the Arabic version).
3. **If not found** and you have no approved answer → reply exactly
   `ما عنديش المعلومات الكافية سأسأل فريق العمل وأجيبك متى تحصلت على المعلومات ` + the Arabic AI disclosure line.
   *An "approved answer" = a `faq` table entry, a `whatsapp-tool quote` result with `status: OK` (§13),
   or something the user explicitly approved.*
4. **Save when told** — "this is common" / "save to faq" / "add to faq" →
   `whatsapp-tool faq add --q "<Q>" --a "<A>" [--context "<note>"]`.
5. **Never guess.** Only the `faq` table, the calculator, or explicitly approved answers.

---

## 12. Voice messages, Calls & Missed Calls

Apply this **before** §11, before any answer, before creating questions/chats, before anything
else. Trigger if **any** of these is true for the chat: a voice message was received; the last
message is a voice message; the client called; there is a missed call (incoming or outgoing);
the chat shows a call-related system message.

When triggered:
- Do **not** transcribe, translate, guess or call back.
- Reply with exactly: `لا أستطيع فهم الرسائل الصوتية بالعربية، من فضلك اكتب سؤالك نصيًا حتى أتمكن من مساعدتك.`
   + the Arabic AI disclosure line (§0.10).
- **Never** create `ai_questions.json` / `unanswered_chats.json` entries for calls or missed
  calls. Send that message and stop.

---

## 13. Pricing — always via `whatsapp-tool quote`

> **Executable client-reply rules live in one place.** The cars-app chat AI's channel-agnostic
> reply rules (tool usage, never-invent-a-price, freight, ALGIERS default, web restrictions,
> Markdown tables, language match, numerals, disclosure) are defined once in
> [`chatbot/rules/client-reply.md`](./chatbot/rules/client-reply.md) and injected into both agent
> personas and the prompt builder automatically. When a rule or a CLI flag changes, edit that file
> (and the operator rules here only where they differ).

All pricing data sits in the shared MySQL database **`ai`** (cars, options, benefits, freight,
rate, clearance). Field-by-field schemas, the write CLI and the formulas live in
[`PRICING.md`](./PRICING.md) — read it **only** when storing new numbers or when the calculator
reports `STALE`/`MISSING`. **Never hand-edit the JSON files** — they are a frozen backup; every
change goes through `whatsapp-tool price set-*`.

**You never compute a price yourself.** Run the calculator and quote its number. But first, never
guess: a price is only exact when the **color** and the **price basis** (**CFR** landed vs
**FOB** vs **TTC**/all-in, retail vs wholesale) are known — if the client's request is missing
either, ask (clear Arabic) and wait; do not quote a default. **The destination port defaults to
ALGIERS** when the client does not name one: quote with `--port ALGIERS` and say so in the reply
("واصلة ميناء الجزائر") — never assume silently. **By default give the price only** (final
`price.usd` / `price.dza` / `price.dzaWords`); show the arithmetic / breakdown **only when the
user or client explicitly asks** (§0.14).

```
whatsapp-tool quote --car "<name>" --mode <mode> [--port ALGIERS] [--line CMA] [--options]
```

- modes: `wholesale-fob`, `wholesale-cfr`, `retail-dza` (default), `retail-usd`,
  `retail-usd-fob`, `allin-dza`, `clearance`. `--list` = the catalog.
- **Shipping line (STRICT):** never ask which line. Pass the **`--port`** (ALGIERS, ANNABA, ORAN,
  …; default **ALGIERS** when the client names none) and `quote` picks the **cheapest valid**
  freight quote to that port automatically (a newer quote only breaks a price tie). `--line` is a
  manual override for the operator only.
- **Options rule (STRICT):** the calculator NEVER adds car options to a quote by
  default. Only re-run with `--options` when the **client explicitly asked** about
  specific options — never include them "just because the car has them". If you
  quoted without options, never present options as part of that price.
- FIVAN → LIVAN, and **MAN ≠ AUTO** (different cars, different costs) — if the result is
  `AMBIGUOUS`, re-run with the exact name.
- Check the `formula` line matches what the client actually asked for (FOB vs CFR/landed is the
  classic trap).
- **Rounding (STRICT, §0.15):** before telling a client a price, round it **up** to the nearest
  10,000 DZD (100 USD if quoting USD). Exact value only in what I report / breakdowns.

**Result statuses**

| status | what to do |
|---|---|
| `OK` | Quote it: `price.usd` / `price.dza` / `price.dzaWords` — **price only by default**; show the arithmetic **only when asked** (§0.14). |
| `MISSING` | A number is not in the files / no freight for that port → never guess. `carId: null` clearance → ask me. Reply `سأطلب من المدير وأعود إليك` + AI line, log it (§14). |
| `AMBIGUOUS` | Several cars matched → re-run with the exact name (see above). |

**No more `STALE` (changed 2026-10-14).** Data older than 7 days, and freight quotes that are
expired, are **no longer a hard stop**: `quote` always returns `OK` with the best available figure
(the **cheapest applicable shipping line**) and records the age in `stale`, in `warnings`, and per
line in `lines[].expired` / `lines[].stale`. **Give the number anyway**, but add a short note that
it may have changed (and offer to confirm with the team). `lines[]` carries one price per
applicable shipping line (the cheapest is the headline `price`; each entry has `chosen`/`expired`/
`stale`). Only `MISSING` (no figure at all / no freight to that port) and `AMBIGUOUS` still defer.

**Inquiry flow (WhatsApp)**

1. **First question**: retail or wholesale? (clear Arabic.) Wholesale → tell them: **full
   containers only, and USD only** (never DZA).
2. **Collect every exact-price input before quoting** — ask the client (clear Arabic) and wait
   until both are known: the **color** and the **price basis** (CFR/landed vs FOB vs TTC/all-in).
   Never quote a default or assume a missing one (§0.13). **The destination port defaults to
   ALGIERS** when the client names none — quote with `--port ALGIERS` and say so in the reply
   ("واصلة ميناء الجزائر"); a port the client names wins. **Never ask about the shipping line** —
   the calculator picks the cheapest valid line to the port automatically.
3. Shipping line is auto-picked (cheapest valid **to the client's `--port`**); never ask the
   client or me for one. `--line` is a manual override for the operator only.
4. When I give new freight/rate/benefit/clearance numbers → read `PRICING.md`, update the files
   first, then quote.
5. **TVN** never included: `لا، السعر لا يشمل TVN` + AI line.
6. Customs clearance question → `--mode clearance`.
7. §12 (voice/calls) applies before all of this; §11 reply rules apply (clear Arabic + AI
   disclosure line).

---

## 13b. Shipping / freight rates — `whatsapp-tool freight`

A **general shipping question** ("كم سعر الشحن إلى الجزائر", "how much is shipping to Oran") names
a **port**, not a car, so `quote` cannot answer it. Use the read-only freight list instead:

```
whatsapp-tool freight [--port <ALGIERS|ANNABA|ORAN|BEJAIA|SKIKDA|MOSTAGANEM|GHAZAOUET|DJEN DJEN>] [--line <LINE>]
```

- **Omit `--port`** (or pass the country, e.g. `--port ALGERIA`) to list **every** Algerian port.
  `--port ALGER` is read as `ALGIERS`. An unknown port name falls back to listing every port.
- It returns, per shipping line, the best quote to the port, **cheapest first**: `container40ftUsd`
  (our cost), `containerChargeUsd` (cost + the 300 USD freight benefit the price carries),
  `perCarBigUsd` (charge ÷ 3) and `perCarSmallUsd` (charge ÷ 4), plus `etd`/`validTo` and an
  `expired` flag. A line serving the port through a combined berth (e.g. "ALGIERS / ANNABA
  (ALGERIA)") is folded onto the same line.
- The cheapest **valid** line is the headline answer; if every line for that port is expired it
  still returns the cheapest and flags `expired: true` — give the figure and note it may have
  changed (never defer).
- **Permission:** the container rate is our cost basis, so `freight` requires **`can_c_cost`**. A
  cost-restricted group is refused (`status: ERROR, code: forbidden_permission`) and must defer the
  shipping question to the team (or take a landed `quote` instead). Unlike `quote` there is no
  redaction path — the whole payload is the cost figure.
- **Never ask the client which shipping line** — the list is ordered cheapest-first; quote the top
  valid line.

---

## 14. Question tracking & pending chats

Two state files, both owned by the scripts — the schemas, creation/dedup rules and the group
message template are in [`STATE.md`](./STATE.md); read it **before creating or updating any
entry**. `whatsapp-tool pending` does the trim/archive/promote/group-scan mechanically.

**Workflow**

1. Client asks something → §12 (voice/call) first, then §11 (FAQ).
2. FAQ hit → answer from it, no queue entries.
3. Needs data that is missing/stale → reply `سأطلب من المدير وأعود إليك` + AI line; create the
   `ai_questions.json` entry (`answered=false`) + the `unanswered_chats.json` entry
   (`waiting_for_data`), linked by `questionIds`; post the AR/EN/ZH question to the `ai questions`
   group (`STATE.md` §3) and set `sentToGroup=true`.
4. Answer arrives (boss or group) → the next `whatsapp-tool pending` detects it, promotes the chat
   to `data_ready` and prints it with the answers.
5. Answer the client in clear Arabic + AI line (run the matching `whatsapp-tool price set-*` command
   first if the answer carries new numbers), then mark the chat `replied` → `closed`.
6. Always keep both files trimmed (2 months) — `whatsapp-tool pending` does it and archives finished
   entries to `data/state-archive.json`.

## 14b. Pending scan (must run first, every turn)

**Before processing any new incoming client message**, run `whatsapp-tool pending`
(`--state` = trim + promote only, no browser; plain = also scans the group).

Act on the report only — **do not Read `ai_questions.json` / `unanswered_chats.json`**; the
report already carries, per `dataReady` chat, the `needed` items and every linked question with
the boss's answer, plus the unsent group questions with their text. (`data/pending.json` holds
the same payload plus the client's original message, for the rare case you need it.)

Then, one chat at a time:
1. Open the chat (`chatKey`, falling back to `chatName`) and read the last messages — check the
   client has not written something new meanwhile.
2. Verify the answers actually cover the `needed` items (judgment, not mechanical).
3. Reply in clear Arabic + the AI disclosure line (§0.10).
4. Mark the chat `replied`, then `closed`.
5. `waiting` chats need no action; `toSendToGroup` items get posted first (`STATE.md` §3).

Never handle a new client message until this is done.

## 14c. Team group "ai questions"

Team answers arrive in the WhatsApp group **`ai questions`** as `[ID: profile#n]` replies (the
profile prefix keeps ids unique across accounts). Posting, formatting, tracking fields
(`sentToGroup`, `groupLastScanAt`, `answeredBy`) and the scanning rule are in `STATE.md` §3 — the
scan itself runs inside `whatsapp-tool pending`, never by hand.

## 15. Car stock — live availability (the cars app)

Availability questions ("do you have <car>", "how many <color> <car> left") are answered from the
**cars app**, a separate system from the `ai` database. Use the read-only command — **never** guess
stock:

```
whatsapp-tool stock "LIVAN"                       # available cars matching the name
whatsapp-tool stock --car "LIVAN" --color white --limit 20
whatsapp-tool stock --vin "LW..." --status any    # any = available + sold
whatsapp-tool stock --port "Alger" --include-hidden
whatsapp-tool stock --status sold --this-month              # sold in the current month
whatsapp-tool stock --status sold --since 2026-10-01 --until 2026-10-31   # sold in a date range
```

`--status` is `available` (default), `sold` or `any`. The date filters (`--this-month`,
`--since`/`--until`, `YYYY-MM-DD`, the **sell** date) apply only with `--status sold` and are
rejected otherwise.

Reading the JSON: `count` + `cars[]`; each row has `id`, `vin`, `car_name`, `color`, `status`
(`Available`/`Sold`), `loading_port`/`discharge_port`, `warehouse_name`, `container_ref`,
`buy_bill_ref`, `sell_bill_ref`, `sell_date`, `client_name`. The **price columns are
permission-scoped** and
appear only when the AI user's role has the matching permission (they are dropped from the output
otherwise): `price_cell` (FOB, `can_c_fob_price`), `freight` (`can_c_freight`), `rate` + `cfr_da`
(CFR, `can_c_cfr_prices`), `buy_price` (`can_c_buy_price`, the purchase cost from `buy_details`).
Quote availability from this output only; for the **price** still use
`whatsapp-tool quote` (§13) — `price_cell` here is raw stock data, not the calculator's answer.

**First-time setup / missing credentials.** The admin creates a dedicated **AI user** in the cars
app. Credentials and the cached token live in the **OS credential store** (Windows Credential
Manager / macOS Keychain / Linux Secret Service), not in `.env` (which holds only the app URL and
tenant).

1. If everything is already set up, just run the command — it reuses the stored token; `whatsapp-tool
   stock creds --status` shows what's stored.
2. If `whatsapp-tool stock` returns `status: ERROR, code: missing_credentials` (or `not_configured`),
   **ask the user** for the AI user's username and password (and the app base URL + tenant if
   those are missing too). Save the non-secret URL/tenant to `.env` (`CARS_BASE_URL` /
   `CARS_TENANT`), then seed the store once:
   ```
   whatsapp-tool stock creds --username <user> [--password <pw>]
   ```
(or pipe the password: `echo "<pw>" | whatsapp-tool stock creds --username <user>`). The command
    logs in and caches the token immediately — if the login is rejected it stores nothing. Never
    invent credentials; never ask a client.
3. **Read-only, with six exceptions.** Nearly every cars-app command only reads — the six that
   write are `whatsapp-tool brands add` (§15c, create a brand, optionally with a logo),
   `whatsapp-tool car-names add` (§15c, create a model), `whatsapp-tool suppliers add` (§15b,
   create a supplier), the `whatsapp-tool inquiries create/add-car/delete` commands (§15d, the
   supplier-quote-request system), `whatsapp-tool clearance set` (§15e, set a model's customs
   clearance, no permission gate), and the `whatsapp-tool buy-payments add/edit/delete` commands
   (§15f, record/change/remove a payment to a supplier). All six run only when you invoke them, and
   only the payment writes touch money (recomputing the bill's paid total).
4. **Permissions:** the AI logs in as a real app user whose role carries **only** `can_c_car_stock`
   (the app's stock-view permission — the account is governed by the app's `roles` +
   `role_permissions` tables, never the `admin` account). The command caches the role/permissions
   at login and refuses to run if that permission is missing: `status: ERROR, code:
   missing_permission`. If that happens, tell the user to re-grant `can_c_car_stock` to the AI
   user's role in the app (RolePermissionsCrud) and then re-run `stock creds --username <user>`.
   The same cached permissions scope the **columns**: gated price columns the role lacks are
   removed from the rows (e.g. an `ai` role without `can_c_buy_price` sees no `buy_price`). This
   mirrors the app's `CarStockTable.vue`; it is a visibility rule, not a write.

## 15b. Suppliers — the cars app's supplier list

Supplier questions ("who are our suppliers", "do we have a supplier called X", "what's their
contact") are answered from the **same cars app**, with the same AI user / stored credentials as
`stock`:

```
whatsapp-tool suppliers                       # all suppliers, ordered by name
whatsapp-tool suppliers "ALI"                 # match by name
whatsapp-tool suppliers --name "ALI" --limit 20
```

Reading the JSON: `count` + `suppliers[]`, each row `id`, `name`, `contact_info`, `notes`. Never
invent a supplier or its contact details — only report what the command returns.

**Create a supplier (a write).** `whatsapp-tool suppliers add` creates a supplier, mirroring
`SuppliersView.vue` `addSupplier()` (`INSERT INTO suppliers (name, contact_info, notes)`). An empty
`--contact`/`--notes` is stored `NULL`. Run it only when asked, and confirm the name first:

```
whatsapp-tool suppliers add --name "abi"                              # name only
whatsapp-tool suppliers add --name "abi" --contact "+213..." --notes ".."   # with details
```

It answers `{status:"OK", action:"supplier-add", id, name}`. The name is unique, so a duplicate is
rejected (`code: query_failed`). Ask the user for the name (and any obligatory field) before
posting — see "Strict rules".

**Permission:** the app's suppliers tab is gated on `can_c_suppliers` (in `CarsView.vue`), so both
commands refuse to run without it (`status: ERROR, code: missing_permission`) — grant it in the
app's RolePermissionsCrud, then re-run `whatsapp-tool stock creds --username <user>`. Setup /
credentials are shared with §15 (run `stock creds` once).

## 15c. Brands & car names — the cars app's catalog

Model/catalog questions ("what brands do we carry", "is there a car named X", "is X a big car",
"which brand is X") come from the **same cars app**, read-only, with the same credentials as
`stock`:

```
whatsapp-tool brands                          # all brands, ordered by name
whatsapp-tool brands "TOYOTA"                 # match by brand name
whatsapp-tool car-names                       # all car names, ordered by name
whatsapp-tool car-names "GOLF" --brand "VW"   # match by car name and/or brand
```

Reading the JSON: `brands[]` rows are `id`, `brand`, `logo_path`; `car_names[]` rows are `id`,
`car_name`, `notes`, `is_big_car` (0/1), `id_brand`, `brand`. Never invent a brand or model — only
report what the commands return.

**Permission:** the app's models tab is gated on `can_purchase_cars` (in `CarsView.vue`), so both
commands refuse to run without it (`status: ERROR, code: missing_permission`) — grant it in the
app's RolePermissionsCrud, then re-run `whatsapp-tool stock creds --username <user>`. Setup /
credentials are shared with §15 (run `stock creds` once).

**Create a brand (a write).** `whatsapp-tool brands add` creates a brand, mirroring the app's
"Add brand" dialog (`useBrands.js` `addBrand()`). Along with `suppliers add` (§15b) and the
`inquiries` writes (§15d) it is one of the cars-app commands that change data — run it only when
asked, and confirm the name first:

```
whatsapp-tool brands add --name "TOYOTA"                    # name only
whatsapp-tool brands add --name "TOYOTA" --logo ./logo.png  # uploads the logo, then links it
```

`--logo` is a **local file path** on this machine (not a URL): the command uploads it to the
tenant's `api/upload.php` (multipart, `base_directory=files`, `destination_folder=uploads`) and
stores the returned relative path as the brand's `logo_path` — the same flow as the fixed web form.
It answers `{status:"OK", action:"brand-add", id, brand, logoPath}`. It is gated on
`can_purchase_cars` (like the models tab), so an `ai` role without it is refused. There is no
delete/rename here — only creation.

**Create a car name (a write).** `whatsapp-tool car-names add` creates a model under an existing
brand, mirroring `useCarNames.js` `addCarName()` (`INSERT INTO cars_names (car_name, notes,
is_big_car, id_brand)`; the name is trimmed and upcased, like the form). The brand must exist first
(list ids with `whatsapp-tool brands`) and **`--big` or `--small` is required** — big/small is a
required column (`cars_names.is_big_car`, no default), so never guess it: if you do not know, ask
the user. Run it only when asked, and confirm the name and size first:

```
whatsapp-tool car-names add --name "GS3 MID" --brand-id 17 --small   # a small model under brand 17
whatsapp-tool car-names add --name "GS4" --brand-id 17 --big         # is_big_car = 1
```

It answers `{status:"OK", action:"car-name-add", id, carName, brandId, isBigCar}`. Gated on
`can_purchase_cars`. There is no delete/rename here — only creation.

## 15d. Inquiries — supplier quote requests (the cars app)

The **inquiry** system is the purchase-side twin of buy bills: a header (one inquiry to a supplier)
plus car lines, mirroring the app's `buy_bill` / `buy_details` (tables `inquiry` /
`inquiry_details`, migration `034`). There is **no inquiry screen in the app yet** — these commands
are how inquiries are created and read. Together with `brands add` / `car-names add` (§15c) and
`suppliers add` (§15b) it is one of the cars-app **write** paths; run the writes only when asked,
and confirm the supplier and car first.

```
whatsapp-tool inquiries                                       # list all (newest first)
whatsapp-tool inquiries list --supplier-id 3 --status open    # filter by supplier / status / ref / id
whatsapp-tool inquiries create --supplier-id 3 --ref "INQ-1" [--date ..] [--status open] [--notes ..]
whatsapp-tool inquiries add-car --inquiry-id 12 --car-name-id 4 --color-id 1 --qty 2 (--used|--new) --target-price 12500 [--year 2020] [--month 6] [--big] [--notes ..]
whatsapp-tool inquiries cars 12                               # the car lines of one inquiry
whatsapp-tool inquiries delete 12                             # header + its lines (FK cascade)
```

Required by the schema (migration 035), so the command refuses without them: `--color-id`, `--qty`,
`--target-price` and **`--used` or `--new`** on `add-car`; `date_inquiry` on `create` (the command
defaults it to **now**, so you normally omit `--date`). `--year`/`--month`/`--big`/`--notes` stay
optional. If the color or the used/new answer is unknown, ask the user — do not guess.

Reading the JSON: `create` answers `{status:"OK", action:"inquiry-create", id, supplierId, date,
ref, status}`; `list` answers `inquiries[]` rows (`id`, `id_supplier`, `supplier_name`,
`date_inquiry`, `inquiry_ref`, `amount`, `status`, `notes`); `cars` answers `cars[]` rows (`id`,
`id_inquiry`, `id_car_name`, `car_name`, `id_color`, `color`, `QTY`, `year`, `month`, `is_used_car`,
`is_big_car`, `target_price`, `notes`). `add-car` attaches one line per call (run it once per car);
`delete` removes the header and, by the foreign key's `ON DELETE CASCADE`, its lines in one step.
Never invent an inquiry, supplier id, car-name id or color id — use `suppliers` / `car-names`
(§15b/§15c) to resolve ids first.

**Permission:** the buy/models area is gated on `can_purchase_cars` (in `CarsView.vue`), so every
`inquiries` subcommand refuses to run without it (`status: ERROR, code: missing_permission`). Setup
/ credentials are shared with §15 (run `stock creds` once).

## 15e. Clearance — per-car customs clearance (the cars app)

Customs-clearance questions ("what's the clearance for <car>") can be read from the **same cars
app**, from the table `clearance` (migration `036`) — one row per model/variant, joined to its
model and brand. This is the cars-app per-model clearance, **not** the `ai`-database pricing
`clearance`/`--mode clearance` of §13 (that one still drives `whatsapp-tool quote`; keep the two
separate). Reads and the `set` write share the credentials with `stock`:

```
whatsapp-tool clearance                                     # all rows, ordered by car name
whatsapp-tool clearance --name "COOLRAY" --brand "GEELY"     # match by car name and/or brand
whatsapp-tool clearance --car-name-id 39 --limit 20          # one model / a slice
```

Reading the JSON: `count` + `clearance[]`, each row `id`, `id_car_name`, `car_name`, `id_brand`,
`brand`, `clearance_dza` (a **string** decimal, DZD, **TVN excluded**), `is_big_car` (0/1),
`is_used_car` (0/1), `notes`, `created_at`. Never invent a clearance amount — only report what the
command returns.

**Set a clearance (a write).** `whatsapp-tool clearance set` upserts one row
(`INSERT ... ON DUPLICATE KEY UPDATE` on the `(id_car_name, is_big_car, is_used_car)` unique key),
so setting an existing model/variant updates it in place. The car name must exist (list ids with
`whatsapp-tool car-names`) and big/small + used/new are **part of the key** — required explicit
choices, never guessed. `--dza` is the amount in DZD (TVN excluded):

```
whatsapp-tool clearance set --car-name-id 39 --dza 950000 --small --new
whatsapp-tool clearance set --car-name-id 4  --dza 800000 --big --used --notes "includes port fees"
```

It answers `{status:"OK", action:"clearance-set", carNameId, dza, isBigCar, isUsedCar, notes,
inserted, affectedRows}` (`inserted`/`affectedRows` follow MySQL's upsert: 1 insert, 2 a changed
row, 0 a no-op). Omitting `--notes` leaves any existing note untouched on update. Run it only when
asked, and confirm the model, variant and amount first — see "Strict rules".

**Permission:** none — clearance is free. Any logged-in AI user may read it and run `clearance set`
(there is no `can_c_cfr_prices` gate). Setup / credentials are shared with §15 (run `stock creds`
once).

## 15f. Buy payments — what we paid a supplier (the cars app)

Payment questions ("how much have we paid supplier X", "is this purchase bill paid", "which bills
are still owing") come from the **same cars app**, read-only, table `buy_payments` joined to
`buy_bill` / `suppliers` / `users`, mirroring the app's `BuyBillPaymentsView.vue`. This is what we
pay the **supplier**, not client receipts.

```
whatsapp-tool buy-payments --bill-id 1                     # one bill: header + its payments
whatsapp-tool buy-payments --supplier-id 3 [--limit 20]    # a supplier's payments across bills
whatsapp-tool buy-payments --id 7                          # a single payment row
whatsapp-tool buy-payments [--limit 20]                    # all payments, newest first
whatsapp-tool buy-payments add --bill-id 1 --amount 77000 --date "2026-10-11 12:00:00" --swift /path/swift.pdf
whatsapp-tool buy-payments edit --id 3 --amount 60000 --date "2026-10-11 13:30:00"   # keep swift/notes
whatsapp-tool buy-payments delete --id 3
```

Reading the JSON: `count` + `payments[]`, each row `id`, `id_buy_bill`, `bill_ref`, `id_supplier`,
`supplier_name`, `date_payment`, `amount`, `swift_path`, `notes`, `id_user`, `created_by`. With
`--bill-id` a `bill` header is also returned: `id`, `bill_ref`, `id_supplier`, `supplier_name`,
`date_buy`, `stored_amount`, `payed`, `calculated_amount` (Σ amount×QTY over `buy_details`, the
amount owed) and `total_paid` (Σ of that bill's payments). Never invent an amount or a supplier —
only report what the command returns.

**Record a payment (a write).** `whatsapp-tool buy-payments add` inserts a row into `buy_payments`
and then recomputes the bill's `payed` total (`buy_bill.payed` = Σ of that bill's payments),
mirroring `BuyBillPaymentsView.vue` `savePayment()` + `updateBillPaidAmount()`. The app's form makes
these **obligatory**, so the command refuses without every one of them — never guess or default:

```
whatsapp-tool buy-payments add --bill-id 1 --amount 77000 --date "2026-10-11 12:00:00" --swift /path/swift.pdf [--notes ..] [--user-id N]
```

- `--bill-id` — the purchase bill being paid (must exist).
- `--amount` — the amount paid.
- `--date` — `date_payment` (`YYYY-MM-DD HH:MM:SS`).
- `--swift` — the **swift document**, a local file path; it is uploaded to `buy_payment_swifts` like
  the app before the row is inserted.
- `--user-id` — optional; defaults to the AI user's id (cached at login). The recording user is
  stored in `id_user`.
- `--notes` — optional.

It answers `{status:"OK", action:"buy-payment-add", id, billId, amount, datePayment, swiftPath,
userId, billPayedResynced}`. Run it only when asked, and confirm the bill, amount, date and swift
file first — see "Strict rules".

**Change or remove a payment (writes).**

- `whatsapp-tool buy-payments edit --id N --amount V --date ".." [--swift <path>] [--notes ..]` —
  `--id`, `--amount` and `--date` are required (the app's form requires amount and date). `--swift`
  is optional: pass a local file path to upload a replacement to `buy_payment_swifts`, or omit it to
  keep the stored document. `--notes` omitted keeps the stored note; `--notes ""` clears it. Answers
  `{action:"buy-payment-edit", id, billId, amount, datePayment, swiftPath, notes, affectedRows,
  billPayedResynced}`.
- `whatsapp-tool buy-payments delete --id 3` — removes the payment (destructive; run it only when the
  user asks). Answers `{action:"buy-payment-delete", id, billId, deleted, billPayedResynced}`.

Both mirror the app's edit/delete and end with the same `buy_bill.payed` resync, resolving the bill
from the payment itself.

**Permission:** payments to suppliers are buy-cost figures, so the read and every write are gated on
`can_pay_purchases` (a dedicated permission, migration `037`) — they refuse to run without it
(`status: ERROR, code: missing_permission`) — grant it in the app's RolePermissionsCrud, then re-run
`whatsapp-tool stock creds --username <user>`. Setup / credentials are shared with §15 (run `stock
creds` once).

> **Two different payment systems — don't mix them.** This section is the **cars app**'s
> `buy_payments` (a payment against a **buy bill**, resyncing `buy_bill.payed`). The **ai** database
> keeps its own supplier payment ledger (`supplier_payments`, linked to the **quotation**/contract we
> recorded — see §16). "How much do we still owe <supplier>" is the ai ledger's job:
> `whatsapp-tool procurement outstanding --supplier-id N` (§16) — sum the deals, **subtract the
> recorded payments**, show `owed − paid`. Never answer it from documents/buy-bills alone.

## 16. Procurement — supplier quotations (the ai database)

Supplier quotations you are asked to save live in the **shared `ai` database** (db/schema.sql),
**not** the cars app — unrelated to the cars-app suppliers/inquiries (§15b/§15d), but **linked to
the pricing catalog**: `cars.supplier_id` says who supplies a car's current price, the price
history carries `supplier_id`/`quotation_id`, and quotation lines carry an optional `car_ref`
(SET NULL) to the catalog car they match. That chain answers "who gives us this price?" and "which
quotation did it come from?". Read/write them with:

```
whatsapp-tool procurement suppliers [--name X] [--limit N]                       # list
whatsapp-tool procurement suppliers add --name "abi"                              # new
whatsapp-tool procurement suppliers update --id 2 --contact "Ruby, +86..." [--notes ..]   # set seller contact/notes
whatsapp-tool procurement quotations [--supplier-id N] [--status S] [--ref X] [--limit N]
whatsapp-tool procurement quotations show <id>
whatsapp-tool procurement quotations create --supplier-id N [--ref ..] [--date ..] [--currency USD] [--status ..] [--valid-until ..]   # --date defaults to now
whatsapp-tool procurement quotations add-item --quotation-id N --car "<name>" (--big|--small) (--used|--new) [--qty N] [--year ..] [--month ..] [--notes ..] [--car-ref <carId>]
whatsapp-tool procurement quotations set-price --item-id N --color "<color>" --price N [--notes ..]
whatsapp-tool procurement quotations set-status <id> --status requested|received|accepted|rejected|expired
whatsapp-tool procurement quotations delete <id>
```

Structure: `suppliers` → `supplier_quotations` (the **timestamp** is `quoted_at`) →
`supplier_quotation_items` (one car line) → `supplier_quotation_item_prices` (**one price per
colour** — that is where "price differs by colour" lives). Reading the JSON: `quotations[]` rows
are `id`, `supplierId`, `supplierName`, `quotedAt`, `reference`, `currency`, `status`,
`validUntil`, `notes`; `show` adds `items[]` each with `prices[]` (`color`, `unitPrice`, `notes`).
Never invent a supplier, quotation, item or price — only record what the user gives you.

**RULE — every supplier price the user gives you IS a supplier quotation.** When the user hands
you a supplier price (e.g. "car X at 12900 from lio"), record the full chain — never just the
cost:
1. create the supplier if it isn't there (`suppliers add --name ...`),
2. create the quotation (`quotations create --supplier-id N`; date defaults to now),
3. `add-item --car-ref <carId>` to link the line to the `cars` row,
4. `set-price` once per colour (as many colours as the user named) — this is where
   "price differs by colour" lives,
5. then record the buy price with **`price set-car --supplier "<name>" --quotation <id>`**, which
   writes `cars.supplier_id` and stamps the appended `car_price_history` row with
   `supplier_id` + `quotation_id`.

That closes the chain supplier → quotation → item → car → price, so the operator can always answer
"who gives us this car's price and which quotation did it come from": `list-cars`/`quote` output
carries `supplierId`; `quotations show` shows the prices per colour — the current `suppliers` are
walle (1), jean fat boy (2), lio (7), kent (11) on the existing catalog; list them with `procurement suppliers`.
kent (11, quotation 27, 2026-10-16) currently supplies: LIVAN MAN 6900, LIVAN AUTO 7700, GS3 MID 9050,
GS3 FULL 10500, GS3 MID (R style) 11200 (special grey 11300), CHANGAN X5 BEYOND PRO 9750 and the new
CHANGAN X5PLUS EXCELLENT PRO 10700 — all FOB, colours grey/white or "any". Same quotation also carries
the kent **modified-spec** Coolrays (new catalog lines, kept apart from the international-spec ones):
COOLRAY MAN (MODIFIED) 7700, COOLRAY L (MODIFIED) 10700, COOLRAY BATTLE (MODIFIED) 10900 / white 11000.

**Documents — the paper trail.** Real supplier documents (proforma invoice, sales contract, …) are
saved in the `documents` table and linked to the quotation(s) they back via `document_quotations`
(M:N), closing document → quotation → item → per-colour price → cars. `source_text` holds the full
transcription so the original wording stays AI-readable; one document row per document number
(the PI *and* its sales contract are separate rows, each `link`ed to the same quotation):

```
whatsapp-tool procurement documents [--supplier-id N] [--type T] [--limit N]   # list
whatsapp-tool procurement documents show <id>        # + source_text + linked quotations
whatsapp-tool procurement documents create --no "PIXSLNR20261011L001" --type proforma_invoice --supplier-id 7 [--date 2026-10-11] [--counterparty "Mr. Nour"] [--currency USD] [--total 67700] [--words ".."] [--text ..|--text-file <path>] [--status received] [--notes ..]
whatsapp-tool procurement documents link --document-id N --quotation-id M     # backs a quotation
whatsapp-tool procurement documents unlink --document-id N --quotation-id M
```

**Supplier payments — the money trail.** What we actually paid a supplier (migration 005) lives in
`supplier_payments`, each linked through the **M:N junction `payment_quotations`** to the
quotation(s)/contract(s) it settles (one payment can cover several invoices — migration 006), and
to its SWIFT MT103 proof in the 1:1 child `payment_swifts` (CASCADE). The swift row stores the
**PDF link you give** (`--file-url`) and the **full manuscript** (`--text`/`--text-file`, AI-readable),
plus extracted `swift-ref` (field 20), `uetr` (field 121), sender/receiver BIC and beneficiary;
re-running `set-swift` overwrites the same row. Manual (not bank-feed) curated data, never guessed:

```
whatsapp-tool procurement payments [--supplier-id N] [--quotation-id N] [--status S] [--limit N]  # list
whatsapp-tool procurement payments show <id>                 # + swift proof + linked quotations
whatsapp-tool procurement payments create --supplier-id 2 --amount 20800 [--currency USD] [--date 2026-10-05] [--value-date ..] [--method swift] [--status sent] [--notes ..]
whatsapp-tool procurement payments edit --id N [--payment-no ..] [--supplier-id N] [--amount V] [--currency ..] [--date ..] [--value-date ..] [--method ..] [--status ..] [--notes ..]
whatsapp-tool procurement payments link --payment-id N --quotation-id M     # payment settles a quotation
whatsapp-tool procurement payments unlink --payment-id N --quotation-id M
whatsapp-tool procurement payments delete --id N             # destructive; removes its swift + links too
whatsapp-tool procurement payments set-swift --payment-id N [--swift-ref ..] [--uetr ..] [--sender-bic ..] [--receiver-bic ..] [--beneficiary-name ..] [--beneficiary-account ..] [--file-url <pdf link>] [--file <local path>] [--text ..|--text-file <path>]
```

Payment `status` is one of `planned|sent|cleared|returned|cancelled` (default `sent`); `method` one
of `swift|fps|chats|cash|other` (default `swift`). `--date` defaults to today. Quotation links are
**not** part of `create`/`edit` — add them with `link` (idempotent) and remove with `unlink`; `link`
refuses if the payment or quotation does not exist. Never invent a payment, amount, or swift
detail — only record what the user gives.

**`payment_no` = a sequential integer — ALWAYS.** Never key a payment on a bank/document reference.
**Omit `--payment-no`** and the tool mints the next integer (one more than the highest numeric
`payment_no`) — `1`, `2`, `3`, … — that is the ledger's unique id. The bank's own reference, **if**
the form shows one, belongs in the swift proof's `--swift-ref`, never in `payment_no`. (Current:
`1` walle, `2` jean fat boy; the next is `3`.)

**How much do we still owe a supplier (outstanding) — NEVER answer this from the documents alone.**
Money question; the payment ledger must be subtracted. Use the calculator, not hand arithmetic:

```
whatsapp-tool procurement outstanding [--supplier-id N]    # per supplier: owed / paid / outstanding
```

It returns, per supplier, `owed` (Σ of that supplier's deals — one amount per **quotation**, so a
deal's PI **and** its sales contract count **once**; a deal with no non-cancelled document carrying
a total shows `amount: null` and is **not** summed), `paid` (Σ of `supplier_payments` rows with
`status` `sent`/`cleared`), and `outstanding = owed − paid`, plus the `deals[]` and `payments[]`
lists for audit. **Answering "how much do I need to pay <supplier>" without subtracting the recorded
payments is wrong** — quote the command's `outstanding`, show the arithmetic (`owed − paid`), and
name the deals and payments behind it. A payment only counts once it is recorded here: every T/T we
send must be saved (`payments create` + `set-swift` + `link`), otherwise the balance stays too high.
Quotations/documents with no captured total are excluded — flag them to the user, never guess.

**STRICT RULE: before any procurement write** (`suppliers add` → `--name`; `quotations create` →
`--supplier-id`; `add-item` → `--quotation-id` + `--car` + `--big`/`--small` + `--used`/`--new`;
`set-price` → `--item-id` + `--color` + `--price`; `documents create` → `--no` + `--type` +
`--supplier-id`; `payments create` → `--supplier-id` + `--amount` (the `payment_no` is auto-minted);
`payments
set-swift` → `--payment-id`, plus whatever swift fields/manuscript the user hands you) **ask the
user for every obligatory field you don't already have, and wait.** Only a
`set-status`/`delete` on an existing quotation is fine to run when asked — confirm a `delete`
explicitly (it removes the items and prices too, cascade), same for `payments delete` (cascades to
its swift row).

## 17. Tasks & reminders — `whatsapp-tool todo` (the ai database)

A **todo** is a thing a client or workmate asked **the logged-in user** (this WhatsApp account) to
do. Todos live in the shared **`ai`** DB (tables `todo_users` / `todos` / `todo_reminders` /
`todo_status_history`, migration `008`), belonging to the login (`WA_PROFILE`) — the owner is
resolved from the session, never asked. There is **no todo screen**; this CLI is it.

```
whatsapp-tool todo add --title "..." [--parent N] [--desc ..] [--priority low|normal|high|urgent] [--due 2026-10-20] [--requested-by ..] [--chat ..] [--owner <profile>]
whatsapp-tool todo list [--owner <profile> | --all] [--status open|done|cancelled] [--priority ..] [--overdue] [--due-before ..] [--roots] [--parent N] [--limit N] [--tree]
whatsapp-tool todo show <id>
whatsapp-tool todo done <id> | reopen <id> | cancel <id>
whatsapp-tool todo edit <id> [--title ..] [--desc ..] [--priority ..] [--due .. | --clear-due] [--requested-by ..] [--chat ..]
whatsapp-tool todo move <id> --parent <id|none>            # re-parent (rejected if it would cycle)
whatsapp-tool todo assign <id> --owner <profile>           # hand a todo to another login
whatsapp-tool todo delete <id>                             # removes the todo AND its whole subtree
whatsapp-tool todo remind add <id> --at "2026-10-20 09:30" [--channel whatsapp|local] [--note ..]
whatsapp-tool todo remind list [--todo N] [--status pending|sent|cancelled] [--limit N]
whatsapp-tool todo remind rm <id> | sent <id>
whatsapp-tool todo due [--before <date>]
whatsapp-tool todo users
```

- **Priority** `low|normal|high|urgent` (default `normal`); **status** `open|done|cancelled`.
- **Deadline** `--due` accepts a date (`2026-10-20`, treated as end-of-day) or a datetime
  (`2026-10-20 09:30`). `--overdue` = open todos past their `due_at`; `--due-before` cuts at a date.
- **Subtasks** nest without limit (`--parent N`). A parent with children auto-becomes `done` when
  **all** its children are `done` or `cancelled`, and auto-**reopens** when a new open child is
  added — cancelled counts as resolved. `--tree` nests the output by parent.
- **Reminders** have no daemon: a due reminder (status `pending`, `remind_at` reached, todo still
  open) shows up in `whatsapp-tool todo due` and as the **`todosDue`** block of the `pending` report
  (§14b). After acting on one, mark it `whatsapp-tool todo remind sent <id>`.
- **Permission:** every write (`add/edit/done/reopen/cancel/move/assign/delete/remind add|rm|sent`)
  needs `can_give_edit_instruction`; the reads (`list/show/due/users/remind list`) are open.
- Never invent a todo, an owner, or a reminder time — only record what the user gives you. Confirm
  a `delete` first (it cascades the subtree).

## Strict rules

**STRICT RULE: before any write/INSERT (cars-app `brands add`, `car-names add`, `suppliers add`, `inquiries create`/`add-car`, `clearance set`, `buy-payments add`/`edit`/`delete`, or any other create/update), you MUST ask the user for every OBLIGATORY field whose value you do not already have from the user, and wait for the answer. Never guess, invent, or let a required field take a hidden default. Required today: `car-names add` → name + `--brand-id` + `--big`/`--small`; `inquiries create` → supplier (date defaults to now); `inquiries add-car` → inquiry + car + `--color-id` + `--qty` + `--target-price` + `--used`/`--new`; `clearance set` → `--car-name-id` + `--dza` + `--big`/`--small` + `--used`/`--new`; `buy-payments add` → `--bill-id` + `--amount` + `--date` + `--swift`; `buy-payments edit` → `--id` + `--amount` + `--date`; `buy-payments delete` → `--id` (and confirm the deletion explicitly); `todo add` → `--title` (parent/priority/due optional; `todo remind add` → todo id + `--at`; `todo delete` cascades the subtree — confirm it first). Only after the user confirms every obligatory value do you post. Ask first, post second — always.**

**Strict rule: never delete any profile (from profiles.json, data/* for that profile, CDP state) without explicit confirmation from the user.**

**Strict rule: never register/link accounts programmatically. Do not run any account-linking/`connect` command. The user connects WhatsApp manually.**

**STRICTEST RULE: DO NOT delete, remove, drop, or modify profiles.json entries, profile folders (data/<name>/), CDP files, or any account-related state WITHOUT EXPLICIT USER CONFIRMATION. Do not run any account-linking/connect command.**
