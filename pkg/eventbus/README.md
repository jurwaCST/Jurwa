# pkg/eventbus

The shared event bus. Services publish protobuf events (`proto/events/v1/events.proto`) to Redis Streams and read them in consumer groups.

## Interfaces (`bus.go`)

```go
type Publisher interface {
    Publish(ctx context.Context, stream string, msg proto.Message) error
}
type Handler func(ctx context.Context, payload []byte) error
type Subscriber interface {
    Subscribe(ctx context.Context, stream, group, consumer string, h Handler) error
}
```

| Implementation | Use |
|---|---|
| `eventbus.NewMemory()` | unit tests, and `make run` without Redis |
| `eventbus.NewRedis(url)` | real bus, when `REDIS_URL` is set |

## Streams

| Constant | Stream | Payload |
|---|---|---|
| `StreamReportSubmitted` | `events:report.submitted` | `ReportSubmitted` |
| `StreamReportRouted` | `events:report.routed` | `ReportRouted` |
| `StreamStatusChanged` | `events:status.changed` | `StatusChanged` |
| `StreamReportReassigned` | `events:report.reassigned` | `ReportReassigned` |
| `StreamSlaLapsed` | `events:sla.lapsed` | `SlaLapsed` |
| `StreamEscalationCompensated` | `events:escalation.compensated` | `EscalationCompensated` |

Who publishes and who reads: `docs/design/database-schema.md` section 6.

## Rules

- **Consumer group = your service name** (for example `audit-log`). Each group gets every message; consumers in one group share them.
- A new group starts at the **first** entry of the stream, so events published before your service started are not lost.
- A handler that returns an error is tried `MaxAttempts` (3) times. Then the message goes to `<stream>.dead` (fields `payload`, `error`, `group`, `source_id`) and is acked.
- A message can arrive more than once (for example after a crash). **Handlers must ignore a repeated `meta.event_id`.**
- `Subscribe` blocks. Run it in a goroutine; it returns `nil` when the context is cancelled.

## Environment

`REDIS_URL`, for example `redis://:PASSWORD@localhost:6379/0`. The password is `REDIS_PASSWORD` in `dev/.env` (never commit it).

## Example

```go
var bus interface{ eventbus.Publisher; eventbus.Subscriber } = eventbus.NewMemory()
if url := os.Getenv("REDIS_URL"); url != "" {
    rb, err := eventbus.NewRedis(url)
    if err != nil { log.Fatal(err) }
    bus = rb
}
go bus.Subscribe(ctx, eventbus.StreamSlaLapsed, "reassignment", hostname, func(ctx context.Context, p []byte) error {
    var e eventsv1.SlaLapsed
    if err := proto.Unmarshal(p, &e); err != nil { return err }
    return app.Escalate(ctx, e.GetMeta().GetEventId(), e.GetReportId())
})
err := bus.Publish(ctx, eventbus.StreamReportSubmitted, &eventsv1.ReportSubmitted{Meta: meta, Report: r})
```

## Tests

```bash
go test ./pkg/eventbus/... -cover                    # in-memory bus and the Redis bus on miniredis
go test -tags=integration ./pkg/eventbus/... -v      # Redis bus on a real Redis container (Docker running)
REDIS_TEST_URL=redis://:PASSWORD@localhost:6379/0 go test -tags=integration ./pkg/eventbus/... -v   # against dev/docker-compose.yml
```

Both run the same checks (`contract_test.go`).
