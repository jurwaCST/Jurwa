# Knative on the k3d cluster

Card: SCRUM-104. Owner: Nyendrak Yoezer Zangmo. Installed and tested on 8 November 2026 on the team cluster made by `deploy/k3d/create-cluster.sh` (k3s v1.32.5). Evidence: `docs/evidence/nyendrak/sprint3/knative.txt`.

Knative runs `sla-escalation-checker` as a serverless function: no pod runs between checks, a PingSource wakes it on a schedule, and Knative removes the pod again when it is idle (scale to zero).

## Versions: why Knative 1.20 and not the newest

The current Knative release is v1.24.0 (knative.dev install pages, 8 November 2026). It does not run on our cluster: every Knative pod stopped with

```
kubernetes version "1.32.5+k3s1" is not compatible, need at least "1.35.0-0"
```

The minimum Kubernetes version per Knative release (from `defaultMinimumVersion` in `knative/pkg`, `version/version.go` on each release branch):

| Knative | Needs Kubernetes |
|---|---|
| 1.18 | 1.31 |
| 1.19, 1.20 | 1.32 |
| 1.21 | 1.33 |
| 1.22 | 1.34 |
| 1.24 | 1.35 |

1.20 is the newest line that supports 1.32, so we use its latest patches: **Serving v1.20.3, Eventing v1.20.3, Kourier v1.20.1**. Upgrading the cluster instead would have changed Choki's k3d setup for everyone on the day before the demo.

## Install (once per cluster)

```bash
S=knative-v1.20.3; K=knative-v1.20.1; E=knative-v1.20.3
kubectl apply -f https://github.com/knative/serving/releases/download/$S/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/$S/serving-core.yaml
kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/download/$K/kourier.yaml
kubectl patch configmap/config-network -n knative-serving --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

# Traefik already uses ports 80 and 443, so Kourier stays inside the cluster.
kubectl -n kourier-system patch svc kourier -p '{"spec":{"type":"ClusterIP"}}'

# The local registry is plain HTTP; Knative must not try to look up image digests there.
kubectl -n knative-serving patch configmap/config-deployment --type merge \
  --patch '{"data":{"registries-skipping-tag-resolving":"kind.local,ko.local,dev.local,jurwa-registry:5050,localhost:5050"}}'

kubectl apply -f https://github.com/knative/eventing/releases/download/$E/eventing-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/$E/eventing-core.yaml

kubectl get pods -n knative-serving     # activator, autoscaler, controller, net-kourier-controller, webhook: Running
kubectl get pods -n knative-eventing    # eventing-controller, eventing-webhook, pingsource-mt-adapter, ...: Running
```

## Deploy the function

```bash
docker build -f functions/sla-escalation-checker/Dockerfile -t localhost:5050/sla-escalation-checker:latest .
docker push localhost:5050/sla-escalation-checker:latest
kubectl apply -f deploy/knative/sla-escalation-checker.yaml -f deploy/knative/pingsource.yaml
kubectl -n jurwa get ksvc,pingsource
```

The function calls sla-tracking (`sla-tracking:9012`) and status-lifecycle (`status-lifecycle:9011`), so those must be deployed in the `jurwa` namespace (`kubectl apply -k deploy/k8s/services/<name>`).

## Check

```bash
kubectl -n jurwa get pods -l serving.knative.dev/service=sla-escalation-checker -w
kubectl -n jurwa logs -l serving.knative.dev/service=sla-escalation-checker -c user-container
```

A pod appears at every even minute (the PingSource schedule `*/2 * * * *`), logs `escalation run finished` with the counts, and is removed about a minute later. Between runs no pod runs.

## Problems found

| Problem | Fix |
|---|---|
| Knative v1.24 pods crash: Kubernetes 1.32 is too old | Use Knative 1.20 (table above) |
| Removing the failed v1.24 install hung: its webhook was not running, so it blocked its own delete | Delete Knative's validating and mutating webhook configurations, then the namespaces finish deleting |
| Kourier asks for a LoadBalancer on ports 80 and 443, which Traefik uses | Make the `kourier` service ClusterIP; the function is called inside the cluster only (`cluster-local`) |
| The registry `jurwa-registry:5050` is HTTP, Knative expects HTTPS when it resolves image tags | `registries-skipping-tag-resolving` in `config-deployment` |

## Fallback (risk R3)

If Knative is not available on the demo cluster, `deploy/knative/cronjob-fallback.yaml` runs the same image as a Kubernetes CronJob every minute (`RUN_ONCE=true`). Use one or the other, not both.
