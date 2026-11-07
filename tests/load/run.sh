#!/usr/bin/env bash
# SCRUM-109: load test of POST /reports through the api-gateway with Apache JMeter.
#   bash tests/load/run.sh                 50 users, 2 minutes (the agreed target)
#   USERS=5 DURATION=10 bash tests/load/run.sh   a quick smoke run
# Needs identity, authorization, report-submission and api-gateway running (see
# tests/load/README.md). Results go to tests/load/out/<time>/ (ignored by git):
# results.jtl, the JMeter HTML report in html/, and summary.txt.
set -euo pipefail

cd "$(dirname "$0")"
GATEWAY="${GATEWAY:-http://localhost:8080}"
USERS="${USERS:-50}"
DURATION="${DURATION:-120}"
RAMPUP="${RAMPUP:-10}"
PER_USER_PER_MINUTE="${PER_USER_PER_MINUTE:-60}"
OUT="out/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUT"

command -v jmeter >/dev/null || { echo "jmeter not found: brew install jmeter" >&2; exit 1; }
curl -sf "$GATEWAY/health" >/dev/null || { echo "api-gateway is not answering on $GATEWAY" >&2; exit 1; }

# One Student per virtual user, made one at a time before the timed part, so
# sign-up (bcrypt) is not part of the measurement. The tokens stay in $OUT only.
echo "Signing up $USERS load-test students..."
stamp="$(date +%s)"
: >"$OUT/tokens.csv"
for i in $(seq 1 "$USERS"); do
  email="load-$stamp-$i@rub.edu.bt"
  curl -sf -X POST "$GATEWAY/auth/register" \
    -d "{\"email\":\"$email\",\"fullName\":\"Load $i\",\"password\":\"load-test-pass\"}" >/dev/null
  curl -sf -X POST "$GATEWAY/auth/login" -d "{\"email\":\"$email\",\"password\":\"load-test-pass\"}" |
    python3 -c 'import sys, json; print(json.load(sys.stdin)["accessToken"])' >>"$OUT/tokens.csv"
done

host="${GATEWAY#http://}"
echo "Running JMeter: $USERS users, ${DURATION}s, ramp-up ${RAMPUP}s, $PER_USER_PER_MINUTE reports a minute per user..."
jmeter -n -t report-submission.jmx \
  -Jhost="${host%%:*}" -Jport="${host##*:}" \
  -Jusers="$USERS" -Jduration="$DURATION" -Jrampup="$RAMPUP" \
  -JperUserPerMinute="$PER_USER_PER_MINUTE" -Jtokens="$OUT/tokens.csv" \
  -l "$OUT/results.jtl" -j "$OUT/jmeter.log" -e -o "$OUT/html" | grep -E "^summary ="

python3 summarize.py "$OUT/results.jtl" | tee "$OUT/summary.txt"
rm -f "$OUT/tokens.csv"
echo "HTML report: tests/load/$OUT/html/index.html"
