#!/usr/bin/env bash
# Installs cert-manager into the cluster (needed by deploy/k8s/ingress).
# Version from https://github.com/cert-manager/cert-manager/releases (checked 6 Nov 2026).
#   bash scripts/install-cert-manager.sh
set -euo pipefail

VERSION="v1.21.3"

kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/$VERSION/cert-manager.yaml"
kubectl -n cert-manager wait --for=condition=available deploy --all --timeout=300s
kubectl -n cert-manager get pods
