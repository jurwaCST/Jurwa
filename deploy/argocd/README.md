# Argo CD (GitOps)

Card: SCRUM-102. Owner: Sonam Zangmo. Argo CD watches `main` and keeps the cluster equal to what Git says: a change merged to `main` is applied by itself, a resource deleted from Git is deleted from the cluster (`prune`), and a change made by hand with `kubectl` is put back (`selfHeal`).

| File | What it is |
|---|---|
| `application.yaml` | Application `jurwa`: everything in `deploy/k8s` (20 services, data stores, ingress). Needs about 16 GB of RAM |
| `application-demo.yaml` | Application `jurwa-demo`: the small set in `demo/`, which fits on an 8 GB laptop |
| `demo/kustomization.yaml` | PostgreSQL, MongoDB, Redis, api-gateway, identity, authorization, category-taxonomy, report-submission (sign-up, sign-in, categories, submit a report). It only points at the folders in `deploy/k8s`, it copies nothing |
| `add-repo-credentials.sh` | Gives Argo CD read access to the private repo with a token you type in; the token goes only into a Secret in the cluster |

Use one Application, not both: they manage the same namespace.

## 1. Cluster

On a 16 GB machine use the normal cluster (`./deploy/k3d/create-cluster.sh`). On an 8 GB laptop, give Docker Desktop 6 GB (Settings, Resources, Memory) and make a one-node cluster:

```bash
k3d cluster create jurwa --servers 1 --agents 0 \
  --port "80:80@loadbalancer" --port "443:443@loadbalancer" \
  --registry-create "jurwa-registry:0.0.0.0:5050" --wait
bash scripts/push-local-registry.sh api-gateway identity authorization category-taxonomy report-submission
bash scripts/create-secrets.sh
```

(For the full set: `bash scripts/push-local-registry.sh`, `bash scripts/install-cert-manager.sh`.)

## 2. Install Argo CD

```bash
kubectl create namespace argocd
kubectl apply -n argocd --server-side -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.6/manifests/install.yaml
# Not needed here (no SSO, notifications or ApplicationSets); saves memory.
kubectl -n argocd scale deploy argocd-dex-server argocd-notifications-controller argocd-applicationset-controller --replicas=0
kubectl -n argocd rollout status deploy/argocd-server
```

## 3. Repo access

The repo is private and the jurwaCST organisation does not allow deploy keys, so Argo CD reads it over HTTPS with a **fine-grained personal access token**: github.com, Settings, Developer settings, Fine-grained tokens, Generate new token. Resource owner `jurwaCST`, only the repository `Jurwa`, permission **Contents: Read-only**, expiry 30 days. Then:

```bash
bash deploy/argocd/add-repo-credentials.sh
```

Never paste the token into a file, a chat or a commit. Revoke it on GitHub when the course ends.

## 4. Create the Application

```bash
kubectl apply -f deploy/argocd/application-demo.yaml     # or application.yaml on 16 GB
kubectl -n argocd get applications                         # SYNC STATUS Synced, HEALTH Healthy
kubectl -n jurwa get pods
```

## 5. Open the Argo CD UI

```bash
kubectl -n argocd port-forward svc/argocd-server 8443:443
# https://localhost:8443  user: admin
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 --decode; echo
```

The password is made by Argo CD in your cluster; do not write it down anywhere shared.

## 6. Show GitOps working

- **Self-heal**: `kubectl -n jurwa scale deploy/api-gateway --replicas=0`. Within seconds Argo CD sets it back to 1, because Git says 1.
- **Git change**: merge a PR that changes a value in `deploy/k8s` (for example a ConfigMap value in `scripts/gen-k8s-manifests.sh`, regenerated). Argo CD polls Git every 3 minutes and applies it; the UI shows the new revision.

Evidence: `docs/evidence/sonam-zangmo/sprint3/argocd.txt`.
