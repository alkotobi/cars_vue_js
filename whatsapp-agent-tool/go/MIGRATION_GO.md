# MIGRATION_GO.md — migrate the whatsapp Node/TS tooling to pure Go

Status: **APPROVED** (user 2026-10-10/11). User said: "I approve u can start" and
authorized browser-layer testing against the **default profile (whatsapp-profile)**,
NOT `sofia`.

Goal: replace the Node/TS CLIs with one static Go binary (`CGO_ENABLED=0`,
`whatsapp-tool`) that reproduces every CLI + MCP tool to **byte-identical JSON**
output, so the solution can be shipped as a single binary without source.

Effort estimate: stages 1–4 ≈ 2–3 sessions (~60% of the LOC, zero browser risk);
full migration totals **7–12 sessions**; data layer first, browser last.

## Deliverables (per stage)

- **Stage 1 (test-authorization + golden fixtures + infra):** clean-check the
  migration repo; golden fixtures captured from the *Node* side into
  `go/testdata/golden/`; Go workspace + `go.mod` + `.gitignore` updated.
  **Test rule:** browser-layer smoke tests run read-only against the **default
  profile** only; a built-in send guard (non-interactive `WA_HEADLESS=1` mode)
  must exist in the Go port before opening any real browser tests. Never touch
  `sofia`.
- **Stage 2 (framework spike + driver choice):** spike `go-rod` (primary) vs
  `chromedp` (fallback); one device + profile ID; conclude on the winner (the
  Chrome/CDP knowledge gathered from Node's Chromium is directly reusable).
- **Stage 3 (data layer):** workers, quote-style calculators, MySQL access,
  keyring-backed credentials, HTTP client for the cars app, state files, plus a
  parameterised (golden-fixture) parity test. **Reproduce every loader/CLI to
  byte-identical JSON** (maps→ordered JSON structs; MySQL `parseTime=false`
  like mysql2 `dateStrings:true`; `ToIso()`; the full quote calculator). **This
  stage is done.**
- **Stage 4 (MCP skeleton + offline pieces):** `whatsapp-tool mcp` (stdio,
  `mark3labs/mcp-go`), the full 8-tool surface registered (is_logged_in,
  get_chats, open_chat, send_message, get_last_messages, mark_as_read,
  session_info, identity) wired to a `wa.Client` interface; the fully
  offline tool (`whatsapp_identity`) works now; every browser-backed tool
  (including `session_info`, which needs the CDP/state layer) returns a
  descriptive `Error: browser layer not implemented yet (stage 5)` until
  Stage 5. Plus the pricing write CLI (`price`), state/pending/triage
  offline parts, secrets, cache. Binary name `whatsapp-tool`; MCP merged as a
  subcommand (not a separate binary).
- **Stage 5 (browser parity):** reproduce the DOM/selection helpers
  (`internal/dom`), the WA connect lifecycle, raw element access, MCP tool
  parity (open-chat / send / read / mark-read), **unsend guard** proxy at the
  driver layer. **This stage is functionally done** (2026-10-11): `internal/chrome`
  (twin of `src/chrome.ts`, incl. CDP port/state/spawn/attach/lifecycle),
  `internal/dom` (pure DOM helpers + tests), and `internal/wa/browser.go`
  (go-rod v0.116.2 real driver, mutex-serialized, human scroll walker)
  replace the stub; `pending`/`mcp` in `main.go` use the real browser.
  Verified against the live keeper Chrome (default profile, port 9223):
  `is_logged_in`, `get_chats` screen + `newest` walker, `get_last_messages`,
  `session_info`, `identity` — **byte-identical to the Node MCP under a same-account
  side-by-side diff** (only transport-envelope key order + uptime-second differ).
  `pending --state` remains byte-identical. Live `open_chat`/`send`/`mark_as_read`
  paths ported but not exercised against real chats (per test rules).
- **Stage 5b (remaining CLIs) — done (2026-10-11):** `triage`
  (`cmd/whatsapp-tool/triage.go`, twin of `src/triage.ts`: `--no-read`, clamps,
  `data/triage.json`, compact summary) and `keepalive`
  (`cmd/whatsapp-tool/keepalive.go`: `start|status|stop`) ported and wired into
  `main.go`. `keepalive status` verified live against the keeper Chrome
  (shape matches Node). Node source then removed (backed up at
  `backups/node-code-backup-2026-10-11.tar.gz`); docs updated to the Go commands.
- **Stage 6 (legacy golem/chromium cleanup, pack):** strip Node/Chromium from
  the shipped artifact; single static binary builds for macOS + Windows;
  docs/setup (keyring, env) updated; `.env` embedded; back up target share
  plan with rotation of the live GitLab PAT first.

## Key decisions (locked)

- Binary name `whatsapp-tool`; module `whatsapp-agent-tool` in `go/`.
- Deps: `github.com/go-rod/rod` primary (chromedp fallback), `github.com/zalando/go-keyring`, `github.com/go-sql-driver/mysql`, `github.com/mark3labs/mcp-go`.
- Language: Go 1.27, darwin/arm64 (CGO_ENABLED=0), static, runnable from anywhere (rootDir walks up to `.env`).
- Data lives in the shared MySQL `ai` database + OS keyring, exactly as in the
  Node tool — never shipped inside the binary.
- Node's ESM module-load trap in `connect.ts` disappears in Go (env read at
  runtime instead of at module load).
