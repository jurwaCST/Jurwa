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
ports/ports.go                   ReportRepository and IDGenerator interfaces
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository (Sprint 1)
adapters/out/uuid/               UUID v4 generator
cmd/server/main.go               wires everything and starts the server
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9001` | gRPC port |
| `SERVICE_NAME` | `report-submission` | Name written in every log line |

Storage is in memory in Sprint 1, so reports are lost on restart. Sprint 2 adds a MongoDB adapter (`DATABASE_URL`) and publishes `events:report.submitted` on Redis Streams.

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

Coverage: 100% on `domain`, `application` and the adapters.
