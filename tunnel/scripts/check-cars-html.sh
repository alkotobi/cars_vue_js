#!/usr/bin/env bash
# Check that https://cars.merhab.com/ returns the expected Hello page HTML.
# Usage: ./scripts/check-cars-html.sh

set -e
URL="${1:-https://cars.merhab.com/}"

echo "Fetching $URL ..."
HTML=$(curl -sk --connect-timeout 15 --max-time 30 "$URL" 2>/dev/null) || {
  echo "Failed to fetch $URL (timeout or connection error)"
  exit 1
}

# Key strings that must appear in the expected page
checks=(
  "Hello | merhab.com"
  "Welcome to merhab.com"
  "Secure & Beautiful"
  "HTTPS Works!"
  "hello-container"
)

missing=0
for s in "${checks[@]}"; do
  if echo "$HTML" | grep -qF "$s"; then
    echo "  OK: found \"$s\""
  else
    echo "  MISSING: \"$s\""
    missing=1
  fi
done

if [[ $missing -eq 1 ]]; then
  echo ""
  echo "Result: Page content does NOT match the expected Hello page."
  exit 1
fi

echo ""
echo "Result: cars.merhab.com is serving the expected Hello page."
exit 0
