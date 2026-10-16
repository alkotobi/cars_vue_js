# REFACTOR.md — separating the reply domains (shared core + 2 channels)

> Status: **Phases 1 ✅ 2 ✅ 3 ✅ done.**

## Context

The tool answers clients through what looks like three "domains":

1. **Reply WhatsApp** — `internal/channel/whatsapp/{wa,mcp,dom}` + the `pending`/`triage`/
   `keepalive` commands. Driven by the opencode session (the operator), **human-gated** (ToS/ban
   risk; AGENTS.md §0 confirm-before-send).
2. **Reply general questions** — the **personas** the cars-app chat AI uses
   (`chatbot/agent/chatbot.md`, `chatbot/agent/chatbot-editor.md`), selected by group permissions.
3. **Reply cars-app chat** — `internal/channel/carsapp/chatbot` + `internal/carshttp/chat.go` +
   the `chat-poll` daemon (single-writer lease, autonomous posting).

They are **not** three equal services. They are **2 channels + 1 shared capability**:

- Channels: **whatsapp** (human-gated) and **carsapp** (autonomous daemon).
- Shared capability: general Q&A (pricing, freight, stock, FAQ, web) + the reply rules — consumed
  by both channels.

## Decision

**Separate by layer, not into 3 silos.** Keep 2 channel adapters, extract one shared core, and
treat "general questions" as a persona/capability — not a top-level domain.

Explicitly **not** doing: three independent binaries/services. The single-writer lease, the
one-at-a-time Chrome profile, and the shared `ai` DB all assume one coordinated tool; three
services would mean three deploys to both production servers, three leases, three failure modes.

## Phase 1 — one source of truth for the shared rules ✅

The same business rules were hand-written in four places (`AGENTS.md`, `prompt.go`, `chatbot.md`,
`chatbot-editor.md`), so one flag change (the stock date filters) required 3–4 edits.

- New `chatbot/rules/client-reply.md` — the channel-agnostic rules (tool usage incl. stock flags,
  never-invent-a-price, freight, ALGIERS default, web restrictions, Markdown tables, mandatory
  language match, Arabic numerals, disclosure note).
- `BuildPrompt` (`internal/channel/carsapp/chatbot/prompt.go`) injects it via `sharedRules()`
  (read once, `sync.Once`; falls back to a short safe stub if the file is unreadable).
- `ensureConfig` (`internal/channel/carsapp/chatbot/llm.go`) splices it into **both** personas at
  the `<!-- shared-rules -->` marker, so the personas carry only their differences.
- Drift guards: `TestSharedRulesSourceOfTruth`, and the injection assertions in
  `TestEnsureConfigInstallsAgents`.

Result: a rule/flag change is **one edit**.

## Phase 2 — name the seams (clarity, mechanical) ✅

Namespace the **adapters** so the boundary is visible:

- `internal/channel/whatsapp/{wa,mcp,dom}` ← `internal/{wa,mcp,dom}`
- `internal/channel/carsapp/chatbot` ← `internal/chatbot`

Package names (`wa`, `mcp`, `dom`, `chatbot`) were **kept**, so only import paths changed — 7 Go
files, then `gofmt`. `go build/vet/test ./...` green; docs (DEVELOPMENT.md §8 repo map, README,
AGENTS.md) updated.

**`internal/core/…` was deliberately NOT created.** Everything that is *not* under
`internal/channel/` already *is* the shared core — `carshttp`, `quote`, `webfetch`, `db/*`,
`groupauth`, `config`, `redact`, … Moving `quote`+`webfetch` into a `core/` folder while leaving
`carshttp` and `db/*` outside would split them arbitrarily (especially `db/{faq,pricing,
procurement}` vs `db/{todo,chatbotstate,groupperms}`), i.e. *less* clarity for the same churn. The
seam is expressed by `internal/channel/*` being the only thing with a `channel/` parent — see the
**§8 repo map**, which now labels `# channels (adapters)` vs `# shared core`.

Known imperfection, **addressed in Phase 3 without moving the file**: the cars-app channel's
transport (`carshttp/chat.go`) still lives in the core package because it shares the HTTP client
with the read-side lookups — but `chatbot` now reaches it only through `Transport`
(`transport_cars.go`), so the loop itself is transport-free.

## Phase 3 — the loop/transport split (scalability + tests) ✅

`PollOnce` used to talk to `carshttp` directly, so the loop could not be driven by any other
destination and could not be tested at all (`bot.go` had zero tests). It now coordinates and
nothing else — every destination call goes through an injected **`Transport`**:

```go
// internal/channel/carsapp/chatbot/transport.go
type Transport interface {
    Identity() (int64, bool, error)             // who the channel's sender is
    Groups(userID int64) ([]Group, error)       // what to listen to
    Perms(groupID int64) (Perms, error)         // capabilities (cars-app: gperms row)
    Read(groupID, afterID, excludeUserID int64) ([]Message, error)
    History(groupID int64, limit int) ([]Message, error)
    MaxID(groupID int64) (int64, error)         // first-time cursor seeding
    Latest(groupID int64) (Message, bool, error)// seed: stale vs fresh window
    Send(groupID, userID int64, text string) error
}
```

- **`Message`/`Group` are the loop's own value types** (JSON tags match the cars-app row), so
  `prompt.go` no longer imports `carshttp` — `formatMessage(m Message)` reads `m.Text`.
- **`transport_cars.go` is the only file in the package that imports `carshttp`** — that is how the
  "transport lives in the core" imperfection was addressed without moving `carshttp/chat.go`
  (it shares the HTTP client with the read-side lookups).
- **Strict parsing stays strict:** a malformed row in `Read` is an error, so the cursor does not
  move and the messages are retried next cycle (never silently dropped). A `History` failure is
  tolerated by `historyOf` — the reply still goes out with no context block.
- **`generateFn = Generate`** is the test seam (same pattern as `sharedBackend`/`dataDir`): tests
  drive the whole pipeline without spawning the opencode CLI.
- **`bot_test.go`** (14 tests) pins the loop's behaviour: cursor advanced to the newest *answered*
  id, unchanged on read/generate/lease-loss failures, advanced-before-send (a failed post is a
  miss, not a duplicate), empty reply skipped, first-seen group seeded (`MaxID`/`Latest` +
  `Chatbot.SeedWindow`) instead of answered, permissions threaded into the brain.

`cmd/whatsapp-tool/chatpoll.go` passes `chatbot.NewTransport()`. The WhatsApp channel still has no
polling loop by design (human-gated, §Context), so `Transport` — not `Channel` with `Poll` — is the
contract: **autonomy stays a per-channel policy**, and a second autonomous channel is a new
`Transport` implementation plus its own driver, not a fork of the loop. A `Name()` on the transport
was deliberately left out until a second implementation needs it.
