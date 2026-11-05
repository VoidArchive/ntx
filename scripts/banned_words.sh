#!/bin/sh
# Recommending words have no place in copy the story engine renders (CODING_STANDARDS.md).
# Checks Go string literals outside tests and import blocks; the Svelte side is an eslint rule.
set -eu
cd "$(dirname "$0")/.."
hits=$(find apps/api/internal -name '*.go' ! -name '*_test.go' -print0 \
  | xargs -0 grep -nE '"[^"]*\b([Bb]uy|[Ss]ell|[Bb]ullish|[Bb]earish|[Tt]arget|[Ss]ignal|[Ss]core)\b[^"]*"' \
  | grep -vE '^[^:]+:[0-9]+:[[:space:]]*"' || true)
if [ -n "$hits" ]; then
  echo "$hits"
  echo "banned word in a Go string literal" >&2
  exit 1
fi
