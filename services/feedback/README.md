# feedback

Owner: Gayley Choden. gRPC port 9019. Contract: `proto/feedback/v1/feedback.proto`.

## What it does

Stores the student's rating (1 to 5, with an optional comment) of how a report was resolved. It asks status-lifecycle whether the report is RESOLVED before it accepts feedback. It never changes a report's status.

## Rules

| Rule | Result |
|---|---|
| `report_id` and `student_user_id` must be UUIDs | otherwise `INVALID_ARGUMENT` |
| `rating` must be 1 to 5 (0 and 6 are rejected) | otherwise `INVALID_ARGUMENT` |
| `comment` is optional, at most 500 characters (counted as letters, so Dzongkha works) | 501+ gives `INVALID_ARGUMENT` |
| One feedback per report | a second one gives `ALREADY_EXISTS`, the first is kept |
| The report must be RESOLVED in status-lifecycle | otherwise `FAILED_PRECONDITION` |
| status-lifecycle cannot be reached | `UNAVAILABLE` (never a wrong "not resolved") |
| `GetFeedback` for a report with no feedback | `NOT_FOUND` |

The comment limit is 500 because that is what the `.proto` contract says (the Sprint 2 guide said 300).

Checks run in this order: the input, then "already has feedback" (local), then "is it RESOLVED" (a network call). Bad input never causes a call to status-lifecycle.

Whether the student is the person who reported it ("Student (own report)" in `docs/design/api-routes.md`) is checked by the api-gateway, which has the JWT.

## RPCs

| RPC | What it does | Errors |
|---|---|---|
| `SubmitFeedback` | Stores the rating for a resolved report | `INVALID_ARGUMENT`, `ALREADY_EXISTS`, `FAILED_PRECONDITION`, `UNAVAILABLE` |
| `GetFeedback` | Returns the feedback for a report | `INVALID_ARGUMENT`, `NOT_FOUND` |

Plus `grpc.health.v1.Health/Check` and server reflection.

## Layout

| Folder | Content |
|---|---|
| `domain/` | `Feedback`, `NewFeedback` rules (100% tested) |
| `application/` | Use cases: `Submit`, `Get` |
| `ports/` | `FeedbackRepository` and `StatusChecker` interfaces |
| `adapters/in/grpc/` | gRPC handler, maps errors to gRPC codes |
| `adapters/out/memory/` | In-memory store (a database in Sprint 3) |
| `adapters/out/statusclient/` | `StatusChecker` that calls status-lifecycle `GetStatus`, with `pkg/resilience` (timeout, retries, circuit breaker) |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9019` | gRPC port |
| `SERVICE_NAME` | `feedback` | Name written in every log line |
| `STATUS_LIFECYCLE_ADDR` | `localhost:9011` | Address of status-lifecycle |

## Run (PowerShell)

Terminal 1, status-lifecycle with one demo report:

```powershell
$env:SEED_REPORT_IDS = "33333333-3333-3333-3333-333333333333"
go run ./services/status-lifecycle/cmd/server
```

Terminal 2, feedback:

```powershell
go run ./services/feedback/cmd/server
```

Terminal 3, resolve the report, then give feedback:

```powershell
grpcurl -plaintext -d '{\"report_id\":\"33333333-3333-3333-3333-333333333333\",\"to_status\":\"ASSIGNED\",\"actor_id\":\"22222222-2222-2222-2222-222222222222\"}' localhost:9011 jurwa.statuslifecycle.v1.StatusLifecycleService/ChangeStatus
grpcurl -plaintext -d '{\"report_id\":\"33333333-3333-3333-3333-333333333333\",\"to_status\":\"IN_PROGRESS\",\"actor_id\":\"22222222-2222-2222-2222-222222222222\"}' localhost:9011 jurwa.statuslifecycle.v1.StatusLifecycleService/ChangeStatus
grpcurl -plaintext -d '{\"report_id\":\"33333333-3333-3333-3333-333333333333\",\"to_status\":\"RESOLVED\",\"actor_id\":\"22222222-2222-2222-2222-222222222222\"}' localhost:9011 jurwa.statuslifecycle.v1.StatusLifecycleService/ChangeStatus
grpcurl -plaintext -d '{\"report_id\":\"33333333-3333-3333-3333-333333333333\",\"student_user_id\":\"22222222-2222-2222-2222-222222222222\",\"rating\":4,\"comment\":\"Fixed within a day\"}' localhost:9019 jurwa.feedback.v1.FeedbackService/SubmitFeedback
grpcurl -plaintext -d '{\"report_id\":\"33333333-3333-3333-3333-333333333333\"}' localhost:9019 jurwa.feedback.v1.FeedbackService/GetFeedback
```

Docker (from the repo root):

```powershell
docker build -f services/feedback/Dockerfile --build-arg SERVICE=feedback -t jurwa/feedback .
docker run --rm -p 9019:9019 -e STATUS_LIFECYCLE_ADDR=host.docker.internal:9011 jurwa/feedback
```

## Test

```powershell
go test ./services/feedback/... -cover
```
