# DEVELOPMENT.md — Building & Changing the WhatsApp MCP Server

Developer guide. This file is read only when you are **working on the code**. For day-to-day
operation (driving the tools, replying to clients, pricing, tracking) read [`AGENTS.md`](./AGENTS.md),
which is the slim operational manual.

> **⚠️ The project is now Go.** The Node/TypeScript source (everything under `src/`, `dist/`,
> `link_manual.sh`, `package.json`) has been **removed** — it lives only in
> `backups/node-code-backup-*.tar.gz`. The live code is the Go module under `go/` (see
> [`go/MIGRATION_GO.md`](./go/MIGRATION_GO.md) for the port map and stage notes). **Build/run with
> the Go commands; the `src/…` tree and `npm run …` snippets further down are the historical Node
> design kept for reference.** Before writing any Go code, read the module rules —
> [`go/RULES.md`](./go/RULES.md): maximum modularity, no file > 1000 lines, AI-friendly naming
> and contracts, max scalability. Command mapping:
>
> | Old (Node) | New (Go) |
> |---|---|
> | `npm run build` | `cd go && go build -o bin/whatsapp-tool ./cmd/whatsapp-tool` |
> | `npm start` / `npm run dev` | `whatsapp-tool mcp` |
> | `npm run link` | `WA_HEADLESS=0 whatsapp-tool keepalive start` (manual QR) |
> | `npm run smoke` | `WA_SMOKE=1 go test -run TestSmokeLive ./internal/wa -v` (in `go/`) |
> | `npm run pending`/`triage`/`quote`/`price`/`faq`/`stock`/`identity` | `whatsapp-tool <name> …` |
> | `npm run keepalive[:status\|:stop]` | `whatsapp-tool keepalive start\|status\|stop` |
> | `npm run profiles[:mark-reply\|:set-active]` | `whatsapp-tool profiles [mark-reply\|set-active] …` |
> | `npm run db:schema`/`db:seed`, `npm run reply-all`, `npm run connect` | removed (no Go equivalent) |

Project: `/Users/merhab/dev/browser_automation/whatsapp-agent-tool` — an MCP server that drives
WhatsApp Web through **real Google Chrome** + go-rod.

---

## 1. Mental model

```
MCP client (opencode/Claude/Cline)
   │  stdio JSON-RPC
   ▼
go/cmd/whatsapp-tool/main.go   dispatch + stdio MCP server, all logs → stderr
   │
   ▼
go/internal/wa/browser.go      WhatsApp client — one chat "open" at a time (STATEFUL)
   │
   ▼
go/internal/chrome/chrome.go   persistent Chrome: spawn detached + CDP attach
   ▼
Google Chrome (real) — OWN user-data-dir:
   profile: data/whatsapp-profile/   ← QR login + IndexedDB auth live here
   CDP port: 9222-9252, first free one, recorded in data/whatsapp-cdp.json
```

The client is **stateful**: exactly one chat is "open" at a time, shared across all tool calls.
`open_chat` → then `send_message` / `get_last_messages` / `mark_as_read`.

**Chrome outlives the process.** `go/internal/chrome/chrome.go` spawns Chrome itself with
`--remote-debugging-port`, detached (new session, released), and the MCP server attaches over CDP.
Detaching the CDP client leaves the browser running, so the next task attaches in ~0.5 s instead of
booting for 10-25 s. See §3.1.

---

## 2. Setup (once per machine)

Requirements: **Google Chrome must be installed** (mandatory — see §7), **Go 1.27+**.

```bash
cd /Users/merhab/dev/browser_automation/whatsapp-agent-tool/go
go build -o bin/whatsapp-tool ./cmd/whatsapp-tool
```

### Link the account (only needed the first time, or after `data/` is deleted)

```bash
WA_HEADLESS=0 /Users/merhab/dev/browser_automation/whatsapp-agent-tool/go/bin/whatsapp-tool keepalive start
```

A **visible Chrome window** opens on the QR screen. On the phone:
**WhatsApp → Linked devices → Link a device** → scan the QR.

- If a 2-step-verification PIN is requested, **type the PIN in that Chrome window yourself**.
- Press `Ctrl+C` in that Terminal (Chrome keeps running; stop it with `whatsapp-tool keepalive stop`).
  The session persists.
- Force a new QR: stop everything, then `rm -rf data/whatsapp-profile` and link again.

### Verify

```bash
cd go
WA_SMOKE=1 go test -run TestSmokeLive ./internal/wa -v   # attaches, prints login state + sample
```

---

## 3. Running the server

```bash
go/bin/whatsapp-tool mcp             # stdio MCP server (headless)
WA_HEADLESS=0 go/bin/whatsapp-tool mcp   # visible Chrome window
```

`WA_HEADLESS` is checked as `!== "0"`, so anything but `0` means headless.

### opencode MCP config

