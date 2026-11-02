# audit-log

Owner: Nyendrak Yoezer Zangmo. gRPC port 9015. Contract: `proto/audit-log/v1/audit-log.proto`.

## What it does

An append-only record of every action on a report: submission, routing, status changes, reassignments, escalations and compensations (FR-17), including a supervisor acting for staff without an email (FR-24). Entries are never changed or deleted: the service has no RPC for it, and the store has no update or delete method.

## Rules

| Rule | Result |
|---|---|
| `service` required, at most 64 characters | otherwise `INVALID_ARGUMENT` |
| `action` is one of REPORT_SUBMITTED, ROUTED, STATUS_CHANGED, REASSIGNED, ESCALATED, COMPENSATED | otherwise `INVALID_ARGUMENT` |
| `report_id` and `actor_id`, when sent, are UUIDs | otherwise `INVALID_ARGUMENT` |
| `on_behalf_of` at most 500 characters | otherwise `INVALID_ARGUMENT` |
| `details` at most 20 entries, keys 1 to 64 characters, values at most 500 | otherwise `INVALID_ARGUMENT` |
| `id` is set by the server; an id in the request is ignored | |
| `occurred_at` missing | server time is used |
| `ListByReport` only for the System Administrator | otherwise `PERMISSION_DENIED` |

`Append` is called by other services and needs no role. `ListByReport` reads the caller's role from the gRPC metadata key `x-user-role`, set by api-gateway from the JWT, the same convention as internal-notes (issue I8).

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `Append` | Stores one entry, returns its id | `INVALID_ARGUMENT` |
| `ListByReport` | A report's entries, oldest first (same time keeps append order). Unknown report gives an empty list | `INVALID_ARGUMENT`, `PERMISSION_DENIED` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Event`, `NewEvent` checks, `CheckCanList` (100% tested) |
| `application/` | Use cases: `Append`, `ListByReport` |
| `ports/` | `EventStore` (append and list only) |
| `adapters/in/grpc/` | gRPC handler, reads `x-user-role`, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory store (MongoDB collection `audit_events` later, see `docs/design/database-schema.md`) |

Not done yet: the MongoDB adapter, and consuming the events listed in `docs/design/bounded-contexts.md` (until then services call `Append` directly).

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9015` | gRPC port |
| `SERVICE_NAME` | `audit-log` | Name written in every log line |

## Run

```powershell
go run ./services/audit-log/cmd/server
```

In Git Bash (PowerShell 5.1 removes the JSON quotes when it calls grpcurl):

```bash
grpcurl -plaintext localhost:9015 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{"event":{"service":"status-lifecycle","action":"STATUS_CHANGED","report_id":"11111111-1111-1111-1111-111111111111","details":{"from":"OPEN","to":"ASSIGNED"}}}' localhost:9015 jurwa.auditlog.v1.AuditLogService/Append
grpcurl -plaintext -H 'x-user-role: SYSTEM_ADMIN' -d '{"report_id":"11111111-1111-1111-1111-111111111111"}' localhost:9015 jurwa.auditlog.v1.AuditLogService/ListByReport
```

Docker (from the repo root):

```powershell
docker build -f services/audit-log/Dockerfile --build-arg SERVICE=audit-log -t jurwa/audit-log .
docker run --rm -p 9015:9015 jurwa/audit-log
```

## Test

```powershell
go test ./services/audit-log/... -cover
```
