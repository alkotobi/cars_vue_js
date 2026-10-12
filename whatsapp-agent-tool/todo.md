# TODO — Token & Speed, round 2

> **Historical (Node era).** This is a work log from before the Go rewrite. The `npm run …`
> commands it mentions are now `whatsapp-tool …` (see [`go/MIGRATION_GO.md`](./go/MIGRATION_GO.md)).

Written from an audit on 2026-10-08 (round 1 — AGENTS/DEVELOPMENT split, archiving, batch
scripts — is done and is summarized at the bottom). Ranked by measured impact. Bold = done.

## Where the tokens go (measured)

| Per turn | ~Tokens |
|---|---|
| `AGENTS.md` (fixed, cached by most providers) | 6,400 |
| `ai_questions.json` + `unanswered_chats.json` (re-read **every** turn) | 6,640 |
| `ALGERIAN_DERJA.md` + `CANNED_REPLIES.md` (reply turns) | 6,915 |
| §13 pricing prose inside AGENTS.md (fixed) | 1,950 |
| `cars.json` + `custom.json` (pricing turns) | 2,745 |

A reply turn ≈ 20–25k input tokens before message history. The recurring waste is the raw state
files, not the system prompt.

## 1. `npm run pending` report must be self-sufficient ✅
- [x] `src/pending.ts` report now carries, per `data_ready` chat: `chatKey`/`chatName`, `needed`,
  `reason`, and the full linked questions `{id, question, answer, answeredBy, askedAt}`.
- [x] Full dump (incl. client `lastMessage`) written to `data/pending.json` for the rare case it
  is needed.
- [x] `AGENTS.md` §14b: **do not Read the raw state files** — the report is the answer. Kills
  the 6.6k tokens/turn of re-reading `ai_questions.json` + `unanswered_chats.json`.
- [x] Still prints no client message bodies to stdout (§0 rule 5 respected).

## 2. `npm run quote` — deterministic pricing ✅
- [x] New `src/quote.ts`: reads `cars/benefits/fright/rate/custom.json`, applies the §13
  formulas **in code**, prints one JSON result (car match, inputs, formula, breakdown, price,
  warnings).
- [x] Modes: `wholesale-fob`, `wholesale-cfr`, `retail-dza`, `retail-usd`, `retail-usd-fob`,
  `allin-dza`, `clearance`.
- [x] Enforces freshness (7-day rule) → `status: "STALE"` with the exact file + age; missing
  data → `status: "MISSING"`. Expired `validTo` freight is dropped automatically.
- [x] Alias rules in code: `FIVAN`→`LIVAN`, `MAN`/`AUTO` are different cars, ambiguous name →
  `status: "AMBIGUOUS"` with candidates (never guesses).
- [x] Replaces mental arithmetic — this is the fix for the 12,674 vs 11,300 FOB/CFR class of
  mistake, which happened *with* §13 fully in context.

## 3. AGENTS.md §13 → point at the calculator ✅
- [x] §13 shrunk from 7.8 KB to ~2.6 KB: inquiry flow + freshness rules + "run `npm run quote`"
  + all-in/customs/TVN notes. Field-by-field data-file schemas and write rules moved to
  [`PRICING.md`](./PRICING.md) (loaded **only** when the boss gives new numbers or the script
  reports STALE/MISSING).

## 4. AGENTS.md §14 → `STATE.md` on demand ✅
- [x] Entry schemas + creation/dedup rules moved to [`STATE.md`](./STATE.md); AGENTS.md keeps
  the every-turn workflow, the 14b scan and the 14c group template.
- [x] Trigger kept in AGENTS.md: read `STATE.md` **before creating/updating** a state entry.

## 5. Glossary dead weight → archive ✅
- [x] `ALGERIAN_DERJA.md` §5b (superseded by §3), §6 (open questions), §7 (change log) and
  `CANNED_REPLIES.md` change log moved to [`ALGERIAN_DERJA_ARCHIVE.md`](./ALGERIAN_DERJA_ARCHIVE.md).
- [x] Nothing deleted — archived, with pointers from the live files. Measured: 17.8 KB → 14.7 KB
  (~0.8k tokens/reply turn back, a bit less than hoped — §1 business context stays and should).

## 6. Deduplicate AGENTS.md ✅
- [x] The AI-disclosure rule is stated **once** (§0.8, with the CANNED_REPLIES vs clear-Arabic
  variants) and referenced as "the AI disclosure line" everywhere else.
- [x] Stale-data rule, "ask, never guess" and the holding reply each stated once.
- [x] `AGENTS.md` 25.8 KB → **14.0 KB (~6.4k → ~3.5k tokens)**, no rule dropped.

## 7. Merge the three chat-listing tools (9 → 7) ✅
- [x] `whatsapp_get_chats(limit?, mode?: "screen" | "newest" | "unread", timeoutMs?)` replaces
  `whatsapp_get_chats` / `whatsapp_get_last_chats` / `whatsapp_get_last_unread_chats`.
- [x] `DEVELOPMENT.md` §4, `README.md` and `AGENTS.md` playbooks updated.
- [x] Verified against the live server: stdin/stdio probe lists exactly the 7 tools.
- [x] `is_logged_in` and `session_info` stay separate on purpose (one boots Chrome, one must not).

## 8. Speed — no code, doc only ✅ / deferred
- [x] "Daily digest" playbook now points at `npm run triage -- N 5 --no-read` (1 turn instead
  of 3N). Manual chat-by-chat pass kept only as the escape hatch.
- [x] Start-of-turn already mandates `npm run pending`; `npm run keepalive` at day start stays
  the fix for the 10–25 s first-call boot.
- [ ] Scripts from `dist/` instead of `tsx` — **deferred**: `tsc` full build (~1–2 s) is slower
  than `tsx` startup (~0.4 s) and a stale `dist/` would silently run old logic. Not worth it.
- [ ] `openChat` latency tuning — **deferred on purpose**: the 6 s poll + 400 ms debounce are
  defensive code for the wrong-chat-send bug. Saving 300–600 ms is not worth that regression.
- [ ] Batch-send for `data_ready` answers — **deferred again**: sends to real clients stay
  human-confirmed (§0.1). The guards live in `src/autoreply.ts send` if ever needed.

## Verification
- [x] `npx tsc --noEmit` clean, `npm run build` clean
- [x] MCP stdio probe: server starts, exactly 7 tools registered
- [x] `npm run smoke` end-to-end (login state + chats + last messages + direction)
- [x] `npm run quote` live: retail-dza / wholesale-fob / wholesale-cfr --line / allin-dza /
  clearance, plus STALE (rate backdated), MISSING (unknown car), AMBIGUOUS (LIVAN), FIVAN→LIVAN
  alias, `--list`
- [x] `npm run pending -- --state` report: `dataReady` carries the boss's answers inline,
  `waiting`/`toSendToGroup` are self-contained; ~3.8 KB (~855 tokens) replaces reading the two
  hot state files (~6.6k tokens). `data/pending.json` written with the full dump.

## Round-1 leftovers & follow-ups
- `data_ready` chat `+213 671 41 56 57` (Q17 — transit/delivery time): the stored group "answer"
  is the AR/EN/ZH **re-post of the question**, not an answer. Verify with the client/boss before
  replying (§14b step 2 is exactly this check).

---

## Round 1 (done, 2026-10-08)
AGENTS/DEVELOPMENT split (11.6k → 6.8k tokens), `pending`/`triage` batch scripts, state
archiving (`trimAndArchive`/`promoteReady`), compact tool outputs, lower defaults, narrowed
glossary rule, `openChat` fast-fail, `restoreAllFilter`.
