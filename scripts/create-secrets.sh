#!/usr/bin/env bash
# Creates (or updates) the Kubernetes Secret "jurwa-secrets" in namespace jurwa
# from dev/.env. The values never go into Git.
#   cp dev/.env.example dev/.env   (then change the passwords)
#   bash scripts/create-secrets.sh
# JWT_SECRET: taken from dev/.env if set, otherwise the value already in the
# cluster is kept, otherwise a random one is made.
set -euo pipefail

cd "$(dirname "$0")/.."
ENV_FILE="dev/.env"
NS="jurwa"

if [ ! -f "$ENV_FILE" ]; then
  echo "Missing $ENV_FILE. Run: cp dev/.env.example dev/.env" >&2
  exit 1
fi
set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

for v in POSTGRES_USER POSTGRES_PASSWORD MONGO_USER MONGO_PASSWORD REDIS_PASSWORD; do
  value="${!v:-}"
  if [ -z "$value" ]; then
    echo "$v is empty in $ENV_FILE" >&2
    exit 1
  fi
  case "$value" in
    *[@:/?#%\ ]*) echo "$v contains one of @ : / ? # % or a space; it would break the connection URLs" >&2; exit 1 ;;
  esac
done

if [ -z "${JWT_SECRET:-}" ]; then
  JWT_SECRET="$(kubectl -n "$NS" get secret jurwa-secrets -o jsonpath='{.data.jwt-secret}' 2>/dev/null | base64 --decode || true)"
fi
if [ -z "${JWT_SECRET:-}" ]; then
  JWT_SECRET="$(openssl rand -hex 32)"
fi

kubectl apply -f deploy/k8s/base/namespace.yaml >/dev/null
kubectl -n "$NS" create secret generic jurwa-secrets \
  --from-literal=postgres-user="$POSTGRES_USER" \
  --from-literal=postgres-password="$POSTGRES_PASSWORD" \
  --from-literal=mongo-user="$MONGO_USER" \
  --from-literal=mongo-password="$MONGO_PASSWORD" \
  --from-literal=redis-password="$REDIS_PASSWORD" \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secret jurwa-secrets is ready in namespace $NS (keys only):"
kubectl -n "$NS" describe secret jurwa-secrets | sed -n '/^Data/,$p'
