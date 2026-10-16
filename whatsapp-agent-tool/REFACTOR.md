# REFACTOR.md — separating the reply domains (shared core + 2 channels)

> Status: **Phase 1 done.** Phases 2–3 are the agreed direction, not yet started.

## Context

The tool answers clients through what looks like three "domains":

1. **Reply WhatsApp** — `internal/wa` + `internal/mcp` + `internal/dom` + the `pending`/`triage`/
   `keepalive` commands. Driven by the opencode session (the operator), **human-gated** (ToS/ban
   risk; AGENTS.md §0 confirm-before-send).
2. **Reply general questions** — the **personas** the cars-app chat AI uses
   (`chatbot/agent/chatbot.md`, `chatbot/agent/chatbot-editor.md`), selected by group permissions.
3. **Reply cars-app chat** — `internal/chatbot` + `internal/carshttp/chat.go` + the `chat-poll`
   daemon (single-writer lease, autonomous posting).

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
- `BuildPrompt` (`internal/chatbot/prompt.go`) injects it via `sharedRules()` (read once, `sync.Once`;
  falls back to a short safe stub if the file is unreadable).
- `ensureConfig` (`internal/chatbot/llm.go`) splices it into **both** personas at the
  `<!-- shared-rules -->` marker, so the personas carry only their differences.
- Drift guards: `TestSharedRulesSourceOfTruth`, and the injection assertions in
  `TestEnsureConfigInstallsAgents`.

Result: a rule/flag change is **one edit**.

## Phase 2 — name the seams (clarity, mechanical)

Namespace packages so the boundary is visible:

- `internal/channel/whatsapp` ← `wa` + `mcp` + `dom`
- `internal/channel/carsapp` ← `carshttp/chat.go` + `chatbot/*`
- `internal/core/…` ← `quote`, `webfetch`, `db/{faq,pricing,procurement}`

Low runtime value, mostly churn — do it only where it documents intent.

## Phase 3 — channel contract (scalability)

```go
type Channel interface {
    Name() string
    Poll(ctx) ([]Inbound, error)
    Send(ctx, Outbound) error
}
```

A shared **Responder** sits between: `Inbound → prompt(persona, channel) → Generate → Finalize →
channel.Send`. cars-app already fits (chat-poll is read → think → post).

**Autonomy differs per channel**: cars-app is a daemon; WhatsApp stays human-gated. The interface
exposes read/send with **autonomy as a per-channel policy** — do not force WhatsApp into an
autonomous loop.
