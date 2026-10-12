# RULES.md — Go development rules (whatsapp-agent-tool)

Non-negotiable conventions for every Go file in this module. Read before writing or refactoring
Go code. These exist so the code stays modular, reviewable and cheap for an AI to edit.

## 1. Maximum modularity

- One concern per file. Name the file after the concern (`scope.go`, `client.go`, `stock.go`), not
  after the package.
- One responsibility per function; functions short enough to read top-to-bottom.
- Split by boundaries, not size: HTTP (pure I/O) is separate from decision logic, which is
  separate from CLI presentation. Business rules never live in the HTTP layer.
- Keep data flow explicit (see DEVELOPMENT.md §2). No hidden globals beyond the package-level
  `config.C` that already exists.

## 2. No file over 1000 lines

- Hard cap: **1000 lines per file**. The cap is a limit, not a target — most files should be well
  under it.
- When a file approaches the cap, extract the next natural unit into its own file in the same
  package. A large file is a smell before it hits 1000.
- A new function that brings a file past the cap must go into a new file.

## 3. AI-friendly

- Package and function comments state the *why* and the invariants, not the mechanics.
- Name things so a reader (human or model) does not need to jump to see what a value means:
  avoid single-letter/abbreviations except genuinely conventional ones (`b` for builder,
  `i` in a range).
- Keep a function's contract in its signature: return typed results and explicit errors rather
  than empty values + a `bool` gold mine.
- Prefer small pure functions with table tests over large integration tests.
- Mirror-external APIs verbatim: the wire formats of the cars app and the `ai` DB are the source
  of truth; our structs/tags replicate them byte-for-byte (see `carshttp.apiResponse`, `db/pricing`).

## 4. Max scalability

- Never hand the app SQL at runtime; fixed, parameterised queries only. Filters are bound, never
  concatenated.
- Keep column/row payloads minimal: fetch only what the consumer needs, and when permissions
  govern columns, strip at read time (see `carshttp/scope.go`).
- Avoid per-row allocations where a slice/`json.RawMessage` pass-through does the job
  (`encoding/json.RawMessage` rows are forwarded untouched until a caller needs them changed).
- Prefer vectorisable/I/O-bound designs: one HTTP call or one SQL round-trip per operation, no
  per-item loops that re-query.
- Guard with the smallest privilege that works: read-only commands, restricted roles,
  OS-store secrets. Assume the AI runs with the least permission, and fail-closed when in doubt.

## 5. Verification

- `cd go && go build ./...` and `go vet ./...` must be clean before a change is considered done.
- Keep `go test ./...` green; add a test for any new transformation (pure functions first).