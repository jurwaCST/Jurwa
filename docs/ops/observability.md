# Observability: metrics, dashboard, logs

Card: SCRUM-105. Owner: Gayley Choden. Code: `pkg/observability/`. Manifests: `deploy/observability/`.

Every Jurwa gRPC service can report the same numbers: how many requests it handled, how many failed and how long they took. Prometheus collects them every 15 seconds and Grafana shows them on one dashboard. Each request also writes one JSON log line with a correlation id.

```
service pod  --/metrics on :9100-->  Prometheus (scrapes every 15 s)  -->  Grafana dashboard "Jurwa services"
```

## Start it

The services must be deployed first (`kubectl apply -k deploy/k8s`, see `deploy/README.md`). Then:

```bash
kubectl apply -k deploy/observability
kubectl -n jurwa get pods -l 'app.kubernetes.io/name in (prometheus,grafana)'
kubectl -n jurwa port-forward svc/grafana 3000:3000
```

Open http://localhost:3000. The home page is the **Jurwa services** dashboard (folder Jurwa). You can look without logging in (Viewer). To edit, log in as `admin`: the password is the key `grafana-admin-password` in the Secret `jurwa-secrets` when it is there, otherwise `admin`, and Grafana asks for a new one.

Prometheus itself: `kubectl -n jurwa port-forward svc/prometheus 9090:9090`, then http://localhost:9090/targets lists every pod it scrapes and whether it is up.

Remove it again: `kubectl delete -k deploy/observability`.

## Add metrics to a service

Three changes, each owner in their own service:

1. `cmd/server/main.go`:

   ```go
   obs, err := observability.Init(serviceName) // reads METRICS_ADDR
   if err != nil { slog.Error("cannot start observability", "error", err); os.Exit(1) }
   srv := grpc.NewServer(grpc.ChainUnaryInterceptor(obs.Unary()))
   ```

   Optional: `logger := obs.Logger()` instead of making your own JSON logger.
2. `Dockerfile`: add `COPY pkg/ pkg/` after `COPY proto/ proto/`, or the image does not build (the Dockerfiles copy only `proto/` and the service folder).
3. Kubernetes (Choki's generated manifest): the environment variable `METRICS_ADDR=:9100`, the container port `9100` named `metrics`, and on the pod template:

   ```yaml
   annotations:
     prometheus.io/scrape: "true"
     prometheus.io/port: "9100"
   ```

Leave `METRICS_ADDR` empty when running locally: no port is opened and nothing changes.

Done in: identity, authorization, notification, feedback, analytics (Gayley). The api-gateway is Zangmo's.

## What each metric means

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `grpc_requests_total` | counter | `service`, `method`, `code` | Requests handled. `code` is the gRPC result: `OK`, or an error such as `NotFound`, `Unauthenticated`, `InvalidArgument`, `Internal` |
| `grpc_request_duration_seconds` | histogram | `service`, `method` | Time per request, buckets from 5 ms to 10 s |
| `up` | gauge (Prometheus) | `app`, `pod` | 1 when the last scrape of that pod worked, 0 when not |
| `go_*`, `process_*` | standard | | Memory, goroutines, CPU of each service |

Health checks (`/grpc.health.v1.Health/...`) are counted, but the dashboard leaves them out: Kubernetes probes call them every few seconds and they would hide the real traffic.

## The dashboard

| Panel | Query | Read it as |
|---|---|---|
| Request rate per service | `sum by (service) (rate(grpc_requests_total[1m]))` | requests per second |
| Error rate per service | the same with `code!="OK"` | failed requests per second. A wrong password counts here too (`Unauthenticated`) |
| p95 duration per service | `histogram_quantile(0.95, ...grpc_request_duration_seconds_bucket[5m])` | 95 of 100 requests were faster than this |
| Pods up | `up{job="jurwa-pods"}` | green 1 per pod that is scraped |

The dashboard is a file, `deploy/observability/dashboards/jurwa-services.json`. Changes made in the Grafana page are lost when the pod restarts: export the JSON and replace the file.

## Logs

Every request gives one line on stdout, for example:

```json
{"time":"2026-11-06T11:04:06.22Z","level":"INFO","msg":"grpc request","service":"identity","method":"/jurwa.identity.v1.IdentityService/Login","code":"OK","duration_ms":48.38,"correlation_id":"dc0da236-..."}
```

`level` is `WARN` when the request failed. Read them with `kubectl -n jurwa logs deploy/<service>`. The `correlation_id` comes from the request metadata `x-correlation-id`, or is new when the caller sent none. When the gateway sends one, the same id shows in every service's line for that request.

## Checked

On 6 November 2026, on a local k3d cluster: Prometheus scraped identity and authorization (`up` = 1 for both), 2.5 minutes of logins and permission checks gave about 1.3 to 1.4 requests per second per service, the wrong-password logins showed as `Unauthenticated` errors, and p95 was about 97 ms for identity (password hashing) and 5 ms for authorization.

Evidence in `docs/evidence/gayley/sprint3/`: `observability-tests.txt` (20 tests, 98.8% coverage), `observability-live.txt` (one service on a laptop), `observability-k8s.txt` (Prometheus queries in the cluster), `observability-grafana.png` (the dashboard on 7 November 2026 with live traffic: identity about 1.2 req/s with wrong-password logins as errors, authorization about 0.8 req/s, p95 about 100 to 150 ms for identity and 5 ms for authorization, both pods up).

## Not done

Tracing with Jaeger (Stretch) is not done: no OpenTelemetry spans yet. The correlation id in the logs is the way to follow one request across services for now.
