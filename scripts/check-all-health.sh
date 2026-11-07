#!/usr/bin/env bash
# Shows which of the 20 Jurwa services are up or down, from one place.
#
#   bash scripts/check-all-health.sh          gRPC health check on localhost:<port>
#   HOST=10.0.0.5 bash scripts/check-all-health.sh
#   bash scripts/check-all-health.sh --k8s    Kubernetes: ready pods against wanted pods
#
# One line per service (name, port or ready/wanted, UP or DOWN), then a total.
# Exit code: 0 when all are up, 1 when any is down, 2 on a usage error.
set -uo pipefail

cd "$(dirname "$0")/.."
HOST="${HOST:-localhost}"
NAMESPACE="${NAMESPACE:-jurwa}"
TIMEOUT=2 # seconds per service

# name and gRPC port (services/_template/README.md)
SERVICES="
report-submission 9001
photo-media 9002
report-history 9003
location-taxonomy 9004
internal-notes 9005
routing-rules 9006
routing-engine 9007
department-directory 9008
category-taxonomy 9009
api-gateway 9010
status-lifecycle 9011
sla-tracking 9012
sla-policy 9013
reassignment 9014
audit-log 9015
identity 9016
authorization 9017
notification 9018
feedback 9019
analytics 9020
"

up=0
down=0

report() { # name, detail, ok (0 = up)
  if [ "$3" -eq 0 ]; then
    printf '%-22s %-6s UP\n' "$1" "$2"
    up=$((up + 1))
  else
    printf '%-22s %-6s DOWN\n' "$1" "$2"
    down=$((down + 1))
  fi
}

check_grpc() {
  while read -r name port; do
    [ -z "$name" ] && continue
    # The proto file is given, so a service without reflection works too.
    out="$(grpcurl -plaintext -max-time "$TIMEOUT" -import-path deploy/k3d -proto health.proto \
      "$HOST:$port" grpc.health.v1.Health/Check 2>/dev/null)"
    echo "$out" | grep -q '"SERVING"'
    report "$name" "$port" $?
  done <<< "$SERVICES"
}

check_k8s() {
  if ! deploys="$(kubectl -n "$NAMESPACE" get deploy \
    -o jsonpath='{range .items[*]}{.metadata.name} {.spec.replicas} {.status.readyReplicas}{"\n"}{end}' 2>/dev/null)"; then
    echo "kubectl failed. Is the cluster running? k3d cluster list" >&2
    exit 2
  fi
  while read -r name _; do
    [ -z "$name" ] && continue
    line="$(echo "$deploys" | awk -v n="$name" '$1 == n')"
    if [ -z "$line" ]; then
      report "$name" "none" 1
      continue
    fi
    want="$(echo "$line" | awk '{print $2}')"
    ready="$(echo "$line" | awk '{print ($3 == "" ? 0 : $3)}')"
    # Scaled to zero counts as down: nothing answers.
    [ "$want" -gt 0 ] && [ "$ready" -ge "$want" ]
    report "$name" "$ready/$want" $?
  done <<< "$SERVICES"
}

case "${1:-}" in
  "") check_grpc ;;
  --k8s) check_k8s ;;
  -h | --help) sed -n '2,10p' "$0"; exit 0 ;;
  *) echo "Unknown option: $1 (use --k8s or nothing)" >&2; exit 2 ;;
esac

echo "$up up, $down down"
[ "$down" -eq 0 ]