`opencode.json` (project or global `~/.config/opencode/opencode.json`):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "whatsapp": {
      "type": "local",
      "command": ["/Users/merhab/dev/browser_automation/whatsapp-agent-tool/go/bin/whatsapp-tool", "mcp"],
      "enabled": true
    }
  }
}
```

Check it: `opencode mcp list`. Disable without deleting: `"enabled": false`, or
`"tools": { "whatsapp_*": false }`.

Claude Desktop / Cline (mcpServers form):

```json
{ "mcpServers": { "whatsapp": { "command": "/Users/merhab/dev/browser_automation/whatsapp-agent-tool/go/bin/whatsapp-tool", "args": ["mcp"] } } }
```

> The first tool call that has to start Chrome pays the boot (~10-25 s); every later call is
> ~0.5-1 s because it attaches to the Chrome that is already running.

### 3.1 Keep the browser between tasks (default behaviour)

Chrome is **not** a child of the MCP server. `go/internal/chrome/chrome.go` starts it detached on the
first free CDP port in 9222-9252 (recorded in `data/whatsapp-cdp.json`) and every later process
attaches to that same browser. `whatsapp_session_info` reports which transport is live. With
`WA_PROFILE=<name>` set, everything moves to `data/<name>/` + `data/<name>-cdp.json` instead —
see §10.

| Command | What it does |
|---|---|
| `whatsapp-tool keepalive start` | Start Chrome now (if needed), verify login, then idle. Chrome stays up even if you Ctrl+C. |
| `whatsapp-tool keepalive status` | Is Chrome up on our port? Is WhatsApp logged in? Sample chats. |
| `whatsapp-tool keepalive stop` | Kill the shared Chrome. |

Optional `whatsapp-tool keepalive start` in a Terminal at the start of the day, or just let the first
tool call spawn it — either way the browser is reused. Linking needs the profile to itself, so stop
Chrome first (`whatsapp-tool keepalive stop`).

Env: `WA_CDP_PORT` (pin the port), `WA_HEADLESS=0` (visible window), `WA_TRANSPORT=launch`
(legacy in-process Chrome, closes on exit — slower, use only to debug), `WA_CHROME_PATH`.

**Never attach to a Chrome we did not start.** If another program owns a port in our range we skip
it; if `WA_CDP_PORT` is pinned to a port owned by someone else, `init()` fails with an error rather
than driving a foreign browser.

---

## 4. Tool reference

Tools registered in `go/internal/mcp/mcp.go`. Outputs are compact JSON (pretty-print removed on large
payloads to save agent tokens). `whatsapp_session_info` stays pretty (it is tiny).

### `whatsapp_is_logged_in`
No args → `{"loggedIn": true|false}`.
Errors mention `run ./link_manual.sh` when not linked. Call this first if anything else misbehaves.

### `whatsapp_get_chats`
`limit?: 1..1000` (default **15** screen / **50** newest-unread), `mode?: "screen" | "newest" | "unread"`
(default `screen`), `timeoutMs?: 5000..600000` (newest/unread only, default 120000)
→ `{ mode, chats, collected?, unreadTotal?, … }`

One tool for all three listing strategies (merged in round 2 of the token/speed pass):

- `mode: "screen"` — fast, first screen only, `limit` clamped to 50. The sidebar virtualizes: only
  ~15 rows exist in the DOM no matter how big `limit` is, so capping it is honest, not lazy.
- `mode: "newest"` — walks the sidebar the way a person does and collects the last N chats: wheels
  back to the newest chat first (results are the *last* N, not wherever the list was left),
  dedupes by name (gotcha 16), stops at `limit` / list end / 4 dry rounds / timeout.
- `mode: "unread"` — clicks the Unread filter chip (the `button[role="tab"]` **inside** the chip's
  `[data-testid="filter-button"]` wrapper — see gotcha 22), collects with the same human-paced
  scroll, **restores All in a `finally`** (even on throw), and returns `unreadTotal` = WhatsApp's
  own counter off the filter chip (Arabic-Indic digits parsed) — so "351 unread, newest 30" is
  distinguishable from "that's all of them".

Common fields: `name` is a display name **or** phone number (use it as `open_chat` input),
`preview` = last message snippet, `time`, `unread`/`unreadCount`. Chat-list banners ("1 unread
message", `رسالة واحدة غير مقروءة`) are stripped from `name`. Archived chats are **not** included
(they sit behind `chatlist-panel-archived-button`), which is why `unreadTotal` can exceed
`collected` at `reachedEnd: true`.

Scrolling is human-paced on purpose: each gesture moves the mouse over the list first, wheels
55-110% of viewport height with occasional longer flicks, pauses 380-950 ms (longer "reading"
pause every ~10%) and adds a small upward correction ~25% of the time. Don't "optimise" this into
`scrollTop = 999999` on a timer — that is the loudest automation signature there is, and reading
the DOM without scrolling never worked on a 1000-chat account.

Timing on this account: 15 chats ≈ 3 s, 200 unread ≈ 13 s, the whole list (881 chats) ≈ 60-70 s,
all 351 unread ≈ 35 s. A full walk drifts by 1-2 chats against WhatsApp's own counter because
messages keep arriving and reordering the list — expected.

### `whatsapp_open_chat`
`nameOrPhone: string` (required) → `{"success": true}`.

Types into the chat-list search box, waits for results, clicks the row whose title contains the
query, and waits until the conversation header + compose box exist. Returns `success: false` when
nothing could be **verified** — treat that as "no such chat", do not retry blindly. It only ever
clicks rows the query provably produced (see gotcha 21): it types with real keystrokes
(`fill()` does not trigger WhatsApp's search), requires the row title to contain the query
(fallback to the first filtered row only when the list demonstrably shrank), and re-checks the
conversation header after clicking. Fast-fail: when the search filters and provably finds
nothing, it returns `false` on the first attempt instead of retrying (~7 s saved per miss).

### `whatsapp_send_message`
`text: string` (required, non-empty) → `{"success": true}`.

Clicks the compose box, types with randomized per-key delay, presses `Enter`. Sends to the
**currently open chat only**. `success: true` means "Enter was pressed", not delivery confirmation.

### `whatsapp_get_last_messages`
`limit?: 1..50` (default **10**) → `{ messages: [{ from: "me"|"them", text, time? }] }`

- `from` is derived from **bubble geometry + panel direction** (`isOutgoing()` in `go/internal/dom`),
  not from the tail test id. See gotcha 8 — the tail markers are mirrored in the Arabic UI and were
  the cause of every message being attributed to the wrong person.
- Text-only: attachments/system notices without `selectable-text` are skipped.
- Returns only what is currently rendered in the conversation panel (WhatsApp virtualizes).

### `whatsapp_mark_as_read`
No args → `{"success": true|false}`. Opens the chat header menu then closes it (best effort).

### `whatsapp_session_info`
No args → `{"transport": "cdp"|"launch"|"none", "attachedOverCdp", "initialized", "cdpEndpoint",
"cdpPort", "chromePid", "chromeRunning", "clientUptimeSec", "pageUrl", "note", "identity"}`.

Does **not** boot Chrome — it only reports state, so it is the cheap first call when something
feels slow. `transport: "cdp"` = attached to the long-lived browser (the good case);
`"none"` = nothing attached yet, the next tool call will attach or spawn. `identity` is the
assistant's code-generated name (see `whatsapp_identity`).

### `whatsapp_identity`
`force?: boolean` (default `false`) → `{ name, token, profile, createdAt }`.

Returns the assistant's unique, irreplicable name, minted **by code** (`crypto.randomBytes`,
never by the model) on the first start and persisted to `data/ai-identity.json` (or
`data/<profile>-ai-identity.json`). Stable across sessions; `force: true` re-mints. Implemented
in `go/internal/identity`; also available as `whatsapp-tool identity [--reset]`.

---

## 5. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `WhatsApp Web refused the bundled Chromium` | ran without real Chrome | install/update Google Chrome |
| `Not linked yet. Run ./link_manual.sh…` | QR screen, and server is headless | run `WA_HEADLESS=0 whatsapp-tool keepalive start`, scan, stop |
| `Linking failed or timed out` + `page text: …` in logs | scan rejected/expired, or 2FA pending | re-link; type PIN in the visible window if prompted |
| `whatsapp_get_chats` returns `[]` | page still hydrating, or logged out | retry after a few seconds; `whatsapp_is_logged_in` |
| `open_chat` → `success: false` | no match / list not loaded | `get_chats` to find the exact name, then retry |
| `send_message` → `success: false` | no chat open | call `open_chat` first |
| Chrome won't start (profile in use) | another Chrome on the profile | close it; `whatsapp-tool keepalive stop` |
| Empty texts / weird names | WhatsApp UI redesign | re-inspect the DOM, update selectors (§7) |

Debug helper: `WA_SMOKE=1 go test -run TestSmokeLive ./internal/wa -v` prints login state, chats and
messages. For DOM archaeology, dump the `data-testid` set from inside a `Page.Eval` in the driver.

---

## 6. Hard-won gotchas (do not regress these)

1. **A bundled Chromium is rejected.** A bundled/headless Chromium loads a WhatsApp page that only
   says "WhatsApp works with Google Chrome 100+". Always launch **real Google Chrome**.
2. **Never use `storageState` for auth.** WhatsApp Web keeps auth in IndexedDB; a saved
   `state.json` came out with 1 cookie and 0 origins and the session was silently lost. The
   persistent Chrome profile is the only reliable storage.
3. **`ignoreDefaultArgs: ["--enable-automation"]`** keeps the pairing QR flow from being aborted.
4. **Profile lock.** Two Chrome instances on `data/whatsapp-profile/` break each other. The lock
   is **per profile** (`data/<profile>/`), which is exactly what lets §10 run two accounts in
   parallel. Within one profile, one Chrome at a time, always.
5. **Hydration is slow.** Right after `goto`, the body says "downloading messages, don't close this
   window". Never decide "not logged in" from a 2 s wait — `waitForAppReady()` polls up to 120 s and
   `loginState()` returns `loading` meanwhile.
6. **`loginState()` states:** `loading | qr | pin | blocked | logged_in`. `blocked` = browser refused.
   Headless + (`qr`|`pin`) = surface the "not linked" error instead of waiting (link with a visible window).
7. **Never define named helpers inside `page.evaluate`.** esbuild's `keepNames` injects
   `__name is not defined` in the page. Do DOM extraction in `evaluate`, then clean/parse the raw
   strings in Go (`cleanText`, `lines`, `TIME_RE`, `chatNameFrom` in `go/internal/dom`).
8. **RTL/Arabic UI mirrors the tail markers.** In this account's Arabic UI, *your own* messages are
   the ones carrying `[data-testid="tail-in"]` and sitting on the **left**; `tail-out` is the other
   person. Mapping `tail-out → "me"` attributed every message to the wrong speaker. `getLastMessages`
   now uses bubble geometry + `direction` (`isOutgoing()`); verify any change to it against the
   **self-chat** (`[data-testid="message-yourself-row"]`), where every message is provably yours.
   Also strip bidi marks `[\u200E\u200F\u202A-\u202E\u2066-\u2069]` and never assume English date words
   (`TIME_RE` includes `أمس`, `اليوم`, `أبارحة`, `الآن`, …).
9. **All TS must live in `src/`.** `tsconfig.json` has `rootDir: "src"`; a root-level `.ts` file
   breaks `npm run build` with TS6059.
10. **Verification = `go build ./...` + `go vet ./...` + `go test ./...` + the gated live smoke
    (`WA_SMOKE=1 go test -run TestSmokeLive ./internal/wa -v`).** *(Historical: the Node era used
    `npx tsc --noEmit` + `npm run build` + `npm run smoke`.)*
11. **The self-chat's display name collides with a contact.** The `message-yourself-row` entry in
    the sidebar is titled like a normal contact, so `open_chat` with that name can open the
    *contact's* chat, where `from` is correctly `them`. To verify attribution, click
    `[data-testid="message-yourself-row"]` directly: every message there must come back `me`.
12. **Named helpers inside `page.evaluate` are the same bug as gotcha 7.** Keep rewriting them as
    inline expressions or `for` loops. `openChat` had two of them (`const title = (row) => …`),
    which is why `npm run smoke` crashed with `__name is not defined` the first time it ran after
    that code was written.
13. **The sidebar is virtualised even when every row is "in the DOM".** With ~900 chats the list
    container reports `scrollHeight ≈ 67457px` while only 15 rows carry `list-item-*` testids: the
    height is a spacer, not content. Any "read the whole list" approach has to scroll.
14. **`getLastChats` must actually reach the top of the sidebar first.** The list is ~67,000px
    tall, so a *bounded* number of upward gestures is not enough when a previous sweep left it at
    the bottom: it silently returned an arbitrary middle slice (250 of 880 chats) and reported
    `reachedEnd: true`. Loop until `top <= 2`, and bail only if the wheel stops moving twice.
    `waitForListSettled()` must also run after clicking a filter chip, because the list is briefly
    short while it re-renders — measuring then reports "already at top".
15. **A sweep's only real "end of list" signal is `moved ≈ 0` at the bottom, held 3×.** Never end on
    "no new rows appeared": the virtualiser recycles rows we already have, so that rule truncated
    real sweeps at 240/795 chats. And don't end on "no new rows *and* dry rounds" either — same bug.
16. **Dedupe sweeps on the chat *name*, not name+preview+time.** A new message landing mid-walk
    rewrites the preview and re-counted the same chat, which is why `collected` drifted upward
    (798 → 810 → 881 on one account in minutes).
17. **A long sweep can kill the tab.** After a 60-80s walk the renderer sometimes goes away and
    every later call in that process fails with "Target page, context or browser has been closed".
    `ensureInit()` therefore checks `contextClosed()` and re-inits instead of trusting
    `initialized`. If you add long loops, keep that heal path intact.
18. **`filter: "all"` must click the All chip, not assume it.** The unread sweep restores All in
    its `finally`, but if that process dies mid-walk the sidebar stays filtered and an "all" sweep
    quietly walks the unread subset (it reported 252 of ~795 chats).
19. **`open_chat` clicks by coordinates, on purpose.** The filtered sidebar keeps re-rendering, so
    a locator can hold a detached row and click nothing (`success: false` after ~15 s). The row is
    located and measured in one `page.evaluate`, then clicked with `page.mouse.click`. Never go back
    to per-row locator reads — each one costs the driver's timeout.
20. **Never leave text in the chat-list search box.** A leftover query filters the sidebar, so every
    later `get_chats` sees only the matching rows. `openChat` and `getChats` both call
    `clearChatSearch()`.
21. **`fill()` does not drive WhatsApp's search — and clicking "the first row" sends messages to
    the wrong clients.** The search box happily accepts `.fill()` text without firing the key
    events WhatsApp listens for, so the sidebar **never filtered** while the box looked filled.
    The old `openChat` then clicked `rows[0]` of the *unfiltered* list and returned `true`:
    on 2026-10-06 five replies were verified "sent" and only **one** reached its intended chat —
    the others landed in `+213 552 64 02 15`, the **MERHAB TEAM** group, `+213 662 29 67 98` and
    `+213 550 58 78 73`. `openChat` now types with `keyboard.type`, proves the box holds the
    query, clicks only a row whose title contains the query (or `rows[0]` **only** when the list
    demonstrably *shrank* = real filtering), and verifies the conversation header afterwards —
    otherwise `false`. Any send-side code must **still** compare `getCurrentChatName()` against
    the intended chat before typing (see the `send` mode in `src/autoreply.ts`).
22. **An Unread filter hides the `ai questions` group (and many chats) from the sidebar.** After a
    `get_chats` with `mode: "unread"` walk — or if any earlier tool call left the filter chip on "Unread" —
    searching for the group with `open_chat`/`openChat` can fail because the group is not in the
    filtered list. Before writing to `ai questions`, **restore the All filter first** — the
    `restoreAllFilter()` method on `WhatsAppClient` does exactly that (click the "All"/الكل chip and
    `waitForListSettled()`). The unread walker restores All in its `finally`, but a process that
    died mid-walk leaves the unread filter active.
23. **The sidebar filter chip is a wrapper; the click target is the inner button.** WhatsApp wraps
    each chip in `<div data-testid="filter-button" role="none">…<button role="tab">label</button></div>`.
    Clicking the `filter-button` div is a **no-op** — the list never switches (`filterLooksActive`
    then fails and the walker bails with the misleading `data-testid="filter-button" not found`
    note). Always click the chip's descendant `button[role="tab"]` (see `clickFilterButton`).

---

## 7. Selector map (WhatsApp Web, as of now)

| Purpose | Selector |
|---|---|
| Chat list container | `[data-testid="chat-list"]` |
| Its scrollable wrapper | ancestor with `overflow-y: auto` (≈338px tall in our window) |
| Chat row | `[data-testid="chat-list"] [data-testid^="list-item-"]` |
| Chat name | `[data-testid="cell-frame-title"]` |
| Chat preview/time | `[data-testid="cell-frame-secondary"]` |
| Unread badge | `[data-testid="icon-unread-count"]` |
| Sidebar filter chips | `[data-testid="filter-button"]` wrapper — All / Unread / Favorites / Groups; click its inner `button[role="tab"]` (gotcha 23) |
| Archived chats | `[data-testid="chatlist-panel-archived-button"]` (not walked by the sweeps) |
| Search box | `[data-testid="chat-list-search-container"] input` / `[contenteditable="true"]` |
| Conversation header title | `[data-testid="conversation-info-header-chat-title"]` |
| Message list | `[data-testid="conversation-panel-messages"]` |
| Message bubble | `[data-testid="msg-container"]` |
| Message text | `[data-testid="selectable-text"]` |
| Message time | `[data-testid="msg-meta"]` |
| Outgoing (sent by me) | bubble on the **left** when panel `direction: rtl` (see gotcha 8) |
| Compose box | `[data-testid="conversation-compose-box-input"]` |

Message actions menu (only if you must): hover bubble → `[aria-label*="خيارات الرسالة"]` (Arabic) →
`[role="menuitem"]`. Deleting via automation was **not** reliable — leave deletions to the user.

---

## 8. Repo map

```
go/cmd/whatsapp-tool/main.go   dispatch: quote|price|faq|stock|identity|profiles|pending|triage|keepalive|chat-poll|mcp
go/cmd/whatsapp-tool/pricing.go  whatsapp-tool price — the ONLY pricing write path (set-car/rate/freight/…)
go/cmd/whatsapp-tool/pending.go  whatsapp-tool pending — pending scan (trim+promote+group scan+report)
go/cmd/whatsapp-tool/triage.go   whatsapp-tool triage — batch unread triage (→ data/triage.json)
go/cmd/whatsapp-tool/keepalive.go whatsapp-tool keepalive — start | status | stop
go/cmd/whatsapp-tool/chatpoll.go whatsapp-tool chat-poll — cars-app chat AI loop (polls, generates, replies)
go/internal/mcp/mcp.go         MCP server: tools, schemas, stderr logging
go/internal/wa/browser.go      WhatsApp client: launch/login-state/open/send/read/markAsRead/session
go/internal/wa/types.go        shared WhatsApp types (ChatInfo, Message, Transport, …)
go/internal/dom/dom.go         pure string/DOM parsers + constants (cleanText, isOutgoing, …)
go/internal/chrome/chrome.go   Chrome lifecycle: CDP port, detached spawn, attach, stop
go/internal/statex/            hot state (ai_questions / unanswered_chats): load/trim/archive/promote
go/internal/quote/             deterministic pricing calculator (§13 formulas in code; reads MySQL)
go/internal/db/                MySQL pool + pricing/faq/procurement/todo loaders + writers
go/internal/db/chatbotstate/   shared chat-AI state (migration 009): cursors + single-writer lease
go/internal/db/todo/           task tables (migration 008): store/query/cascade/reminders/users
go/cmd/whatsapp-tool/todo.go   whatsapp-tool todo — the task CLI (add/list/show/.../due/users)
go/internal/config/            centralised env parsing + .env loader (DB_* / CARS_* / CHAT_*)
go/internal/carshttp/          cars app HTTP client (login → token, OS-store cache, retry) + stock/chat SELECT
go/internal/chatbot/           cars-app chat AI: cursor state, prompt build, opencode exec, poll orchestration
go/internal/chatbot/lease.go   single-writer lease heartbeat (renew on TTL/3, abort cycle if lost)
go/internal/keyringx/          OS-native credential store (Credential Manager/Keychain/Secret Service)
go/internal/identity/          code-minted assistant identity (data/ai-identity.json)
go/internal/profiles/          profiles.json registry: list/mark-reply/set-active
go/internal/jsonx,timex/       ordered JSON (SetEscapeHTML=false) + ISO time helpers
go/bin/whatsapp-tool           the built binary (referenced by opencode.json)
AGENTS.md        operational manual — how to drive the tools day to day (read every session)
DEVELOPMENT.md   this file — how the code works (read only when changing code)
ALGERIAN_DERJA.md  living glossary: derja words + context + currency rules (READ before translating)
CANNED_REPLIES.md  standard derja replies to clients (READ before sending anything)
PRICING.md         pricing data schemas + write rules + formulas (READ only when storing numbers)
STATE.md           state-file schemas + group template (READ only when creating/updating entries)
ALGERIAN_DERJA_ARCHIVE.md  glossary history / superseded rows / change logs (not loaded per turn)
data/whatsapp-profile/     Chrome profile (session) — gitignored; per-profile data/<name>/ when WA_PROFILE set
data/whatsapp-cdp.json     pid/port of the shared Chrome — gitignored; per-profile <name>-cdp.json when set
profiles.json              account registry: id/name/number/description/active/linkedAt/lastAiReplyAt
data/state-archive.json    archived answered questions + finished chats (keeps hot files tiny)
db/schema.sql    shared MySQL schema (`ai` database): cars/options/config/freight/clearance/faq + procurement (suppliers/quotations/documents/payments) + todos (todo_users/todos/todo_reminders/todo_status_history)
.env.example     template for the DB_* + CARS_* vars — copy to .env (gitignored)
data/pending.json  full pending report incl. the client's original message (rarely needed)
data/triage.json   batch triage payload (one Read instead of 3 calls per chat)
```

Useful methods on `WhatsAppClient`: `init(headless?)`, `loginState()`, `waitForAppReady(timeout?)`,
`waitForLoginInteractive(timeout?)`, `isQrVisible()`, `isLoggedIn()`, `getChats(limit?)`,
`openChat(nameOrPhone)`, `getCurrentChatName()`, `sendMessage(text)`, `getLastMessages(limit?)`,
`markAsRead()`, `getLastChats(limit?, { timeoutMs?, filter? })`,
`getLastUnreadChats(limit?, { timeoutMs? })`, `restoreAllFilter()`, `sessionInfo()`, `close()`.
Each tool method calls `ensureInit()` first — never assume a page exists.

### Adding a tool

1. Add a public method to the client in `go/internal/wa/browser.go` (it calls `Init()` first).
2. Register it in `go/internal/mcp/mcp.go`: a tool schema, and a handler that `try`/`catch`-style
   returns `{ content: [{ type: "text", text }] }` / `isError: true`. Keep large payloads compact
   (`jsonx.Marshal`) — pretty JSON costs ~20% more tokens for zero information. Use `jsonx`, never
   `encoding/json` (it must not escape HTML; see `go/MIGRATION_GO.md` "Wire format").
3. `cd go && gofmt -w . && go build ./... && go vet ./... && go test ./...`, then rebuild the binary.
4. Update the tool tables in `DEVELOPMENT.md`, `AGENTS.md` and `README.md`.

---

## 9. Token/speed hygiene (why the code is shaped this way)

Every `AGENTS.md`/tool-schema token is re-sent on **every** model request, so a 20-call turn
pays them 20 times. Design rules that come out of that:

- **Deterministic work stays in commands**, not in the model's loop: `whatsapp-tool pending`,
  `whatsapp-tool triage`, `whatsapp-tool quote`. Each returns one compact report the model
  reads once — `pending` carries the boss's answers inline and `quote` carries the arithmetic, so
  the model never re-reads the raw state files or does money math itself.
- **On-demand docs**: `PRICING.md`, `STATE.md` and the archive hold everything that is only
  needed on the rare triggering turn (new numbers, new entries, history). AGENTS.md keeps just the
  trigger + the rule.
- **Hot state files stay empty-ish**: answered/finished entries are archived to
  `data/state-archive.json` by `go/internal/statex`, so the per-turn reads stay small.
- **Node-side cleanup beats page-side**: parse raw DOM strings in Node, keep `page.evaluate`
  minimal, never dump un-filtered markup into a tool result.
- One script yielding a 2 KB report is cheaper than a 20 KB transcript of tool calls doing the
  same thing — tokens AND wall-clock.

---

## Shared MySQL database (`ai`) — pricing + FAQ

Shared, whole-team state lives in MySQL; everything else stays in local JSON. This is the
deliberate split:

| Shared (MySQL `ai`) | Local per machine (JSON) |
|---|---|
| `cars`, `car_options`, `car_colors` | WhatsApp session (`data/whatsapp-profile/`, `data/<name>/`) |
| `car_price_history` | `ai_questions.json`, `unanswered_chats.json` (queue) |
| `options`, `config` | `profiles.json` + `data/*-cdp.json` |
| `freight` (+ `loading_ports`/`discharge_ports`/`shipping_lines`), `clearance`, `clearance_aliases` | run reports: `pending.json`, `triage.json`, `autoreply*.json` |
| `suppliers`, `supplier_quotations`, `supplier_quotation_items`, `supplier_quotation_item_prices`, `documents`, `document_quotations`, `supplier_payments`, `payment_quotations`, `payment_swifts` | `state-archive.json` |
| `faq` | `state-archive.json` |

Why: pricing/FAQ must be identical for the whole team, but a WhatsApp login, its CDP port and the
question queue are physical to one machine (and the session is a secret). Accounts run on
different machines, so the shared **`ai questions`** group (not the DB) bridges the queues.

Schema/scalability notes (`db/schema.sql`): InnoDB + `utf8mb4`; surrogate `BIGINT` PKs + a unique
business key (`car_id`, option `name`, supplier `name`); children normalised (`car_options`,
`car_colors`, `clearance_aliases`, `loading_ports`/`discharge_ports`/`shipping_lines`, the
per-colour `supplier_quotation_item_prices`); append-only history (`car_price_history`,
`supplier_quotations`) and one config time series (`config`, keys `benefits`/`rate`/`customs`,
newest `effective_at` wins); DECIMAL money; indexed lookups; FULLTEXT on `faq`.

Data flow: `go/internal/config` (env/`.env`) → `go/internal/db` (pool) → the pricing + FAQ loaders →
consumers `go/internal/quote`, `go/cmd/whatsapp-tool/pricing.go`, `go/cmd/whatsapp-tool/main.go`
(faq). Keep the db layer pure I/O; business rules stay in `quote` / the CLI.

Row-level `mysql2` pool options matter: `decimalNumbers` (money as numbers), `dateStrings` +
`timezone: "Z"` (UTC, formatted by `toIso`). Schema changes go in `db/schema.sql` (idempotent,
tracked in `schema_migrations`). The original JSON catalog is archived (and removed from the repo)
at `backups/pricing-json-backup-*.tar.gz`. The one-time JSON→MySQL seed was done with the old Node
`db:seed` script (no Go equivalent).

### Procurement — suppliers, quotations, documents & payments (the ai database)

`whatsapp-tool procurement ...` manages the ai DB's supplier-quotation-document-payment system
(`suppliers`, `supplier_quotations`, `supplier_quotation_items`,
`supplier_quotation_item_prices`, `documents`, `document_quotations`, `supplier_payments`,
`payment_swifts`). It is **linked to the
pricing catalog**: `cars.supplier_id` (→ `suppliers`), `car_price_history.supplier_id`/`quotation_id`,
`supplier_quotation_items.car_ref` (→ `cars`, SET NULL, set via `add-item --car-ref <carId>`),
and each `documents` row's `supplier_id`. `source_text` keeps the full transcription of a real
document (PI / sales contract / …) so it stays readable by the AI; `document_quotations` links a
document to the quotation(s) it backs (M:N), closing the chain
document → supplier_quotations → items → per-colour prices → cars. `supplier_payments` records what
we actually paid a supplier, linked through the **M:N junction `payment_quotations`** to the
quotation(s)/contract(s) it settles (one payment can cover several invoices — migration 006);
`payment_swifts` (1:1, CASCADE) stores the MT103 proof — `file_url` (the PDF link), `source_text`
(the full manuscript, AI-readable) plus extracted `swift_ref`/`uetr`/BICs/beneficiary. It is
unrelated to the cars-app suppliers/inquiries; a quotation's `quoted_at` is the timestamp, and each
item carries one `unit_price` per colour (`supplier_quotation_item_prices`).
`price set-car --supplier "<name>" [--quotation <id>]` attributes a buy price to its supplier
(`FindSupplierID` resolves the name) and stamps the appended `car_price_history` row.
**Balance = deals − payments**: `outstanding` sums each supplier's deals (one amount per quotation —
a deal's PI *and* its sales contract count once, both linking the same quotation) and subtracts the
recorded payments (`status` sent/cleared). "How much do we still owe <supplier>" must always be
answered from this command, never from the documents alone.

- `whatsapp-tool procurement suppliers [--name X] [--limit N]` — list ai suppliers
- `whatsapp-tool procurement suppliers add --name X [--contact ..] [--notes ..]`
- `whatsapp-tool procurement suppliers update --id N [--contact ..] [--notes ..]` — set contact/notes (non-empty fields only)
- `whatsapp-tool procurement quotations [--supplier-id N] [--status S] [--ref X] [--limit N]`
- `whatsapp-tool procurement quotations show <id>`
- `whatsapp-tool procurement quotations create --supplier-id N [--date ..] [--ref ..] [--currency USD] [--status ..] [--valid-until ..] [--notes ..]`
- `whatsapp-tool procurement quotations add-item --quotation-id N --car "<name>" (--big|--small) (--used|--new) [--qty N] [--year ..] [--month ..] [--car-ref <carId>]`
- `whatsapp-tool procurement quotations set-price --item-id N --color "<color>" --price N [--notes ..]`
- `whatsapp-tool procurement quotations set-status <id> --status S`
- `whatsapp-tool procurement quotations delete <id>`
- `whatsapp-tool procurement documents [--supplier-id N] [--type T] [--limit N]` — list documents
- `whatsapp-tool procurement documents show <id>` — document + its source_text + linked quotations
- `whatsapp-tool procurement documents create --no "PIXSLNR..." --type proforma_invoice --supplier-id N [--date YYYY-MM-DD] [--counterparty ".."] [--currency USD] [--total N] [--words ".."] [--text ".."|--text-file <path>] [--status received] [--notes ..]`
- `whatsapp-tool procurement documents link --document-id N --quotation-id M`
- `whatsapp-tool procurement documents unlink --document-id N --quotation-id M`
- `whatsapp-tool procurement payments [--supplier-id N] [--quotation-id N] [--status S] [--limit N]` — list payments
- `whatsapp-tool procurement payments show <id>` — payment + its swift proof + linked quotations
- `whatsapp-tool procurement payments create --supplier-id N --amount V [--currency USD] [--date YYYY-MM-DD] [--value-date ..] [--method swift] [--status sent] [--notes ..]` — `payment_no` is auto-minted as the next **sequential integer** (`NextPaymentNo` = max numeric `payment_no` + 1); the bank's own reference, if any, goes in `set-swift --swift-ref`, never the key
- `whatsapp-tool procurement payments edit --id N [--payment-no ..] [--supplier-id N] [--amount V] [--currency ..] [--date ..] [--value-date ..] [--method ..] [--status ..] [--notes ..]`
- `whatsapp-tool procurement payments link --payment-id N --quotation-id M` — payment settles a quotation (idempotent)
- `whatsapp-tool procurement payments unlink --payment-id N --quotation-id M`
- `whatsapp-tool procurement payments delete --id N` — also removes its swift row + quotation links (CASCADE)
- `whatsapp-tool procurement payments set-swift --payment-id N [--swift-ref ..] [--uetr ..] [--sender-bic ..] [--receiver-bic ..] [--beneficiary-name ..] [--beneficiary-account ..] [--file-url <link>] [--file <path>] [--text ".."|--text-file <path>]` — upsert (1:1)
- `whatsapp-tool procurement outstanding [--supplier-id N]` — per supplier: `owed` (Σ deals, one amount per quotation — a deal's PI + contract count once), `paid` (Σ payments with status sent/cleared), `outstanding = owed − paid`, plus `deals[]`/`payments[]` for audit (`go/internal/db/procurement/balances.go`)

Code: `go/internal/db/procurement` (loaders/writers) + `go/cmd/whatsapp-tool/procurement.go`.

### Tasks & reminders — the ai database

`whatsapp-tool todo ...` manages the ai DB's task tables (migration `008`): `todo_users`
(a WhatsApp login — `profile` UNIQUE), `todos` (self-referencing `parent_id` = unlimited subtask
tree, `owner_id` → `todo_users`, `status` open/done/cancelled, `priority` low/normal/high/urgent,
nullable `due_at`), `todo_reminders` (`remind_at`, `channel`, `status` pending/sent/cancelled,
`sent_at`) and the append-only `todo_status_history`. The **owner is the login**
(`config.WAProfile()`; `EnsureUser` upserts it on first use), so todos are per-account — the CLI
shows only the current login unless `--owner`/`--all`.

**Cascade rule** (`cascade.go`, transactional): a child is *resolved* = done **or** cancelled; a
parent with children is `done` iff every child is resolved (adding an open child under a done parent
**reopens** it and its ancestors); a childless todo keeps its manual status. `move` re-derives both
the old and the new ancestor chain and rejects a move under a descendant; `delete` cascades the
subtree (FK) then recomputes the ex-parent. `SetStatus`/`Move`/`Create`/`Delete` write a
`todo_status_history` row (`auto = 1` for cascade changes).

**Reminders have no daemon**: `Due()` returns pending reminders whose `remind_at` has passed for
still-open todos; it is surfaced as the `todosDue` block of the `pending` report
(`cmd/whatsapp-tool/pending.go`) and by `todo due`. There is no scheduler process.

All writes require `can_give_edit_instruction` (`requiredPermission` in `main.go`); the reads are
open.

- `whatsapp-tool todo add --title X [--parent N] [--desc ..] [--priority P] [--due D] [--requested-by ..] [--chat ..] [--owner <profile>]`
- `whatsapp-tool todo list [--owner <profile>|--all] [--status S] [--priority P] [--overdue] [--due-before D] [--roots] [--parent N] [--limit N] [--tree]`
- `whatsapp-tool todo show <id>` — todo + children + reminders + history
- `whatsapp-tool todo done|reopen|cancel <id>`
- `whatsapp-tool todo edit <id> [--title ..] [--desc ..] [--priority ..] [--due ..|--clear-due] [--requested-by ..] [--chat ..]`
- `whatsapp-tool todo move <id> --parent <id|none>`
- `whatsapp-tool todo assign <id> --owner <profile>`
- `whatsapp-tool todo delete <id>` — removes the todo and its whole subtree
- `whatsapp-tool todo remind add <id> --at "YYYY-MM-DD HH:MM" [--channel whatsapp|local] [--note ..]`
- `whatsapp-tool todo remind list [--todo N] [--status pending|sent|cancelled] [--limit N]`
- `whatsapp-tool todo remind rm <id>` / `todo remind sent <id>`
- `whatsapp-tool todo due [--before D]`
- `whatsapp-tool todo users`

Code: `go/internal/db/todo` (`todo.go` types/consts, `users.go`, `store.go`, `query.go`,
`cascade.go`, `reminders.go`, `parse.go`) + `go/cmd/whatsapp-tool/todo.go`.

## Car stock — the cars app (separate system)

Live stock is **not** in the `ai` database. It lives in the **cars app**, and the tool reaches it
over that app's HTTP API (`/<tenant>/api/api.php`) because the tool's MySQL user (`ai_app`) is
scoped to the `ai` schema and cannot read tenant databases.

Auth (see `go/internal/carshttp/`): the app has no sessions — an admin
creates a dedicated **AI user**, and we call its public `login` action to receive a token that the
server keeps in `users.api_token` (`api/lib/auth.php`). One login invalidates the previous token,
so the token is cached in the **OS credential store** (Go keyring — Windows Credential
Manager / macOS Keychain / Linux Secret Service) and reused; a `not_authenticated` reply triggers
one re-login + retry. The AI user's username/password live in the store too, so `.env` holds no
secrets — only `CARS_BASE_URL`/`CARS_TENANT` (and an optional `CARS_API_TOKEN` override). If the
store is empty, the CLI reports `code: missing_credentials` so the agent asks the user, then seeds
it via `whatsapp-tool stock creds --username <user>`. Secrets are never written to disk; the store is
fail-closed (an unavailable store is an error, not a plaintext fallback). `saveAndVerifyCredentials`
persists credentials only after the login succeeds.

The AI user is a real app user, so the app's role model applies: the `login` envelope carries the
user's `role_name` and the role's permission list (server-side `auth/auth.php:auth_permissions`
joins `permissions`/`role_permissions` by `role_id`) — the Go client caches those names in the
store at every fresh login (`carshttp.cacheRoleAndPermissions`), and the stock command enforces
the app's own stock-view permission (`can_c_car_stock`) via `carshttp.HasPermission` unless an
explicit `CARS_API_TOKEN` was supplied. A missing one fails with `code: missing_permission`. The
tenant's AI account runs under a dedicated role (`ai`) holding only `can_c_car_stock` — separate
from `admin`, which grants everything.

Scope is deliberately read-only **except for brand/car-name/supplier creation, the inquiries writes, the clearance upsert and the buy-payment add/edit/delete writes**:
`go/internal/carshttp/stock.go` builds one fixed `SELECT` (no caller SQL), and filters bind as
parameters. The join set mirrors
`src/components/car-stock/CarStockTable.vue` so the agent sees the same Available/Sold view staff
do.

**The writes** are `whatsapp-tool brands add` (`go/internal/carshttp/brandcreate.go`, `write.go`,
`upload.go`), mirroring `useBrands.js` `addBrand()`; `whatsapp-tool car-names add`
(`go/internal/carshttp/carnamecreate.go`, `cmd/whatsapp-tool/carnameadd.go`), mirroring
`useCarNames.js` `addCarName()`; `whatsapp-tool suppliers add`
(`go/internal/carshttp/suppliercreate.go`), mirroring `SuppliersView.vue` `addSupplier()`; and the
`whatsapp-tool inquiries` create/add-car/delete commands (`go/internal/carshttp/inquiry.go`,
`inquirycreate.go`, `cmd/whatsapp-tool/inquiries.go`), and `whatsapp-tool clearance set`
(`go/internal/carshttp/clearance.go` for the read, `clearanceset.go` for the upsert,
`cmd/whatsapp-tool/clearance.go`) — an `INSERT ... ON DUPLICATE KEY UPDATE` on the
`(id_car_name, is_big_car, is_used_car)` unique key, so setting an existing model/variant updates it
in place. For brands, two channels, both fixed and
parameterised: a logo (if `--logo <path>`) goes up as multipart to the tenant's `api/upload.php`
(`base_directory=files`, `destination_folder=uploads`, token as a field — a form body cannot set
`X-Api-Token`), and the row goes in over the same `action=query` passthrough the app uses, whose
INSERT branch answers `{success, lastInsertId}` (parsed leniently by `flexInt64`, since PDO returns
it as a JSON string). `carshttp.ExecAsAI` is the write twin of `QueryAsAI`, sharing the token retry
via `withToken`; `brands add` / `car-names add` / `inquiries` are gated on `can_purchase_cars` (the
models/buy tabs), and `suppliers add` on `can_c_suppliers` (the suppliers tab); `clearance` (read +
`set`) is **ungated** — any logged-in AI user may use it.

Several fields are **required at the DB level** (migration `035`: `cars_names.is_big_car`,
`inquiry.date_inquiry`, `inquiry_details.id_color`/`QTY`/`target_price`/`is_used_car`), so the CLI
refuses rather than letting MySQL default them: `car-names add` requires `--big` or `--small`
(mutually exclusive, read by the `boolChoice` helper in `main.go`), `inquiries add-car` requires
`--color-id`, `--qty`, `--target-price` and `--used`/`--new`, and `inquiries create` defaults
`date_inquiry` to now (`time.Now().Format("2006-01-02 15:04:05")`). `clearance set` likewise
requires `--car-name-id`, `--dza` and explicit `--big`/`--small` + `--used`/`--new` (the last two are
part of the table's unique key, so they are never defaulted). The Vue forms supply the same
fields (`CarStockForm.vue`, `CarNameFormDialog.vue`).

The inquiry tables (`inquiry` / `inquiry_details`) mirror `buy_bill` / `buy_details` (migrations
`034`/`035`); `inquiry_details.id_inquiry` carries `ON DELETE CASCADE`, so `DELETE FROM inquiry WHERE
id = ?` removes the header and its lines in the one statement the query passthrough allows.
`INSERT` column lists are chosen from fixed literals by `insertQuery` — never built from input —
and every value binds as a parameter. The `clearance` table (migration `036`) is per-model/variant
customs clearance (DZD, TVN excluded) with a `(id_car_name, is_big_car, is_used_car)` unique key;
its seeded values live in the migration itself (reference data), so it is in
`TENANT_SEED_TABLES`, not `TENANT_BUSINESS_TABLES`. The buy-payment reads
(`go/internal/carshttp/buypayments.go`, `cmd/whatsapp-tool/buypayments.go`) mirror
`BuyBillPaymentsView.vue`: `buy_payments` joined to `buy_bill` / `suppliers` / `users`, with the
bill header carrying `calculated_amount` (Σ amount×QTY over `buy_details`) and `total_paid`; gated
on `can_pay_purchases` (a dedicated permission, migration `037`; payments to suppliers are buy-cost
figures). Its writes (`go/internal/carshttp/buypaymentcreate.go`) mirror `savePayment()` (add and
edit) + `handleDeletePayment()` + `updateBillPaidAmount()`: `add` a fixed `INSERT INTO buy_payments`
(obligatory `--bill-id`, `--amount`, `--date`, `--swift`, the swift uploaded to `buy_payment_swifts`
via `UploadFileNamed` first; `id_user` from the id cached at login, see `carshttp.CurrentUserID`),
`edit` a fixed `UPDATE buy_payments SET amount, date_payment, swift_path, notes WHERE id` (swift
optional — omitted keeps the stored document via `GetBuyPayment`; notes likewise), and `delete` a
fixed `DELETE FROM buy_payments WHERE id`. Each then runs `UPDATE buy_bill SET payed = (Σ of that
bill's payments)` for the payment's owning bill.

Data flow: `go/internal/config` (env) → `go/internal/keyringx` (credentials/token) + `go/internal/carshttp`
(HTTP) → the stock query → the CLI. Keep the HTTP layer pure I/O; filters
stay in the stock command.

## Local AI (Ollama) — optional, fully offline replies

Everything above works with any opencode model. To run the agent **100% local** (no cloud model,
client chats never leave the machine), point opencode at local Ollama models:

- The agent's job here is **not coding** — it reads Derja/Arabic chats, follows `AGENTS.md` and
  replies through the WhatsApp tools. So pick for **Arabic/dialectal strength + reliable tool
  calling**, not SWE-bench.
- Recommended: `qwen3.6:27b` (dense; strong on Maghrebi/dialectal Arabic + tool calling).
  Fast fallback: `gemma4:26b` (MoE, very fast, also native function-calling).
  Arabic-native models (ALLaM/Jais/Fanar) are stronger at MSA but their tool-calling is weak —
  they break the WhatsApp tool loop. Rejected as the driver.
- Setup (needs `registry.ollama.ai` reachable — VPN/proxy up):
  `bash ollama/setup.sh` — pulls both base models and creates the `-ctx64` derived models
  (Modelfiles in `ollama/` pin `num_ctx 65536`, because opencode needs ≥64K while Ollama defaults
  to 4096). ~36 GB disk.
- `opencode.json` already carries the `ollama` provider (`@ai-sdk/openai-compatible` →
  `http://localhost:11434/v1`) with both derived models. To make it the default for new sessions,
  add at the top: `"model": "ollama/qwen3.6:27b-ctx64"` (and optionally
  `"small_model": "ollama/qwen3.6:27b-ctx64"` so session titles stay local too) — do this **after**
  the models are pulled, or opencode will refuse to start a session.
- Switch per session anytime with `/models`; the cloud provider stays configured as a fallback.
- If tool calls drop out in a session, that is usually context pressure — increase `num_ctx`
  in the Modelfile and re-run `ollama create`.

## Chat AI daemon (`chat-poll`) — one tenant, many machines

`whatsapp-tool chat-poll` runs the **cars-app chat AI**: each cycle it lists the groups the AI user
belongs to, reads every client message newer than the per-group cursor, generates one reply per
group with the opencode CLI (see §9/Local AI), and posts it back as the AI user via the cars-app
API. It never drives WhatsApp Web (that is the `whatsapp_*` path).

**Several machines may run it against the same tenant; exactly one is active at a time.** The
coordination is a **single-writer lease** + **shared cursors** in the `ai` DB (migration 009):

- `chatbot_leases` — one row per tenant. The active daemon claims it (`owner = host:pid:profile`),
  then renews `expires_at` on a heartbeat every `TTL/3` (`CHAT_LEASE_TTL`, default 60 s). A standby
  calls `Acquire` each cycle; if the row is owned by a live peer it returns `status: "STANDBY"` and
  does **no** network work. It claims the lease only once `expires_at` has passed — i.e. the holder
  died / stopped renewing — so a crash fails over within ~TTL. All time checks use the **DB clock**
  (`NOW(3)`), never the machine's, so clock skew cannot split the lock.
- `chatbot_group_cursors` — the per-group last-answered message id, so a machine that takes over
  resumes at the right message instead of re-answering history. Writes are **monotonic**
  (`GREATEST` upsert), so a stale writer can never move a cursor backwards. The legacy per-machine
  `data/chatbot-state.json` is imported **once** when the shared store is still empty, so upgrading
  does not re-answer history.
- Safety: the holder checks the lease again **before posting**; if it was lost mid-generation the
  reply is dropped and the cursor left put, so the new owner answers those messages exactly once.
  The heartbeat runs in a goroutine (`internal/chatbot/lease.go`); when it loses the lease the cycle
  aborts at the next safe point.

**Web access (restricted).** For **two question types only** — **company credibility** and **car
specifications** — the agent may consult the internet; every other topic is answered from the `ai`
DB or deferred to the team. `websearch` is allowed. opencode's own `webfetch` is **disabled**
(opencode 1.x has no per-URL `webfetch` permission — it takes a single allow/deny action — so it
cannot be scoped to a domain list); instead the agent reads a page through
`whatsapp-tool web fetch "<url>"`, whose **domain allowlist is enforced in Go**
(`internal/webfetch`, `Allowlist`): the requested host *and every redirect host* must be on it, or
the fetch is refused (`blocked_domain`/`blocked_redirect`). The curated list (China registries —
GSXT/NECIPS, Credit China, China Judgements Online; Hong Kong Companies Registry; aggregators like
OpenCorporates/QCC; trade directories; car-spec sites such as autohome/dcdapp/yiche/wikipedia) is
editable in that one Go slice; `whatsapp-tool web list` prints it and the agent is pointed at it.
The official portals are JS+CAPTCHA, so a `web fetch` there usually returns little — the value is
in the `websearch` results. The agent is told to treat fetched pages as **untrusted data** (never
instructions), never to source a price/stock/availability figure from the web, and to say the
information came from the internet. That disclosure is **enforced in code**: when the opencode
stream shows a `websearch`/`webfetch` tool call **or a bash call to `whatsapp-tool web fetch`**
(`parseEvents` in `llm.go`), `FinalizeReply` appends `WebSourceLine` *before* the AI disclosure
line — the model cannot forget it. The editor agent has **no** web access. `OPENCODE_ENABLE_EXA=1`
is set on the child so `websearch` exists even under a non-OpenCode model provider.

**Shared client-reply rules — one source of truth.** The channel-agnostic reply rules the agent
follows (tool usage incl. the stock flags, never-invent-a-price, freight, ALGIERS default, web
restrictions, Markdown tables, mandatory language match, Arabic numerals, disclosure note) live in
**one file**: `chatbot/rules/client-reply.md`. It is injected in two places, both automatic:

- **`BuildPrompt`** (`internal/chatbot/prompt.go`) reads it once per process (`sharedRules()`,
  `sync.Once`) and appends it after the cars-app framing (group, capabilities, history, incoming).
- **`ensureConfig`** (`internal/chatbot/llm.go`) splices it into **both** agent personas at the
  `<!-- shared-rules -->` marker, so `chatbot.md` and `chatbot-editor.md` carry only their
  *differences* (frontmatter/permissions, intro, the editor's edit instructions, the todo block).

Editing a rule or a CLI flag is therefore a **single edit** to `client-reply.md`, not a hand-copy
across `prompt.go`, `chatbot.md` and `chatbot-editor.md`. Two tests pin this
(`TestSharedRulesSourceOfTruth`, and the injection assertions in `TestEnsureConfigInstallsAgents`);
if the marker or the file goes missing, the tests fail and (at runtime) `sharedRules()` falls back
to a short safe stub rather than answering unguided. `AGENTS.md` §13/§15 remains the operator
manual for the WhatsApp path; where a rule is executable by the chat AI, `client-reply.md` is the
source of truth.

Deploying to another machine: same `.env` (`DB_*`, `CARS_BASE_URL`, `CARS_TENANT`) pointing at the
**same** `ai` DB, its own `whatsapp-tool stock creds` (OS keyring), opencode + a model, then just
run `whatsapp-tool chat-poll` and leave it running. No primary/secondary config — whichever holds
the lease answers. `CHAT_LEASE_OWNER` only overrides the owner label (observability).

## 10. Multiple WhatsApp accounts (`WA_PROFILE`)

One server instance can talk to exactly one WhatsApp account (one Chrome profile = one login).
To run several accounts side by side, start the **single** server once per account *invocation*,
steering it with `WA_PROFILE` (see the MCP note below) and ideally pinning `WA_CDP_PORT`.

All profile-scoping lives in `go/internal/chrome/chrome.go` and `go/internal/statex/`:

- `PROFILE` / `PROFILE_IS_DEFAULT`: `WA_PROFILE` unset (or literally `whatsapp-profile`) keeps the
  classic account untouched — `data/whatsapp-profile/`, `data/whatsapp-cdp.json`,
  `ai_questions.json`, `unanswered_chats.json`, `data/{pending,triage,state-archive}.json`.
- Any other value `<name>` redirects to `data/<name>/`, `data/<name>-cdp.json`,
  `data/<name>-questions.json`, `data/<name>-chats.json`, `data/<name>-archive.json`,
  `data/<name>-pending.json`, `data/<name>-triage.json` (`runFile()`).
- Ports: if `WA_CDP_PORT` is set, every instance honours it; otherwise each process scans the
  first free port in 9222–9252 and records it in **its own** state file, so two instances
  naturally land on different ports. **Pin the port anyway**: the free scan only looks at "CDP
  alive", and you do not want an unattached instance silently adopting a neighbour's port after
  a restart.
- The profile lock (`profileLockedByOther`) is per `DATA_DIR`, so two accounts can have a
  `.link`/keepalive running at the same time — same account, same lock, always.

What is **shared** across accounts (deliberately): the pricing catalog + FAQ (now in the shared
MySQL `ai` database), the glossary, and every guide — one business, several phones.

MCP side (opencode): a **single** `mcp.whatsapp` entry whose `environment` maps the caller's var —
`"environment": { "WA_PROFILE": "{env:WA_PROFILE}" }` (already in `opencode.json`). Whoever runs
`opencode` steers the account from the shell and the tool names stay `whatsapp_*` for every
profile:

```bash
opencode run --auto "…"                 # default account
WA_PROFILE=business opencode run --auto "…"   # business account only
```

There is deliberately **no** per-account MCP entry: two live servers would both expose
`whatsapp_*` sets and let the model pick — the whole point is that the shell, not the model,
chooses. Local MCP servers inherit opencode's env, and the `{env:WA_PROFILE}` mapping also
carries it explicitly (unset → empty string → default account). The model's own Bash tool inherits
the same var, so `whatsapp-tool pending` reports match the active account.

Lifecycle per account:

```bash
WA_PROFILE=business WA_HEADLESS=0 whatsapp-tool keepalive start   # headed QR window, data/business/
WA_PROFILE=business whatsapp-tool keepalive start                 # warm Chrome; status/stop as usual
WA_PROFILE=business whatsapp-tool pending                         # that account's queue only
```

New accounts are linked **manually** — the old Node `connect` tool is gone and is never to be run
(see the strict rules below). The user drives the visible-Chrome link flow; registration in
`profiles.json` is by hand (`go/internal/profiles/`).

`sessionInfo()` now reports `profile`, and every startup log prefixes `profile=<name>` — the
first thing to check when two instances feel slow is that each is on its own port/profile.

### State and profiles
- `AiQuestion`, `PendingChat`, `QuestionContext` include `profileId?: number | null` (null = default account).
- `getCurrentProfileId()` returns the profiles.json id for the current `WA_PROFILE`, else null.
- When new state entries are created in future code, set `profileId = getCurrentProfileId()`.


**Strict rule: never delete any profile (from profiles.json, data/* for that profile, CDP state) without explicit confirmation from the user.**

**Strict rule: never register/link accounts programmatically. Do not run any account-linking/`connect` command. The user connects WhatsApp manually.**

**STRICTEST RULE: DO NOT delete, remove, drop, or modify profiles.json entries, profile folders (data/<name>/), CDP files, or any account-related state WITHOUT EXPLICIT USER CONFIRMATION. Do not run any account-linking/connect command.**

## Code Modularity, Scalability, Maintainability & AI-Friendly Guidelines

### File size & organization
- **Keep files under 800 lines.** If a file approaches 800 lines, refactor into smaller, focused modules.
- **Single responsibility:** Each module should do one thing well (e.g., chrome management, whatsapp client, state management, CLI commands).
- **Group by domain:** Put related code in subdirectories (e.g., `go/internal/chrome/`, `go/internal/wa/`, `go/internal/statex/`, `go/internal/mcp/`) to avoid a flat, crowded tree.
- **Avoid god files:** Split large classes/functions into smaller units. Extract helpers, types, and constants into their own files.

### Scalability
- **Explicit interfaces:** Define types/interfaces for all data structures (profiles, state, quotes, messages). Co-locate types with the module that owns them (Go: the package).
- **Dependency injection friendly:** Avoid hidden globals; pass dependencies explicitly where practical to make testing and extension easier.
- **Feature flags/env:** Centralize environment parsing in `go/internal/config/` (e.g., profile, ports, headless). Never read `os.Getenv` ad-hoc in deep business logic—read once and pass values.
- **Stateless where possible:** Separate pure logic (pricing, parsing, normalization) from I/O (filesystem, browser, network). Pure modules are easy to test and reuse.
- **Extensible registries:** When adding new commands/tools, prefer a registry/loader pattern rather than hardcoding long switch statements in a single file.

### Maintainability & readability
- **Consistent naming:** Use clear, descriptive names. Avoid abbreviations unless domain-standard.
- **Small functions:** Keep functions short and focused. Extract complex conditionals into named predicates.
- **DRY:** Eliminate duplication; extract shared utilities into `go/internal/`.
- **Comments:** Only add comments when necessary to explain "why", not "what". Avoid noisy comments.
- **Error handling:** Use typed errors, meaningful messages, and consistent logging (stderr for diagnostics). Never log sensitive data.
- **Formatting & linting:** `gofmt -l .` clean + `go vet ./...` + `go test ./...` after changes.

### AI-friendly practices
- **Predictable structure:** Keep module boundaries clear so an AI can reason about changes without reading the entire codebase.
- **Minimal context:** Prefer smaller files and focused exports. Avoid circular dependencies.
- **Explicit contracts:** Export clear APIs from each module; document key types and functions with JSDoc sparingly but usefully.
- **Deterministic behavior:** Avoid side effects at import time where possible (especially for browser/port selection). Defer initialization to explicit functions.
- **Safe defaults:** Fail fast with clear errors; validate inputs early.
- **Human-in-the-loop safety:** Keep destructive operations (delete/modify profiles, mass messaging) behind explicit confirmation and well-isolated guards.
