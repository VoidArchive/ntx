#!/bin/sh
# Recommending words have no place in copy the story engine renders (CODING_STANDARDS.md).
# Checks Go string literals outside tests; the Svelte side is an eslint rule.
set -eu
cd "$(dirname "$0")/.."
# An import path is a lone quoted token with no spaces; a sentence is not.
hits=$(find apps/api/internal -name '*.go' ! -name '*_test.go' -print0 \
  | xargs -0 grep -inE '"[^"]*\b(buy|sell|bullish|bearish|target|signal|score)\b[^"]*"' \
  | grep -vE '^[^:]+:[0-9]+:[[:space:]]*([A-Za-z_]+[[:space:]]+)?"[^" ]+"[[:space:]]*$' || true)
if [ -n "$hits" ]; then
  echo "$hits"
  echo "banned word in a Go string literal" >&2
  exit 1
fi
