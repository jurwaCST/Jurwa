# Archive handover (SCRUM-85)

Owner: Gayley Choden. For whoever picks up Jurwa after Sprint 3.

## Where everything is

| Item | Location |
|---|---|
| Source code | https://github.com/jurwaCST/Jurwa |
| Final version | Git tag `v1.0.0` and the GitHub Release made from it (the source zip is attached automatically). **Waiting for** the tag, made after the last merge |
| Jira project | https://jurwa-cst.atlassian.net, project SCRUM. CSV export: **Waiting for** the export, kept in the team's shared drive, not in Git |
| Demo video | Team Drive, folder Final_Project_Group6 > "Demo rehearsal recording": https://drive.google.com/drive/folders/1orsT4XKETlzrTWPo3OBTCUdQNrLUFF3H (shared with the team; Sonam Choki uploads the recording after the rehearsal on 10 November) |
| Project documents | `docs/pm/` (charter, requirements, schedule and EVM, risks, issues, quality plan, demo script, this closure), `docs/sprint-reviews/`, `docs/testing/`, `docs/design/`, `docs/ops/` |
| Evidence | `docs/evidence/<member>/` |

## Test accounts

The identity service creates one demo account per role when started with `SEED_DEMO_USERS=true` (SCRUM-126). All share the password given in `SEED_DEMO_PASSWORD`, which is never in Git (it is in `dev/.env` locally and the Secret `jurwa-secrets` in Kubernetes).

| Email | Role |
|---|---|
| `admin.demo@jurwa.example` | System Administrator |
| `dsa.demo@jurwa.example` | Category Head (Hostel) |
| `daa.demo@jurwa.example` | Category Head (Academic Block) |
| `sso.demo@jurwa.example` | Category Head (Others) |
| `staff.demo@jurwa.example` | Staff |
| `president.demo@jurwa.example` | College President |

Students register through `POST /auth/register` and always get the Student role.

## Start the whole system from scratch

On a 16 GB machine, from the repository root, in Git Bash (details: `deploy/README.md`, `deploy/argocd/README.md`):

```bash
./deploy/k3d/create-cluster.sh                       # the k3d cluster
bash scripts/push-local-registry.sh                  # build the 20 images, push to localhost:5050
bash scripts/create-secrets.sh                       # Secret jurwa-secrets from dev/.env
bash scripts/install-cert-manager.sh                 # HTTPS (docs/ops/tls.md)
kubectl create namespace argocd
kubectl apply -n argocd --server-side -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.6/manifests/install.yaml
bash deploy/argocd/add-repo-credentials.sh
kubectl apply -f deploy/argocd/application.yaml      # Argo CD syncs everything from main
kubectl -n argocd get applications                   # Synced, Healthy
kubectl -n jurwa get pods                            # all Running
```

On an 8 GB laptop, use the small demo set instead (`deploy/argocd/README.md`, section 1, and `application-demo.yaml`).

Without Kubernetes, for development and the end-to-end tests: `web/cypress/start-stack.sh` and `docs/testing/e2e-tests.md`.

## Open work carried to Sprint 4

See `docs/pm/closure-report.md` section 2 (moved scope) and `docs/testing/test-plan.md` section 7 (known gaps).
