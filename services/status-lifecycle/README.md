# status-lifecycle

Owner: Nyendrak Yoezer Zangmo. gRPC port 9011. Contract: `proto/status-lifecycle/v1/status-lifecycle.proto`.

## What it does

Owns the current status of every report and enforces the status flow. Every change is checked, saved with a new version number and published as a `status.changed` event. It never stores report content (report-submission does) or the full history (report-history and audit-log do).

## Status flow

| From | To | Who |
|---|---|---|
| OPEN | ASSIGNED | user or system (routing) |
| ASSIGNED | IN_PROGRESS | user or system |
| IN_PROGRESS | RESOLVED | user or system |
| ESCALATED | IN_PROGRESS | user or system (new assignee starts work) |
| OPEN, ASSIGNED, IN_PROGRESS | ESCALATED | system only (SLA lapsed) |
| ESCALATED | its previous status | system only (Saga compensation) |

Everything else is rejected with `FAILED_PRECONDITION`, for example RESOLVED to OPEN, OPEN to RESOLVED and RESOLVED to ESCALATED. "System" means the request has an empty `actor_id`.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `ChangeStatus` | Moves a report to `to_status`, publishes `status.changed` | `INVALID_ARGUMENT` bad id or status, `FAILED_PRECONDITION` flow does not allow it, `ABORTED` `expected_version` does not match, `NOT_FOUND` unknown report |
| `GetStatus` | Returns the current status, previous status and version | `INVALID_ARGUMENT`, `NOT_FOUND` |

Plus the standard `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Status`, `CheckTransition`, `ReportStatus.ChangeTo` (the rules, 100% tested) |
| `application/` | Use cases: `Register`, `GetStatus`, `ChangeStatus` |
| `ports/` | `StatusRepository`, `EventPublisher` interfaces |
| `adapters/in/grpc/` | gRPC handler, maps domain errors to gRPC codes |
| `adapters/out/memory/` | In-memory repository (Sprint 1; PostgreSQL in Sprint 2, schema in `docs/design/schema/status-lifecycle.sql`) |
| `adapters/out/logpublisher/` | Logs events as JSON (Sprint 1; Redis Streams in Sprint 2) |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9011` | gRPC port |
| `SERVICE_NAME` | `status-lifecycle` | Name written in every log line |
| `SEED_REPORT_IDS` | empty | Comma-separated report UUIDs registered as OPEN at start-up. For demos until `report.submitted` is consumed in Sprint 2 |

## Run (PowerShell)

```powershell
$env:SEED_REPORT_IDS = "11111111-1111-1111-1111-111111111111,33333333-3333-3333-3333-333333333333"
make run SERVICE=status-lifecycle
```

In a second terminal:

```powershell
grpcurl -plaintext localhost:9011 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{\"report_id\":\"11111111-1111-1111-1111-111111111111\",\"to_status\":\"ASSIGNED\"}' localhost:9011 jurwa.statuslifecycle.v1.StatusLifecycleService/ChangeStatus
grpcurl -plaintext -d '{\"report_id\":\"33333333-3333-3333-3333-333333333333\",\"to_status\":\"RESOLVED\"}' localhost:9011 jurwa.statuslifecycle.v1.StatusLifecycleService/ChangeStatus
```

The second call succeeds. The third returns `FailedPrecondition`. In Git Bash, use single quotes without the backslashes: `-d '{"report_id":"...","to_status":"ASSIGNED"}'`.

Docker:

```powershell
make docker-build SERVICE=status-lifecycle
docker run --rm -p 9011:9011 -e SEED_REPORT_IDS=11111111-1111-1111-1111-111111111111 jurwa/status-lifecycle
```

## Test

```powershell
go test ./services/status-lifecycle/... -cover
```
