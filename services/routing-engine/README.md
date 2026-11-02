# routing-engine

## What it does

Decides who receives a new report, using **only the category and the building** (decision 1, FR-11). There is no description input anywhere: `TestRequestHasNoDescription` fails if one is added.

1. Checks the request: the category must be HOSTEL, ACADEMIC_BLOCK or OTHERS, and the building must not be blank.
2. Asks routing-rules `GetRule` for the matching rule (for example HOSTEL to `DSA`).
3. Asks department-directory `GetStaff` who holds that position.
4. Returns the holder's user id, the position and the rule id. When `report_id` is set it publishes one `events:report.routed` event. An empty `report_id` is a dry run and publishes nothing.

When routing-rules or department-directory cannot be reached, the call fails with UNAVAILABLE. It never guesses a Category Head.

The service is stateless (database-schema.md, row 7), so it has no repository and no `adapters/out/memory`. The outbound adapters are:

| Adapter | Port (interface) | What it does |
|---|---|---|
| `adapters/out/rulesclient` | `ports.RuleLookup` | gRPC client for routing-rules `GetRule` |
| `adapters/out/directoryclient` | `ports.HolderLookup` | gRPC client for department-directory `GetStaff` |
| `adapters/out/logpublisher` | `ports.EventPublisher` | Writes `report.routed` as a JSON log line. Redis Streams replaces it in Sprint 3 |

Tests use fakes for all three ports. Both gRPC connections use `pkg/resilience` (SCRUM-91): a 2 second timeout per attempt, up to 3 attempts for Unavailable or DeadlineExceeded, and a circuit breaker per downstream service that opens after 5 failures in a row for 10 seconds.

## RPCs

Contract: `proto/routing-engine/v1/routing-engine.proto`, service `jurwa.routingengine.v1.RoutingEngineService`.

| RPC | What it does | Errors |
|---|---|---|
| `RouteReport` | Category Head for a category and location | INVALID_ARGUMENT for an unspecified category or blank building, NOT_FOUND when no rule matches, FAILED_PRECONDITION when the position has no holder, UNAVAILABLE when routing-rules or department-directory is down |

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9007` | gRPC port |
| `SERVICE_NAME` | `routing-engine` | Name written in every log line |
| `ROUTING_RULES_ADDR` | `localhost:9006` | routing-rules gRPC address |
| `DEPARTMENT_DIRECTORY_ADDR` | `localhost:9008` | department-directory gRPC address |

## Run

Three terminals:

```bash
make run SERVICE=routing-rules          # 9006
make run SERVICE=department-directory   # 9008
make run SERVICE=routing-engine         # 9007
```

Then:

```bash
grpcurl -plaintext localhost:9007 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:9007 list
grpcurl -plaintext -d '{"report_id":"6f1c2a9e-1b7d-4c55-9a43-2f1e0b9d7c11","category":"HOSTEL","location":{"building":"Hostel B"}}' localhost:9007 jurwa.routingengine.v1.RoutingEngineService/RouteReport
```

Docker:

```bash
make docker-build SERVICE=routing-engine
docker run --rm -p 9007:9007 -e ROUTING_RULES_ADDR=host.docker.internal:9006 -e DEPARTMENT_DIRECTORY_ADDR=host.docker.internal:9008 jurwa/routing-engine
```

## Test

```bash
make test
go test ./services/routing-engine/... -cover
```

The domain and the use case are written test-first (commit "SCRUM-90 tests first") and have 100% statement coverage.
