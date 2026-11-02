# reassignment

Owner: Nyendrak Yoezer Zangmo. gRPC port 9014. Contract: `proto/reassignment/v1/reassignment.proto`.

## What it does

Keeps who each report is assigned to, and the full history of hand-overs (FR-16, FR-19). The first assignment comes from routing (ROUTED). After that a Category Head, or the staff member holding the report, can pass it to someone else (REASSIGNED), and every change publishes `report.reassigned`. Exactly one assignment per report is current. It never changes a report's status.

## Rules

| Rule | Result |
|---|---|
| `report_id`, `to_user_id`, `actor_id` must be UUIDs | otherwise `INVALID_ARGUMENT` |
| `reason` not blank, at most 500 characters (trimmed) | otherwise `INVALID_ARGUMENT` |
| `to_user_id` must differ from the current assignee | otherwise `INVALID_ARGUMENT` |
| Report must have an assignment | otherwise `NOT_FOUND` |
| Reassign: a Category Head, or STAFF who is the current assignee | otherwise `PERMISSION_DENIED` |
| List: Category Head, Staff or System Administrator | otherwise `PERMISSION_DENIED` |
| Two reassignments at the same time | the second gets `ABORTED` and must retry |

The proto messages have no role field, so the caller's role travels in the gRPC metadata key `x-user-role` (for example `STAFF`), the same convention as internal-notes (issue I8). api-gateway sets it from the JWT, together with `actor_id`. A call without it is denied.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `ReassignReport` | Hands the report to another user, publishes `report.reassigned` | `INVALID_ARGUMENT`, `NOT_FOUND`, `PERMISSION_DENIED`, `ABORTED` |
| `ListReassignments` | All assignments, oldest first; the last one has `current = true`. Unknown report gives an empty list | `INVALID_ARGUMENT`, `PERMISSION_DENIED` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Assignment`, `NewRouted`, `Reassign`, `CheckCanReassign`, `CheckCanList` (100% tested) |
| `application/` | Use cases: `RecordRouted`, `ReassignReport`, `ListReassignments` |
| `ports/` | `AssignmentRepository`, `EventPublisher`, `ReportReassigned` |
| `adapters/in/grpc/` | gRPC handler, reads `x-user-role`, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory history (PostgreSQL later, schema in `docs/design/schema/reassignment.sql`) |
| `adapters/out/logpublisher/` | Writes `report.reassigned` as a JSON log line (Redis Streams later) |

Not done yet: consuming `report.routed` (until then `SEED_ASSIGNMENTS` records routed reports), and the ESCALATED and COMPENSATED steps of the escalation Saga (SCRUM-89).

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9014` | gRPC port |
| `SERVICE_NAME` | `reassignment` | Name written in every log line |
| `SEED_ASSIGNMENTS` | empty | `reportID:userID:position,...` (position optional). Records ROUTED assignments at start, for demos |

## Run

PowerShell:

```powershell
$env:SEED_ASSIGNMENTS = "11111111-1111-1111-1111-111111111111:22222222-2222-2222-2222-222222222222:DSA"
go run ./services/reassignment/cmd/server
```

In Git Bash (PowerShell 5.1 removes the JSON quotes when it calls grpcurl):

```bash
grpcurl -plaintext localhost:9014 grpc.health.v1.Health/Check
grpcurl -plaintext -H 'x-user-role: CATEGORY_HEAD' -d '{"report_id":"11111111-1111-1111-1111-111111111111","to_user_id":"33333333-3333-3333-3333-333333333333","reason":"Plumbing, needs estate staff","actor_id":"22222222-2222-2222-2222-222222222222"}' localhost:9014 jurwa.reassignment.v1.ReassignmentService/ReassignReport
grpcurl -plaintext -H 'x-user-role: STAFF' -d '{"report_id":"11111111-1111-1111-1111-111111111111"}' localhost:9014 jurwa.reassignment.v1.ReassignmentService/ListReassignments
```

Docker (from the repo root):

```powershell
docker build -f services/reassignment/Dockerfile --build-arg SERVICE=reassignment -t jurwa/reassignment .
docker run --rm -p 9014:9014 jurwa/reassignment
```

## Test

```powershell
go test ./services/reassignment/... -cover
```
