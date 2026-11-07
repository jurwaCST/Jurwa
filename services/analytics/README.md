# analytics

Owner: Gayley Choden. gRPC port 9020. Contract: `proto/analytics/v1/analytics.proto`.

## What it does

Serves the dashboard counts for the System Administrator and the read-only College President view: how many reports are in each status, split by category, and how long reports take to resolve. It is built from events only (`report.submitted` and `status.changed`). It never calls other services and never changes a report.

## Rules

| Rule | Result |
|---|---|
| `open`, `assigned`, `in_progress`, `resolved` | reports by their status now |
| `escalated` | reports escalated at least once, even if later resolved (from the `.proto`) |
| `total` | every report, so it is not always the sum of the other fields |
| `average_resolution_hours` | mean time from submitted to RESOLVED, for reports resolved now; 0 when none |
| Period | counts reports created in it. `from` is included, `to` is not. Unset means no limit |
| `from` after `to` | `INVALID_ARGUMENT` |
| Category breakdown | always HOSTEL, ACADEMIC_BLOCK, OTHERS in that order, even with 0 reports |
| No events yet | every count is 0 |
| Same event id twice | ignored (events can be delivered more than once) |
| Unknown status or category, status change for an unknown report, missing ids | ignored, never a crash |

## How events get in

| Sprint | Source |
|---|---|
| 2 | `SEED_DEMO_DATA=true` loads 15 demo events (6 reports) at start-up, see `cmd/server/demo.go` |
| 3 | A Redis Streams consumer reads `events:report.submitted` and `events:status.changed` and calls `application.Service.Record` |

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `GetSummary` | Counts per status and average resolution hours | `INVALID_ARGUMENT` |
| `GetCategoryBreakdown` | The same counts per category | `INVALID_ARGUMENT` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Event`, `Projection.Apply` (duplicates and bad events ignored), `Summarize`, `Breakdown`, `Period` (100% tested) |
| `application/` | Use cases: `Record`, `Summary`, `Breakdown` |
| `ports/` | `ProjectionStore` interface |
| `adapters/in/grpc/` | gRPC handler, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory store, used when `DATABASE_URL` is not set |
| `adapters/out/postgres/` | PostgreSQL store, used when `DATABASE_URL` is set |
| `adapters/out/repotest/` | Checks both must pass |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9020` | gRPC port |
| `SERVICE_NAME` | `analytics` | Name written in every log line |
| `SEED_DEMO_DATA` | empty | `true` loads the demo events at start-up. Sprint 2 demos only |
| `DATABASE_URL` | none | For example `postgres://jurwa:<password>@localhost:5432/analytics?sslmode=disable`. Unset: in memory, lost on restart. A wrong URL stops the service instead of falling back. The password comes from the environment (a Kubernetes Secret), never from the repository |

Storage: with `DATABASE_URL` the service uses PostgreSQL (database `analytics`, table `report_facts` and `processed_events`). On start it runs `docs/design/schema/analytics.sql` (embedded as `adapters/out/postgres/schema.sql`, a test keeps the two equal), which creates the tables if needed. The ids of applied events are kept, so a repeated event, or the demo seed after a restart, is never counted twice. Both stores pass the same checks (`adapters/out/repotest`).

## Run (PowerShell)

```powershell
$env:SEED_DEMO_DATA = "true"
go run ./services/analytics/cmd/server
```

In a second terminal:

```powershell
grpcurl -plaintext localhost:9020 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9020 jurwa.analytics.v1.AnalyticsService/GetSummary
grpcurl -plaintext localhost:9020 jurwa.analytics.v1.AnalyticsService/GetCategoryBreakdown
grpcurl -plaintext -d '{\"period\":{\"from\":\"2026-11-01T00:00:00Z\"}}' localhost:9020 jurwa.analytics.v1.AnalyticsService/GetSummary
```

Docker (from the repo root):

```powershell
docker build -f services/analytics/Dockerfile --build-arg SERVICE=analytics -t jurwa/analytics .
docker run --rm -p 9020:9020 -e SEED_DEMO_DATA=true jurwa/analytics
```

## Test

```powershell
go test ./services/analytics/... -cover
```

Integration tests on a real PostgreSQL 16 (Docker Desktop must be running):

```powershell
go test -tags=integration ./services/analytics/adapters/out/postgres/... -v
```

Evidence: `docs/evidence/gayley/sprint3/db-adapters-3b-integration.txt` and `db-adapters-3b-live.txt`.
