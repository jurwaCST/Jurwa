#!/usr/bin/env bash
# Tests for scripts/check-all-health.sh. grpcurl and kubectl are replaced by
# fakes on PATH, so nothing needs to run.
#   bash scripts/tests/check-all-health_test.sh
set -uo pipefail

cd "$(dirname "$0")/../.."
SCRIPT="scripts/check-all-health.sh"
FAKES="$(mktemp -d)"
trap 'rm -rf "$FAKES"' EXIT

# Fake grpcurl: SERVING, except on the port in DOWN_PORT (connection refused).
cat > "$FAKES/grpcurl" <<'FAKE'
#!/usr/bin/env bash
for a in "$@"; do case "$a" in *:9*) port="${a##*:}" ;; esac; done
if [ "${port:-}" = "${DOWN_PORT:-none}" ]; then
  echo "Failed to dial target host \"localhost:$port\": connection refused" >&2
  exit 1
fi
printf '{\n  "status": "SERVING"\n}\n'
FAKE

# Fake kubectl: every deployment 1/1, except DOWN_DEPLOY (0/1) and ZERO_DEPLOY (0/0).
cat > "$FAKES/kubectl" <<'FAKE'
#!/usr/bin/env bash
for name in report-submission photo-media report-history location-taxonomy internal-notes \
  routing-rules routing-engine department-directory category-taxonomy api-gateway \
  status-lifecycle sla-tracking sla-policy reassignment audit-log identity authorization \
  notification feedback analytics; do
  ready=1; want=1
  [ "$name" = "${DOWN_DEPLOY:-none}" ] && ready=0
  [ "$name" = "${ZERO_DEPLOY:-none}" ] && { ready=0; want=0; }
  printf '%s %s %s\n' "$name" "$want" "$ready"
done
FAKE
chmod +x "$FAKES/grpcurl" "$FAKES/kubectl"

failures=0
check() { # description, condition result
  if [ "$2" = "0" ]; then echo "PASS  $1"; else echo "FAIL  $1"; failures=$((failures + 1)); fi
}

run() { PATH="$FAKES:$PATH" bash "$SCRIPT" "$@" 2>&1; }

# 1. All up
out="$(run)"; code=$?
check "all up: exit code 0" "$([ $code -eq 0 ]; echo $?)"
check "all up: 20 service lines" "$([ "$(echo "$out" | grep -cE ' (UP|DOWN)$')" -eq 20 ]; echo $?)"
check "all up: total line says 20 up, 0 down" "$(echo "$out" | grep -q '^20 up, 0 down$'; echo $?)"
check "all up: report-submission on 9001 is UP" "$(echo "$out" | grep -qE '^report-submission +9001 +UP$'; echo $?)"
check "all up: analytics on 9020 is UP" "$(echo "$out" | grep -qE '^analytics +9020 +UP$'; echo $?)"

# 2. One service down
out="$(DOWN_PORT=9019 run)"; code=$?
check "one down: exit code 1" "$([ $code -eq 1 ]; echo $?)"
check "one down: feedback on 9019 is DOWN" "$(echo "$out" | grep -qE '^feedback +9019 +DOWN$'; echo $?)"
check "one down: total line says 19 up, 1 down" "$(echo "$out" | grep -q '^19 up, 1 down$'; echo $?)"

# 3. Kubernetes mode, all ready
out="$(run --k8s)"; code=$?
check "k8s all ready: exit code 0" "$([ $code -eq 0 ]; echo $?)"
check "k8s all ready: 20 lines with ready/desired" "$([ "$(echo "$out" | grep -cE ' 1/1 +UP$')" -eq 20 ]; echo $?)"

# 4. Kubernetes mode, one pod not ready and one scaled to zero
out="$(DOWN_DEPLOY=identity ZERO_DEPLOY=notification run --k8s)"; code=$?
check "k8s: exit code 1" "$([ $code -eq 1 ]; echo $?)"
check "k8s: identity 0/1 is DOWN" "$(echo "$out" | grep -qE '^identity +0/1 +DOWN$'; echo $?)"
check "k8s: notification scaled to 0/0 is DOWN" "$(echo "$out" | grep -qE '^notification +0/0 +DOWN$'; echo $?)"
check "k8s: total line says 18 up, 2 down" "$(echo "$out" | grep -q '^18 up, 2 down$'; echo $?)"

# 5. Unknown option
out="$(run --nope)"; code=$?
check "unknown option: exit code 2" "$([ $code -eq 2 ]; echo $?)"

echo
if [ "$failures" -eq 0 ]; then echo "All checks passed."; else echo "$failures check(s) failed."; fi
[ "$failures" -eq 0 ]
