#!/usr/bin/env bash
# Build the Docker image jurwa/<name> for every service in services/ (except _template).
# Prints OK or FAILED per service and exits non-zero if any build failed.
# Usage (from anywhere): bash scripts/docker-build-all.sh
# Works in Git Bash on Windows and in bash 3.2 on macOS.
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT" || exit 1

if ! docker info > /dev/null 2>&1; then
  echo "Docker is not running. Open Docker Desktop and wait for \"Engine running\"." >&2
  exit 1
fi

LOG_DIR="$(mktemp -d)"
ok=()
failed=()

for dir in services/*/; do
  name="$(basename "$dir")"
  case "$name" in
    _*) continue ;; # _template and other helper folders are not services
  esac
  if [ ! -f "$dir/Dockerfile" ]; then
    echo "SKIPPED  $name (no Dockerfile)"
    continue
  fi

  printf 'building %s ... ' "$name"
  if make docker-build SERVICE="$name" > "$LOG_DIR/$name.log" 2>&1; then
    echo "OK"
    ok+=("$name")
  else
    echo "FAILED (log: $LOG_DIR/$name.log)"
    failed+=("$name")
  fi
done

echo
echo "Built ${#ok[@]} image(s), ${#failed[@]} failed."
if [ "${#failed[@]}" -gt 0 ]; then
  echo "Failed: ${failed[*]}"
  echo "Send the log to the service owner (or to Sonam Zangmo if every service fails: that is the template Dockerfile)."
  exit 1
fi