- License: never reproduce proprietary source verbatim; behavioral parity only.

## Byte-parity rules (every JSON output MUST follow these)

- **Never call `json.Marshal`/`json.MarshalIndent` directly.** Go escapes `<`,
  `>` and `&` to `\u003c` / `\u003e` / `\u0026`; Node's `JSON.stringify` leaves
  them literal. Route every encode through `internal/jsonx` (`Marshal` /
  `MarshalIndent`, both `SetEscapeHTML(false)`, no trailing newline). This bit us
  on `unanswered_chats.json` (a `reason` containing `<3 years`). Note a custom
  `MarshalJSON` does **not** save you: the outer `json.Marshal` re-compacts the
  returned bytes with `escapeHTML=true`, so the fix must be on the top-level
  encoder — and inside `ordered.MarshalJSON` too.
- **Key order = Node insertion order.** Go maps sort keys; use the `ordered`
  type (or structs with fields declared in Node's order) wherever the golden
  files carry Node key order.
- **Pretty vs compact.** `quote` + `identity` + `session_info` print pretty
  (`JSON.stringify(x, null, 2)`); `faq`/`stock`/`profiles`/`price`/`pending`
  print compact (`JSON.stringify(x)`). State files (`statex.writeJson`),
  identity, profiles and the `pending.json` dump are pretty + trailing `\n`.
- **null vs omission.** Node emits `undefined` keys as absent and `null` for
  explicit nulls; model absent optional fields with `omitempty` and explicit
  nulls with a pointer / `any(nil)`.
- **Empty arrays.** Node emits `[]`, never `null`; coerce nil slices to `[]`.
- Known residual edge: Go escapes U+2028/U+2029 even with `SetEscapeHTML(false)`
  (Node does not). Not seen in this data; revisit only if a golden fails on it.

## Stage 5 lessons (2026-10-11)

- **Never let the state file lie twice.** `chrome.CdpPort()` now trusts an alive
  CDP port from the state file even when the recorded pid is stale (a keepalive
  Chrome can restart in place / be re-flagged on the same port; the free-port scan
  can never *discover* a live port, so dropping it strands the account).
- **Default case = "locked":** with a profile singleton-lock held by another
  **alive** pid and CDP unreachable on our port, `spawnAndAttach` must return a
  hard error and stop — NOT fall through to a legacy spawn (the spawn succeeds,
  then the orphan dies and leaves a corrupt `data/whatsapp-cdp.json`). Node throws
  here; so does Go now. `launchPersistent` also refuses a locked profile.
- **Wire parity traps:** Go maps reorder JSON keys ⇒ use a struct for
  `{mode, chats}` (mode first, like Node's `JSON.stringify({ mode, ...res })`); a
  `bool` with `omitempty` silently drops `"unread":false` — drop `omitempty` on
  booleans Node always emits. The mcp-go transport envelope orders JSON-RPC keys
  differently from the TS SDK but the **tool payload text** is what must match.
- **ID stability:** the tool identity (`vivid-sphinx-44b3b584ebaf`) is minted by
  code (`crypto`) per account and is stable across both servers — treat it as the
  assistant's real name in every future session.