#!/usr/bin/env bash
# Tests for scripts/check-coverage.sh. "go" is replaced by a fake on PATH, so
# the tests can make a package low, failing or without tests.
#   bash scripts/tests/check-coverage_test.sh
set -uo pipefail

cd "$(dirname "$0")/../.."
SCRIPT="scripts/check-coverage.sh"
FAKES="$(mktemp -d)"
trap 'rm -rf "$FAKES"' EXIT

# Fake go: "go test -cover <pkg>" prints one line like the real one.
# LOW_PKG gets 55.0%, FAIL_PKG fails, NOTEST_PKG has no test files.
cat > "$FAKES/go" <<'FAKE'
#!/usr/bin/env bash
pkg="${@: -1}"
full="github.com/jurwaCST/Jurwa/${pkg#./}"
case "$pkg" in
  *"${FAIL_PKG:-none}") echo "--- FAIL: TestSomething"; echo "FAIL	$full	0.1s"; exit 1 ;;
  *"${NOTEST_PKG:-none}") echo "?   	$full	[no test files]"; exit 0 ;;
  *"${LOW_PKG:-none}") echo "ok  	$full	0.1s	coverage: 55.0% of statements"; exit 0 ;;
  *) echo "ok  	$full	0.1s	coverage: 92.5% of statements"; exit 0 ;;
esac
FAKE
chmod +x "$FAKES/go"

failures=0
check() { # description, condition result
  if [ "$2" = "0" ]; then echo "PASS  $1"; else echo "FAIL  $1"; failures=$((failures + 1)); fi
}
run() { PATH="$FAKES:$PATH" bash "$SCRIPT" "$@" 2>&1; }

# 1. All packages at or above 70%
out="$(run)"; code=$?
check "all pass: exit code 0" "$([ $code -eq 0 ]; echo $?)"
check "all pass: one line per critical package (10)" "$([ "$(echo "$out" | grep -cE ' (OK|LOW|FAIL|NO TESTS)$')" -eq 10 ]; echo $?)"
check "all pass: routing-rules/domain listed with its coverage" "$(echo "$out" | grep -qE '^services/routing-rules/domain +92\.5% +OK$'; echo $?)"
check "all pass: status-lifecycle/application listed" "$(echo "$out" | grep -qE '^services/status-lifecycle/application +92\.5% +OK$'; echo $?)"
check "all pass: summary line" "$(echo "$out" | grep -q '^10 of 10 critical packages at or above 70%$'; echo $?)"

# 2. One package under 70%
out="$(LOW_PKG=services/sla-policy/application run)"; code=$?
check "low: exit code 1" "$([ $code -eq 1 ]; echo $?)"
check "low: sla-policy/application is LOW" "$(echo "$out" | grep -qE '^services/sla-policy/application +55\.0% +LOW$'; echo $?)"
check "low: summary names 9 of 10" "$(echo "$out" | grep -q '^9 of 10 critical packages at or above 70%$'; echo $?)"

# 3. Failing tests count as a failure, not as 0%
out="$(FAIL_PKG=services/routing-engine/domain run)"; code=$?
check "failing tests: exit code 1" "$([ $code -eq 1 ]; echo $?)"
check "failing tests: routing-engine/domain is FAIL" "$(echo "$out" | grep -qE '^services/routing-engine/domain +- +FAIL$'; echo $?)"

# 4. A package without tests fails the gate
out="$(NOTEST_PKG=services/sla-tracking/domain run)"; code=$?
check "no tests: exit code 1" "$([ $code -eq 1 ]; echo $?)"
check "no tests: sla-tracking/domain is NO TESTS" "$(echo "$out" | grep -qE '^services/sla-tracking/domain +- +NO TESTS$'; echo $?)"

# 5. The threshold can be changed
out="$(MIN_COVERAGE=95 run)"; code=$?
check "MIN_COVERAGE=95: 92.5% is LOW, exit code 1" "$([ $code -eq 1 ] && echo "$out" | grep -q 'LOW$'; echo $?)"

echo
if [ "$failures" -eq 0 ]; then echo "All checks passed."; else echo "$failures check(s) failed."; fi
[ "$failures" -eq 0 ]
