#!/usr/bin/env bash
# Lets Argo CD read the private Jurwa repo with a fine-grained GitHub token
# (Contents: Read-only, only jurwaCST/Jurwa). Run it yourself: the token is read
# without echo and goes only into the Secret "jurwa-repo" in namespace argocd,
# never into Git, a file or the shell history.
#   bash deploy/argocd/add-repo-credentials.sh
set -euo pipefail

read -r -p "GitHub username: " username
read -r -s -p "Token (hidden): " token
echo
[ -n "$token" ] || { echo "No token given" >&2; exit 1; }

kubectl -n argocd create secret generic jurwa-repo \
  --from-literal=type=git \
  --from-literal=url=https://github.com/jurwaCST/Jurwa.git \
  --from-literal=username="$username" \
  --from-literal=password="$token" \
  --dry-run=client -o yaml |
  kubectl label --local -f - argocd.argoproj.io/secret-type=repository -o yaml |
  kubectl apply -f -
unset token
echo "Argo CD can now read https://github.com/jurwaCST/Jurwa.git"
