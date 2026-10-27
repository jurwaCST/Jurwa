#!/usr/bin/env bash
# Creates the local k3d cluster "jurwa": 1 server, 2 agents, a local image
# registry, and ports 80/443 mapped to the load balancer (Traefik stays on).
set -euo pipefail

CLUSTER_NAME="jurwa"
REGISTRY_NAME="jurwa-registry"
REGISTRY_PORT="5050"             # 5000 is taken by AirPlay on macOS

if k3d cluster list "$CLUSTER_NAME" >/dev/null 2>&1; then
  echo "Cluster '$CLUSTER_NAME' already exists. Run deploy/k3d/delete-cluster.sh first to recreate it."
  exit 0
fi

k3d cluster create "$CLUSTER_NAME" \
  --servers 1 \
  --agents 2 \
  --port "80:80@loadbalancer" \
  --port "443:443@loadbalancer" \
  --registry-create "$REGISTRY_NAME:0.0.0.0:$REGISTRY_PORT" \
  --wait

kubectl config use-context "k3d-$CLUSTER_NAME"
kubectl get nodes

echo
echo "Registry: push from your laptop to localhost:$REGISTRY_PORT/<image>:<tag>"
echo "          reference in manifests as $REGISTRY_NAME:$REGISTRY_PORT/<image>:<tag>"
