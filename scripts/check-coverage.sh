#!/usr/bin/env bash
# Coverage quality gate for the critical code (proposal objective 4: at least 70%
# on routing, SLA and status-flow code). Runs in CI (.github/workflows/quality-gate.yml).
#
#   bash scripts/check-coverage.sh
#   MIN_COVERAGE=80 bash scripts/check-coverage.sh
#
# One line per package: path, coverage, OK / LOW / FAIL (tests fail) / NO TESTS.
# Exit code 0 when every package passes, 1 otherwise.
set -uo pipefail

cd "$(dirname "$0")/.."
MIN="${MIN_COVERAGE:-70}"

# domain and application of the routing, SLA and status-flow services.
PACKAGES="
services/routing-rules/domain
services/routing-rules/application
services/routing-engine/domain
services/routing-engine/application
services/sla-policy/domain
services/sla-policy/application
services/sla-tracking/domain
services/sla-tracking/application
services/status-lifecycle/domain
services/status-lifecycle/application
"

total=0
passed=0
for pkg in $PACKAGES; do
  total=$((total + 1))
  out="$(go test -count=1 -cover "./$pkg" 2>&1)"
  code=$?
  pct="$(echo "$out" | grep -oE 'coverage: [0-9.]+%' | tail -1 | grep -oE '[0-9.]+')"

  if [ "$code" -ne 0 ]; then
    result="FAIL"; shown="-"
  elif echo "$out" | grep -q 'no test files'; then
    result="NO TESTS"; shown="-"
  elif [ -z "$pct" ]; then
    result="NO TESTS"; shown="-"
  elif awk -v p="$pct" -v m="$MIN" 'BEGIN { exit !(p >= m) }'; then
    result="OK"; shown="$pct%"
    passed=$((passed + 1))
  else
    result="LOW"; shown="$pct%"
  fi
  printf '%-40s %-7s %s\n' "$pkg" "$shown" "$result"
  if [ "$result" = "FAIL" ]; then
    echo "$out" | grep -E -- '--- FAIL|FAIL' | head -5 | sed 's/^/    /'
  fi
done

echo "$passed of $total critical packages at or above $MIN%"
[ "$passed" -eq "$total" ]
