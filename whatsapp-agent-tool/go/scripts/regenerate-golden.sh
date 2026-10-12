#!/usr/bin/env bash
# Regenerate the Go golden fixtures from the (still-reference) Node tooling.
#
#   ./go/scripts/regenerate-golden.sh        # from the repo root
#
# Golden outputs are gitignored on purpose: they embed per-account client data
# (chat phone numbers in pending_state.json). Tests read them from
# go/testdata/golden/ — which every dev machine creates locally.
#
# Requires a working `ai` MySQL (quote/faq/stock) and the local .env.
set -euo pipefail
cd "$(dirname "$0")/../.."

G=go/testdata/golden
mkdir -p "$G"

node --import tsx src/quote.ts --list > "$G/quote_list.json" 2>"$G/quote_list.stderr"
for m in wholesale-fob wholesale-cfr retail-dza retail-usd retail-usd-fob allin-dza clearance; do
  node --import tsx src/quote.ts --car "COOLRAY BATTLE" --mode "$m" \
    > "$G/quote_coolray_${m}.json" 2>"$G/quote_coolray_${m}.stderr"
done
node --import tsx src/faq.ts list > "$G/faq_list.json" 2>"$G/faq_list.stderr"
node --import tsx src/pending.ts --state > "$G/pending_state.json" 2>"$G/pending_state.stderr"
node --import tsx src/identity-cli.ts > "$G/identity.json" 2>"$G/identity.stderr"
node --import tsx src/profiles-cli.ts list > "$G/profiles.json" 2>"$G/profiles.stderr"
node --import tsx src/stock.ts -- "LIVAN" > "$G/stock_livan.json" 2>"$G/stock_livan.stderr"

echo "golden fixtures regenerated -> $G"