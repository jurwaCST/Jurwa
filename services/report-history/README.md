# report-history

Owner: Sonam Choki. gRPC port 9003. Contract: `proto/report-history/v1/report-history.proto`.

## What it does

Keeps a read model of reports for the PWA: one summary per report (for "My reports") and a timeline of everything that happened to it. It is built from events and never changes a report itself; status-lifecycle stays the source of truth for status.

## RPCs

| RPC | Does | Errors |
|---|---|---|
| `ListReportsByReporter` | One page of a student's reports, newest first, with `next_page_token` | `INVALID_ARGUMENT` for a missing reporter id, page size outside 1 to 100, or a bad page token |
| `GetReportTimeline` | Every timeline entry of one report, oldest first | `INVALID_ARGUMENT` when the id is not a UUID, `NOT_FOUND` when the report is unknown |

Two application methods fill the store. They have no RPC, because only events call them:

| Method | Called by (Sprint 3) | Does |
|---|---|---|
| `RecordSubmitted` | `events:report.submitted` consumer | Creates the summary (status `OPEN`) and the first timeline entry |
| `Record` | `events:status.changed`, `events:report.reassigned`, `events:escalation.compensated` consumers | Adds one timeline entry and updates the summary |

Rules (in `domain/history.go`, tested in `domain/history_test.go`):

| Rule | Detail |
|---|---|
| Timeline order | Oldest first. Entries with the same time keep the order they were recorded |
| List order | Newest report first; a student only ever sees their own reports |
| Late events | An entry older than the summary's `updated_at` is kept in the timeline but does not move the status back |
| Paging | `page_size` 0 means 20; 1 to 100 allowed. `page_token` is the `next_page_token` of the previous call; `""` on the last page |
| Actor | Empty when the system acted (for example an SLA escalation), otherwise a UUID |

## Layout

```
domain/history.go                Summary, Entry, statuses, ordering, paging and rules
application/service.go           RecordSubmitted, Record, ListReportsByReporter, GetReportTimeline
ports/ports.go                   HistoryRepository interface
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository
cmd/server/main.go               wires everything and starts the server
cmd/server/sample.go             optional sample data for local demos
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9003` | gRPC port |
| `SERVICE_NAME` | `report-history` | Name written in every log line |
| `SEED_SAMPLE` | unset | `true` loads two sample reports of student `1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d`. Local demos only |

Storage is in memory in Sprint 2. Sprint 3 adds a PostgreSQL adapter (`DATABASE_URL`, `docs/design/schema/report-history.sql`) and the Redis Streams consumers.

## Run

```bash
SEED_SAMPLE=true make run SERVICE=report-history
```

In a second terminal:

```bash
grpcurl -plaintext localhost:9003 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"reporter_id":"1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"}' \
  localhost:9003 jurwa.reporthistory.v1.ReportHistoryService/ListReportsByReporter

grpcurl -plaintext -d '{"report_id":"6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"}' \
  localhost:9003 jurwa.reporthistory.v1.ReportHistoryService/GetReportTimeline
```

Expected: health `SERVING`; two reports, the Wi-Fi one first; the hostel report's timeline is `OPEN`, `ASSIGNED`, `IN_PROGRESS`.

Docker:

```bash
make docker-build SERVICE=report-history
docker run --rm -e SEED_SAMPLE=true -p 9003:9003 jurwa/report-history
```

## Test

```bash
make test
go test ./services/report-history/... -cover
```

Coverage: 100% on `domain` and both adapters, 97.6% on `application`. Evidence: `docs/evidence/sonam-choki/scrum-22-tests.txt` and `scrum-22-grpcurl.txt`.
