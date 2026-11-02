# api-gateway

## What it does

The single entry point for the React PWA. It serves **HTTP JSON on port 8080**, keeps the standard **gRPC health check on port 9010**, and forwards each route to the owning service over gRPC. It has no business logic of its own. The full route list and rules are in `docs/design/api-routes.md`.

For every request it:

1. **Rate limits**: at most 20 requests a second per client IP (in-memory token bucket, `RATE_LIMIT_PER_SECOND`). The 21st gets `429` with `Retry-After: 1`. Redis counters replace this in Sprint 3 so several gateway pods share one limit.
2. **Checks the JWT** on every route except `/health`, `/auth/register` and `/auth/login`: `Authorization: Bearer <token>` is sent to identity `ValidateToken`. A missing or invalid token is `401`.
3. **Checks the role** for the route's action. Until Gayley's authorization service is merged, `adapters/out/staticauthz` applies the role table from `api-routes.md` (TODO in `cmd/server/main.go`). A student may read only their own report.
4. **Forwards** the call. User ids such as `reporter_id` come from the JWT, never from the body. Bodies are decoded strictly: bad JSON, an unknown field or an unknown enum value is `400`.
5. **Maps gRPC errors to HTTP**, and never passes internal error text to the browser.

Every gRPC client goes through `pkg/resilience` (2 s timeout, up to 3 attempts for Unavailable or DeadlineExceeded, one circuit breaker per service).

## Routes in Sprint 2

| Method and path | Service, RPC | Roles |
|---|---|---|
| `GET /health` | none | public |
| `POST /auth/register` | identity Register (returns 201) | public |
| `POST /auth/login` | identity Login | public |
| `GET /me` | identity ValidateToken | any signed-in user |
| `GET /categories` | category-taxonomy ListCategories | any signed-in user |
| `GET /categories/{category}/locations` | location-taxonomy ListLocations | any signed-in user |
| `POST /reports` | report-submission SubmitReport (returns 201) | Student |
| `GET /reports/{reportId}` | report-submission GetReport | Student (own), Category Head, Staff, System Administrator |

The other routes in `api-routes.md` are added as their services are merged.

## Errors

| gRPC code | HTTP |
|---|---|
| INVALID_ARGUMENT | 400 |
| UNAUTHENTICATED | 401 |
| PERMISSION_DENIED | 403 |
| NOT_FOUND | 404 |
| ALREADY_EXISTS, ABORTED | 409 |
| FAILED_PRECONDITION | 422 |
| RESOURCE_EXHAUSTED | 429 |
| UNAVAILABLE | 503 |
| DEADLINE_EXCEEDED | 504 |
| anything else | 500 (`"internal error"`, details only in the gateway log) |

Error bodies are `{"error": "<message>"}`.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `HTTP_PORT` | `8080` | HTTP port for the PWA |
| `PORT` | `9010` | gRPC health check port |
| `SERVICE_NAME` | `api-gateway` | Name written in every log line |
| `RATE_LIMIT_PER_SECOND` | `20` | Requests a second per client IP |
| `IDENTITY_ADDR` | `localhost:9016` | identity |
| `CATEGORY_TAXONOMY_ADDR` | `localhost:9009` | category-taxonomy |
| `LOCATION_TAXONOMY_ADDR` | `localhost:9004` | location-taxonomy |
| `REPORT_SUBMISSION_ADDR` | `localhost:9001` | report-submission |

`AUTHORIZATION_ADDR` (`localhost:9017`) is added with the authorization client.

## Run

Start the services the routes need, each in its own terminal (identity needs a `JWT_SECRET`, never committed):

```bash
JWT_SECRET=<any long random string> make run SERVICE=identity
make run SERVICE=category-taxonomy
make run SERVICE=report-submission
make run SERVICE=api-gateway
```

Then:

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/auth/register -d '{"email":"student@rub.edu.bt","fullName":"Test Student","password":"secret123"}'
curl -i -X POST http://localhost:8080/auth/login -d '{"email":"student@rub.edu.bt","password":"secret123"}'
TOKEN=<accessToken from the login reply>
curl -i http://localhost:8080/categories -H "Authorization: Bearer $TOKEN"
curl -i -X POST http://localhost:8080/reports -H "Authorization: Bearer $TOKEN" -d '{"category":"HOSTEL","location":{"building":"Hostel B"},"description":"Tap leaking in bathroom 2"}'
grpcurl -plaintext localhost:9010 grpc.health.v1.Health/Check
```

Docker:

```bash
make docker-build SERVICE=api-gateway
docker run --rm -p 8080:8080 -p 9010:9010 -e IDENTITY_ADDR=host.docker.internal:9016 jurwa/api-gateway
```

## Test

```bash
make test
go test ./services/api-gateway/... -cover
```

Tests were written first (commit "SCRUM-38 tests first"). The HTTP tests use `httptest` with fakes for every service, and the rate limiter uses a fake clock.
