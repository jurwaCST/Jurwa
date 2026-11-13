# notification

Owner: Gayley Choden. gRPC port 9018. Contract: `proto/notification/v1/notification.proto`.

## What it does

Sends in-app messages to users and keeps an outbox of them, so the PWA can show "your report is In Progress" or "a report was assigned to you". Event consumers (Sprint 3) and the `digest-dispatcher` function call `SendNotification`. It never decides who should be told: the caller chooses the recipient.

In Sprint 2 the `Sender` only writes a log line. Real email delivery is out of scope.

## Rules

| Rule | Result |
|---|---|
| `recipient_user_id` must be a UUID | otherwise `INVALID_ARGUMENT` |
| `channel` must be `EMAIL` or `IN_APP` | otherwise `INVALID_ARGUMENT` |
| `event_type` must be a known event (below) | otherwise `INVALID_ARGUMENT` |
| `report_id` is optional, but must be a UUID when given | otherwise `INVALID_ARGUMENT` |
| `subject` is required, not blank | otherwise `INVALID_ARGUMENT` |
| `body` is required, at most 200 characters (counted as letters, so Dzongkha works) | empty or 201+ gives `INVALID_ARGUMENT` |
| The sender fails | stored as `FAILED` (kept for a retry in Sprint 3), call returns `UNAVAILABLE` |
| Sent | stored as `SENT` with `sent_at` |
| `ListNotifications` | only that user's `SENT` messages, newest first. Failed ones are hidden because the user never received them |
| `page_size` | 0 means 20; 1 to 100; otherwise `INVALID_ARGUMENT` |

Event types are the stream names in `proto/events/v1/events.proto`, plus the digest:

| `event_type` | When |
|---|---|
| `report.submitted` | the student's report was received |
| `report.routed` | a report was assigned to a Category Head |
| `status.changed` | the report status changed |
| `report.reassigned` | the report was given to someone else |
| `sla.lapsed` | the report was escalated |
| `digest.daily` | the daily digest from `digest-dispatcher` |

## Messages from events (SCRUM-155)

When `REDIS_URL` is set, the service reads two streams in the consumer group `notification` and tells the student about their report:

| Event | Message to the student |
|---|---|
| `report.submitted` | "We received your report" with the place |
| `status.changed` to `IN_PROGRESS` | "Your report is in progress" |
| `status.changed` to `RESOLVED` | "Your report is resolved" (and that they can rate the fix) |
| `status.changed` to `ESCALATED` | "Your report was escalated" (it went to the Category Head) |
| `status.changed` to `OPEN` or `ASSIGNED` | nothing: Assigned comes a moment after "received" |

- `status.changed` does not name the student, so the report is looked up in report-submission (`GetReport`). This works after a restart and for reports sent before the service started.
- Each message's id is made from the event id (UUID v5), and it is added with `INSERT ... ON CONFLICT (id) DO NOTHING`. The same event read twice gives one message, also after a restart, and a message already read stays read.
- report-submission down: the handler returns the error, so the bus tries the event again (3 times) and then keeps it in `events:status.changed.dead`. It is never dropped. A report report-submission does not know is skipped, since trying again would not help.
- With the store in memory the streams are read again from the start on every start, so messages come back after a restart.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `SendNotification` | Delivers and stores one message | `INVALID_ARGUMENT`, `UNAVAILABLE` |
| `ListNotifications` | One user's messages, newest first, with paging | `INVALID_ARGUMENT` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Notification`, `NewNotification` rules, `MarkSent`, `MarkFailed` (100% tested) |
| `application/` | Use cases: `Send`, `SendForEvent` (once per event), `List` |
| `ports/` | `NotificationRepository`, `Sender` and `Reports` interfaces |
| `adapters/in/grpc/` | gRPC handler, maps errors to gRPC codes |
| `adapters/in/events/` | Reads `report.submitted` and `status.changed` and sends the student a message |
| `adapters/out/reportclient/` | `Reports` over gRPC: report-submission `GetReport` |
| `adapters/out/memory/` | In-memory outbox, used when `DATABASE_URL` is not set |
| `adapters/out/postgres/` | PostgreSQL outbox, used when `DATABASE_URL` is set |
| `adapters/out/repotest/` | Checks both must pass |
| `adapters/out/logsender/` | `Sender` that writes a JSON log line (the body is not logged) |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9018` | gRPC port |
| `SERVICE_NAME` | `notification` | Name written in every log line |
| `REDIS_URL` | none | For example `redis://localhost:6379`. Unset: no messages from events |
| `REPORT_SUBMISSION_ADDR` | `localhost:9001` | report-submission, to find the student of a report |
| `DATABASE_URL` | none | For example `postgres://jurwa:<password>@localhost:5432/notification?sslmode=disable`. Unset: in memory, lost on restart. A wrong URL stops the service instead of falling back. The password comes from the environment (a Kubernetes Secret), never from the repository |

Storage: with `DATABASE_URL` the service uses PostgreSQL (database `notification`, table `notifications`). On start it runs `docs/design/schema/notification.sql` (embedded as `adapters/out/postgres/schema.sql`, a test keeps the two equal), which creates the table if needed. Both stores pass the same checks (`adapters/out/repotest`).

## Run (PowerShell)

```powershell
go run ./services/notification/cmd/server
```

In a second terminal:

```powershell
grpcurl -plaintext localhost:9018 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{\"recipient_user_id\":\"22222222-2222-2222-2222-222222222222\",\"channel\":\"IN_APP\",\"event_type\":\"status.changed\",\"subject\":\"Report updated\",\"body\":\"Your report is now In Progress.\"}' localhost:9018 jurwa.notification.v1.NotificationService/SendNotification
grpcurl -plaintext -d '{\"recipient_user_id\":\"22222222-2222-2222-2222-222222222222\"}' localhost:9018 jurwa.notification.v1.NotificationService/ListNotifications
```

Docker (from the repo root):

```powershell
docker build -f services/notification/Dockerfile --build-arg SERVICE=notification -t jurwa/notification .
docker run --rm -p 9018:9018 jurwa/notification
```

## Test

```powershell
go test ./services/notification/... -cover
```

Integration tests on a real PostgreSQL 16 (Docker Desktop must be running):

```powershell
go test -tags=integration ./services/notification/adapters/out/postgres/... -v
```

Evidence: `docs/evidence/gayley/sprint3/db-adapters-3b-integration.txt` and `db-adapters-3b-live.txt`.
