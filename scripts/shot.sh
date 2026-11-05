#!/bin/sh
# Headless screenshot for the spec screenshot pass: scripts/shot.sh URL OUT.png WIDTH [HEIGHT]
# A URL with a #fragment renders blank; shoot the page without it with a tall HEIGHT and crop.
set -eu
url=$1; out=$2; width=$3; height=${4:-2000}
timeout 120 chromium --headless=new --no-sandbox --disable-gpu --hide-scrollbars \
  --window-size="$width,$height" --virtual-time-budget=10000 --screenshot="$out" "$url" >/dev/null 2>&1
echo "$out $(stat -c %s "$out") bytes"
