# report-submission

Owner: Sonam Choki. gRPC port 9001. Contract: `proto/report-submission/v1/report-submission.proto`.

## What it does

Accepts the student's report form, checks the form rules and stores the report with status OPEN. It never routes reports (routing-engine does, using only category and location) and never changes a report's status after creation (status-lifecycle does).

## RPCs

| RPC | Does | Errors |
|---|---|---|
| `SubmitReport` | Validates the form, stores the report as OPEN, returns it with a new UUID and `created_at` | `INVALID_ARGUMENT` when a form rule is broken |
| `GetReport` | Returns one report by id | `INVALID_ARGUMENT` when the id is not a UUID, `NOT_FOUND` when unknown |

Form rules (in `domain/report.go`, tested in `domain/report_test.go`):

| Field | Rule |
|---|---|
| `reporter_id` | Required UUID. Set by api-gateway from the JWT, never from the form |
| `category` | Required: `HOSTEL`, `ACADEMIC_BLOCK` or `OTHERS` |
| `location.building` | Required; `wing` and `area` are optional |
| `description` | Required, not blank, at most 500 characters (counted as characters, so Dzongkha text works). Never used for routing |
| `photo_id` | Optional; when given it must not be blank. The JPG/PNG and 5 MB rules are checked by photo-media at upload |

## Layout

```
domain/report.go                 Report, form rules, RoutingKey (category + location only)
application/service.go           SubmitReport and GetReport use cases
ports/ports.go                   ReportRepository, EventPublisher and IDGenerator interfaces
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository (no MONGO_URL)
adapters/out/mongo/              MongoDB repository, collection "reports" (MONGO_URL)
adapters/out/repotest/           checks every repository must pass (memory and mongo)
adapters/out/logpublisher/       logs events (no REDIS_URL)
adapters/out/eventpublisher/     publishes events:report.submitted with pkg/eventbus (REDIS_URL)
adapters/out/uuid/               UUID v4 generator
cmd/server/main.go               wires everything and starts the server
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9001` | gRPC port |
| `SERVICE_NAME` | `report-submission` | Name written in every log line |
| `MONGO_URL` | none | For example `mongodb://jurwa:PASSWORD@localhost:27017/?authSource=admin`. Not set: reports are kept in memory and lost on restart |
| `MONGO_DB` | `report_submission` | Database name |
| `REDIS_URL` | none | For example `redis://:PASSWORD@localhost:6379/0`. Set: every new report is published on `events:report.submitted` (payload `ReportSubmitted`, correlation id = report id). Not set: the event is only logged |

The first log lines say which is used: `storage: mongodb` or `storage: memory`, `events: redis streams` or `events: log only`. In Kubernetes both variables come from the Secret `jurwa-secrets` (see `deploy/README.md`). Passwords are in `dev/.env` locally, never in Git.

## Run

```bash
make run SERVICE=report-submission
```

In a second terminal:

```bash
grpcurl -plaintext localhost:9001 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"reporter_id":"0b9e8d7c-6a5b-4c3d-9e2f-1a0b9c8d7e6f","category":"HOSTEL","location":{"building":"Hostel B","wing":"Wing 2","area":"Bathroom 2"},"description":"Tap in the bathroom is leaking."}' \
  localhost:9001 jurwa.reportsubmission.v1.ReportSubmissionService/SubmitReport

grpcurl -plaintext -d '{"report_id":"<id from SubmitReport>"}' \
  localhost:9001 jurwa.reportsubmission.v1.ReportSubmissionService/GetReport
```

Expected: health `SERVING`; SubmitReport returns the report with `"status": "OPEN"`; an empty `description` returns `Code: InvalidArgument`.

If port 9001 is already used on your machine (`lsof -nP -iTCP:9001 -sTCP:LISTEN`), run on another port: `PORT=9101 make run SERVICE=report-submission` and call `localhost:9101`.

Docker:

```bash
make docker-build SERVICE=report-submission
docker run --rm -p 9001:9001 jurwa/report-submission
```

## Test

```bash
make test
go test ./services/report-submission/... -cover
```

Coverage: 100% on `domain`, `application` and the adapters. The MongoDB adapter is tested on a real MongoDB (Docker must be running):

```bash
go test -tags=integration ./services/report-submission/adapters/out/mongo/... -v
```

Evidence: `docs/evidence/sonam-choki/sprint3/report-submission-integration.txt`.
