#!/usr/bin/env bash
# Starts what the Cypress tests need (SCRUM-108): api-gateway and six services,
# in memory, with the demo accounts seeded. Git Bash, from the repo root:
#
#   export SEED_DEMO_PASSWORD='<any 8+ characters, never commit it>'
#   bash web/cypress/start-stack.sh          # Ctrl+C stops everything
#
# Ports can be moved when one is taken, for example GATEWAY_HTTP_PORT=18080.
set -euo pipefail

: "${SEED_DEMO_PASSWORD:?set SEED_DEMO_PASSWORD first, the password of the demo accounts}"
GATEWAY_HTTP_PORT="${GATEWAY_HTTP_PORT:-8080}"
IDENTITY_PORT="${IDENTITY_PORT:-9016}"
AUTHORIZATION_PORT="${AUTHORIZATION_PORT:-9017}"
CATEGORY_PORT="${CATEGORY_PORT:-9009}"

bin="$(mktemp -d)"
ext="$(go env GOEXE)" # ".exe" on Windows
logs="web/cypress/logs"
mkdir -p "$logs"
echo "Building into $bin ..."
for s in api-gateway identity authorization category-taxonomy location-taxonomy report-submission internal-notes; do
  go build -o "$bin/$s$ext" "./services/$s/cmd/server"
done

# A new signing secret on every start: tokens from an earlier run stop working.
JWT_SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
export JWT_SECRET SEED_DEMO_PASSWORD

pids=()
start() { # name, then VAR=value pairs
  local name="$1"; shift
  env "$@" "$bin/$name$ext" > "$logs/$name.log" 2>&1 &
  pids+=("$!")
}
trap 'kill "${pids[@]}" 2>/dev/null; rm -rf "$bin"' EXIT
trap 'exit 130' INT TERM # Ctrl+C runs the EXIT trap above

start identity PORT="$IDENTITY_PORT" SEED_DEMO_USERS=true
start authorization PORT="$AUTHORIZATION_PORT"
start category-taxonomy PORT="$CATEGORY_PORT"
start location-taxonomy
start report-submission
start internal-notes
sleep 2
start api-gateway HTTP_PORT="$GATEWAY_HTTP_PORT" \
  IDENTITY_ADDR="localhost:$IDENTITY_PORT" \
  AUTHORIZATION_ADDR="localhost:$AUTHORIZATION_PORT" \
  CATEGORY_TAXONOMY_ADDR="localhost:$CATEGORY_PORT"

for i in $(seq 1 30); do
  if curl -fs "http://localhost:$GATEWAY_HTTP_PORT/health" > /dev/null; then
    echo "Ready: gateway on http://localhost:$GATEWAY_HTTP_PORT, logs in $logs/. Ctrl+C stops it."
    wait
  fi
  sleep 1
done
echo "The gateway did not answer on port $GATEWAY_HTTP_PORT. See $logs/api-gateway.log"
exit 1
