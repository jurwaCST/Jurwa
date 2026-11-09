#!/usr/bin/env bash
# Builds every service image and pushes it to the k3d local registry
# (localhost:5050 from the laptop, jurwa-registry:5050 inside the cluster).
#   bash scripts/push-local-registry.sh                       all 20 services
#   bash scripts/push-local-registry.sh api-gateway identity  only these
#   bash scripts/push-local-registry.sh web                   the web app (web/Dockerfile)
# Then delete the pod to pull the new image (Argo CD reverts a rollout restart):
#   kubectl -n jurwa delete pod -l app.kubernetes.io/name=<name>
set -euo pipefail

cd "$(dirname "$0")/.."
REGISTRY="localhost:5050"

if [ "$#" -gt 0 ]; then
  services=("$@")
else
  services=()
  for dir in services/*/; do
    name="$(basename "$dir")"
    [ "$name" = "_template" ] && continue
    services+=("$name")
  done
fi

if ! curl -sf "http://$REGISTRY/v2/" >/dev/null; then
  echo "No registry on $REGISTRY. Create the cluster first: ./deploy/k3d/create-cluster.sh" >&2
  exit 1
fi

failed=()
for name in "${services[@]}"; do
  log="/tmp/jurwa-push-$name.log"
  if [ "$name" = "web" ]; then
    # Not a Go service: its own Dockerfile, with web/ as the build context.
    build=(docker build -f web/Dockerfile -t "jurwa/web" web)
  else
    build=(make docker-build SERVICE="$name")
  fi
  if "${build[@]}" >"$log" 2>&1 \
    && docker tag "jurwa/$name" "$REGISTRY/$name:latest" >>"$log" 2>&1 \
    && docker push "$REGISTRY/$name:latest" >>"$log" 2>&1; then
    echo "OK      $name"
  else
    echo "FAILED  $name (log: $log)"
    failed+=("$name")
  fi
done

echo "Pushed $((${#services[@]} - ${#failed[@]})) image(s), ${#failed[@]} failed."
[ "${#failed[@]}" -eq 0 ]
