# sla-tracking

Owner: Nyendrak Yoezer Zangmo. gRPC port 9012. Contract: `proto/sla-tracking/v1/sla-tracking.proto`.

## What it does

Runs one SLA timer per report (FR-14 to FR-16). When a report is routed the timer starts, with a deadline of now plus the window from sla-policy for the report's category. When the report is Resolved the timer stops. When the deadline passes, the timer is marked LAPSED and an `sla.lapsed` event is published, which starts the escalation Saga (`docs/design/bounded-contexts.md`, section 5). It never changes a report's status itself.

## Rules

| Rule | Result |
|---|---|
| `report_id` must be a UUID | otherwise `INVALID_ARGUMENT` |
| Category must be HOSTEL, ACADEMIC_BLOCK or OTHERS | otherwise `INVALID_ARGUMENT` |
| One timer per report | second `StartTimer` gives `ALREADY_EXISTS` |
| sla-policy has no window for the category | `FAILED_PRECONDITION` (the System Administrator must set one first) |
| sla-policy cannot be reached | `UNAVAILABLE` |
| A timer is due when it is RUNNING and its deadline is at or before now | STOPPED and LAPSED timers are never due |
| `StopTimer` works from RUNNING or LAPSED, and twice is safe | |

### ListDueTimers marks timers lapsed

The sla-escalation-checker function calls `ListDueTimers` on a schedule. For each due timer, sla-tracking publishes `sla.lapsed`, saves the timer as LAPSED, and returns it. So a timer is escalated only once, even if the checker runs again before the Saga finishes.

- The event is published before the timer is saved. If saving fails, the next run publishes again (at least once), so consumers must ignore duplicates.
- `now` is optional and may not be later than server time (`INVALID_ARGUMENT`), so a caller cannot lapse timers early.
- `limit` is 1 to 500, 0 means 100. Earliest deadline first.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `StartTimer` | Starts the timer of a report | `INVALID_ARGUMENT`, `ALREADY_EXISTS`, `FAILED_PRECONDITION`, `UNAVAILABLE` |
| `StopTimer` | Stops it when the report is Resolved | `INVALID_ARGUMENT`, `NOT_FOUND` |
| `GetTimer` | Returns the timer | `INVALID_ARGUMENT`, `NOT_FOUND` |
| `ListDueTimers` | Lapses due timers, publishes `sla.lapsed`, returns them | `INVALID_ARGUMENT` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Timer`, `NewTimer`, `IsDue`, `Lapse`, `Stop` (100% tested) |
| `application/` | Use cases: `StartTimer`, `StopTimer`, `GetTimer`, `ListDueTimers` |
| `ports/` | `TimerRepository`, `PolicyClient`, `EventPublisher`, `SLALapsed` |
| `adapters/in/grpc/` | gRPC handler, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory timers (Redis keys `sla:deadlines` and `sla:timer:<report_id>` later, see `docs/design/database-schema.md` section 6) |
| `adapters/out/slapolicy/` | gRPC client for sla-policy `GetPolicy` (3 second timeout) |
| `adapters/out/logpublisher/` | Writes `sla.lapsed` as a JSON log line (Redis Streams `events:sla.lapsed` later) |

Not done yet: the Redis adapter and Redis Streams publisher (need a Redis client in `go.mod`, a team decision), and consuming `report.routed` and `status.changed` to start and stop timers automatically. Until then, timers are started and stopped through the RPCs.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9012` | gRPC port |
| `SERVICE_NAME` | `sla-tracking` | Name written in every log line |
| `SLA_POLICY_ADDR` | `localhost:9013` | Address of sla-policy |

## Run (PowerShell)

sla-tracking needs sla-policy for `StartTimer`. Start both, each in its own terminal:

```powershell
go run ./services/sla-policy/cmd/server
go run ./services/sla-tracking/cmd/server
```

In a third terminal, set a window and start a timer. Use Git Bash for the JSON, because PowerShell 5.1 removes the quotes when it calls grpcurl:

```bash
grpcurl -plaintext localhost:9012 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{"category":"HOSTEL","window_hours":48,"updated_by":"22222222-2222-2222-2222-222222222222"}' localhost:9013 jurwa.slapolicy.v1.SlaPolicyService/UpsertPolicy
grpcurl -plaintext -d '{"report_id":"11111111-1111-1111-1111-111111111111","category":"HOSTEL"}' localhost:9012 jurwa.slatracking.v1.SlaTrackingService/StartTimer
grpcurl -plaintext -d '{}' localhost:9012 jurwa.slatracking.v1.SlaTrackingService/ListDueTimers
grpcurl -plaintext -d '{"report_id":"11111111-1111-1111-1111-111111111111"}' localhost:9012 jurwa.slatracking.v1.SlaTrackingService/StopTimer
```

Docker (from the repo root):

```powershell
docker build -f services/sla-tracking/Dockerfile --build-arg SERVICE=sla-tracking -t jurwa/sla-tracking .
docker run --rm -p 9012:9012 -e SLA_POLICY_ADDR=host.docker.internal:9013 jurwa/sla-tracking
```

## Test

```powershell
go test ./services/sla-tracking/... -cover
```
