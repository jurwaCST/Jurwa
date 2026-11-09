# deploy/knative

Owner: Nyendrak Yoezer Zangmo (SCRUM-104). How the `sla-escalation-checker` function runs on the cluster. Knative install, versions and problems found: `docs/ops/knative.md`. Use the Knative Service or the CronJob fallback, not both.

| File | What |
|---|---|
| `cronjob-fallback.yaml` | Kubernetes CronJob that runs the function once a minute (risk R3 fallback). Works without Knative |
| `sla-escalation-checker.yaml` | Knative Service: scales to zero between runs, reachable only inside the cluster |
| `pingsource.yaml` | PingSource `sla-check`: wakes the function every 2 minutes |

## CronJob fallback

The function normally is a small HTTP server (one check per request, for Knative). With `RUN_ONCE=true` it does one check, prints the counts as JSON and exits: code 0 when the check ran, 1 when sla-tracking could not be reached (the Job is then marked failed, and the next minute tries again). A run that takes longer than a minute is not doubled (`concurrencyPolicy: Forbid`).

### Deploy

The image is not built by `scripts/docker-build-all.sh` (that covers the services), so build and push it once:

```bash
docker build -f functions/sla-escalation-checker/Dockerfile -t localhost:5050/sla-escalation-checker:latest .
docker push localhost:5050/sla-escalation-checker:latest
kubectl apply -f deploy/knative/cronjob-fallback.yaml
```

Inside the cluster the registry is `jurwa-registry:5050`, as in `deploy/k8s`.

### Check

```bash
kubectl -n jurwa get cronjob sla-escalation-checker
kubectl -n jurwa get jobs --watch                     # a new Job every minute
kubectl -n jurwa logs job/<job name>                  # {"lapsed":0,"escalated":0,"skipped":0,"errors":0}
```

### For the demo (step 6)

1. Set the window first: sla-policy `UpsertPolicy` for the category, `window_hours: 1` (the shortest).
2. Submit the escalation report at least 1 hour before step 6 (the demo script says 90 minutes). report.routed starts its timer.
3. Keep sla-tracking running in that hour: timers are in memory until the database adapters are merged, so a restart loses them.
4. Within a minute after the deadline, a Job run escalates the report: its status becomes ESCALATED and it goes back to the Category Head (escalation Saga, `docs/design/escalation-saga.md`).
