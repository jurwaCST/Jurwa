# pkg/observability

Card: SCRUM-105. Owner: Gayley Choden. The same metrics and request logs for every Jurwa gRPC service.

## Add it to a service (three lines)

In `services/<service>/cmd/server/main.go`:

```go
obs, err := observability.Init(serviceName) // reads METRICS_ADDR
if err != nil { slog.Error("cannot start metrics", "error", err); os.Exit(1) }
srv := grpc.NewServer(grpc.ChainUnaryInterceptor(obs.Unary()))
```

and set `METRICS_ADDR=:9100` in the service's Kubernetes manifest, with the pod annotations `prometheus.io/scrape: "true"` and `prometheus.io/port: "9100"`. Locally leave `METRICS_ADDR` empty: no port is opened. Optional: `slog.SetDefault(obs.Logger())` for the service's own lines.

## What each request gives

| Output | Meaning |
|---|---|
| `grpc_requests_total{service,method,code}` | Counter. One per request, `code` is the gRPC result (`OK`, `NotFound`, `Unauthenticated`, ...). A plain Go error counts as `Unknown` |
| `grpc_request_duration_seconds{service,method}` | Histogram of request time, buckets from 5 ms to 10 s. Grafana uses it for p95 |
| `go_*`, `process_*` | Standard Go and process metrics (memory, goroutines, CPU) |
| One JSON log line | `service`, `method`, `code`, `duration_ms`, `correlation_id`. `INFO` when OK, `WARN` on an error |

Health checks (`/grpc.health.v1.Health/...`, called by Kubernetes probes every few seconds) are counted but get a log line only when they fail.

## Correlation id

The interceptor reads the metadata key `x-correlation-id` (`observability.CorrelationHeader`). When it is missing a new UUID is made. Handlers read it with `observability.CorrelationID(ctx)`, for example to put it on an outgoing call so the next service logs the same id.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `METRICS_ADDR` | empty (off) | Where `/metrics` is served, for example `:9100`. A wrong address stops the service instead of running without metrics |

## Test

```bash
go test ./pkg/observability/... -cover
```

Evidence: `docs/evidence/gayley/sprint3/observability-tests.txt`.
