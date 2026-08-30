#!/usr/bin/env bash
set -euo pipefail

STAGING_URL="${1:-http://localhost:8090}"
PROD_URL="${2:-}"
PASS=0
FAIL=0

check() {
  local desc="$1" url="$2" expect="$3"
  local resp
  resp=$(curl -sf --max-time 10 "$url" 2>/dev/null) || {
    echo "FAIL  $desc — $url — connection failed"
    FAIL=$((FAIL + 1))
    return
  }
  if grep -qi "$expect" <<< "$resp"; then
    echo "PASS  $desc"
    PASS=$((PASS + 1))
  else
    echo "FAIL  $desc — expected '$expect' in response"
    echo "      got: ${resp:0:200}"
    FAIL=$((FAIL + 1))
  fi
}

run_checks() {
  local base="$1" label="$2"
  echo "Smoke tests ($label): $base"
  echo "---"
  check "Health endpoint"       "$base/api/health"       '"status":"ok"'
  check "Bills list has data"   "$base/api/bills"        '"id":'
  check "Frontend serves HTML"  "$base/"                 '<!doctype html>'
  echo ""
}

echo "=== MIACTION Smoke Tests ==="
echo ""

run_checks "$STAGING_URL" "staging"

if [ -n "$PROD_URL" ]; then
  run_checks "$PROD_URL" "production"
fi

echo "---"
echo "$PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
