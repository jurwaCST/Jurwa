#!/usr/bin/env bash
# Deletes the local k3d cluster "jurwa" and its registry.
set -euo pipefail

CLUSTER_NAME="jurwa"

k3d cluster delete "$CLUSTER_NAME"
