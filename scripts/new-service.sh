#!/usr/bin/env bash
# Create a new service from services/_template.
# Usage: scripts/new-service.sh NAME PORT
# Example: scripts/new-service.sh routing-rules 9006
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 NAME PORT   (example: $0 routing-rules 9006)" >&2
  exit 1
fi

NAME="$1"
PORT="$2"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TEMPLATE="$ROOT/services/_template"
TARGET="$ROOT/services/$NAME"

if ! [[ "$NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "Error: NAME must be lowercase letters, digits and hyphens, for example routing-rules" >&2
  exit 1
fi
if ! [[ "$PORT" =~ ^[0-9]+$ ]] || [ "$PORT" -lt 1024 ] || [ "$PORT" -gt 65535 ]; then
  echo "Error: PORT must be a number between 1024 and 65535" >&2
  exit 1
fi
if [ -e "$TARGET" ]; then
  echo "Error: $TARGET already exists. Refusing to overwrite it." >&2
  exit 1
fi

cp -R "$TEMPLATE" "$TARGET"

# Replace the placeholders in every file. -i.bak works on both GNU sed (Git Bash) and BSD sed (macOS).
find "$TARGET" -type f | while read -r f; do
  sed -i.bak -e "s/__SERVICE__/$NAME/g" -e "s/__PORT__/$PORT/g" "$f"
  rm -f "$f.bak"
done

echo "Created services/$NAME on port $PORT."
echo
echo "Next steps:"
echo "  1. make proto                              (generate code from proto/$NAME/v1/)"
echo "  2. write tests first in services/$NAME/domain/"
echo "  3. register your gRPC service in services/$NAME/cmd/server/main.go"
echo "  4. make test && make lint"
echo "  5. make run SERVICE=$NAME"
echo "     grpcurl -plaintext localhost:$PORT grpc.health.v1.Health/Check"
