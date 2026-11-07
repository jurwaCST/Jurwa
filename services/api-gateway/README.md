# api-gateway

## What it does

The single entry point for the React PWA. It serves **HTTP JSON on port 8080**, keeps the standard **gRPC health check on port 9010**, and forwards each route to the owning service over gRPC. It has no business logic of its own. The full route list and rules are in `docs/design/api-routes.md`.

For every request it:

1. **Rate limits**: at most 20 requests a second per client IP (in-memory token bucket, `RATE_LIMIT_PER_SECOND`). The 21st gets `429` with `Retry-After: 1`. Redis counters replace this in Sprint 3 so several gateway pods share one limit.
2. **Checks the JWT** on every route except `/health`, `/auth/register` and `/auth/login`: `Authorization: Bearer <token>` is sent to identity `ValidateToken`. A missing or invalid token is `401`.
3. **Checks the role** for the route's action with authorization `CheckPermission` (Gayley's service, `AUTHORIZATION_ADDR`). A denial is `403`. If authorization cannot be reached the request is refused with `503`, never let through. A student may read only their own report (checked in the gateway, which has the report).
   For internal-notes, whose messages have no role field, the gateway sends the allowed role as the gRPC metadata `x-user-role` (for example `STAFF`), and `author_id` from the JWT. The body of `POST .../notes` is `{"body": "..."}`; the report id comes from the path.
4. **Forwards** the call. User ids such as `reporter_id` come from the JWT, never from the body. Bodies are decoded strictly: bad JSON, an unknown field or an unknown enum value is `400`.
5. **Maps gRPC errors to HTTP**, and never passes internal error text to the browser.

Every gRPC client goes through `pkg/resilience` (2 s timeout, up to 3 attempts for Unavailable or DeadlineExceeded, one circuit breaker per service).

## Routes

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
| `POST /reports/{reportId}/notes` | internal-notes AddNote (returns 201) | Category Head, Staff |
| `GET /reports/{reportId}/notes` | internal-notes ListNotes | Category Head, Staff, System Administrator |
| `GET /me/reports` | report-history ListReportsByReporter (`?pageSize=` 1 to 100, `?pageToken=`) | Student; the reporter id comes from the JWT |
| `GET /reports/{reportId}/timeline` | report-history GetReportTimeline | Student (own, checked with report-submission), Category Head, Staff, System Administrator |
| `POST /reports/{reportId}/status` | status-lifecycle ChangeStatus (body `{"toStatus": "IN_PROGRESS", "expectedVersion": 1, "reason": "..."}`) | Category Head, Staff; the actor id comes from the JWT |
| `GET /reports/{reportId}/status` | status-lifecycle GetStatus | Student (own, checked with report-submission), Category Head, Staff, System Administrator |
| `POST /photos` | photo-media UploadPhoto (returns 201 `{"photoId": "..."}`) | Student; the uploader id comes from the JWT |
| `GET /photos/{photoId}` | photo-media GetPhoto (returns the raw image) | Student (own), Category Head, Staff, System Administrator |

The other routes in `api-routes.md` are added as their services are merged.

## Status changes

The report id comes from the path and the actor id from the JWT; both are ignored in the body. The actor is never empty, because an empty actor means the system, which is the only one allowed to set `ESCALATED`. status-lifecycle checks the flow (OPEN, ASSIGNED, IN_PROGRESS, RESOLVED): a change the flow does not allow is 422, and a stale `expectedVersion` (someone changed it first) is 409. `api-routes.md` says "Staff (assignee)": the gateway checks the role only for now, because no service knows the assignee yet; the assignee check follows when report-history has it (SCRUM-131).

## Photos

`POST /photos` takes `multipart/form-data` with the file in the field `photo`. The gateway passes the part's `Content-Type` and file name to photo-media as they are and does no type check: photo-media checks the real bytes (JPG or PNG, at most 5 MB) and its errors come back as 400. Other form fields are ignored, including any uploader id. The PWA then sends the `photoId` with `POST /reports`.

| Photo size | Who refuses it | Response |
|---|---|---|
| up to 5 MB | nobody | 201 `{"photoId": "..."}` |
| over 5 MB, body up to 6 MB | photo-media | 400 `{"error": "invalid photo: photo is larger than 5 MB"}` |
| body over 6 MB | the gateway, before calling photo-media | 400, the same message |

Only `POST /photos` gets the 6 MB body limit; every other route keeps 64 KB. The photo-media gRPC connection allows 6 MB messages in both directions (the grpc-go default for a received message is 4 MB, which would block `GET /photos` for larger photos); photo-media's server accepts 6 MB too.

`GET /photos/{photoId}` sends the image bytes with `Content-Type: image/jpeg` or `image/png` (anything else is sent as `application/octet-stream`), `X-Content-Type-Options: nosniff` and `Cache-Control: private, no-store`. The route needs the JWT, so the PWA fetches it with the `Authorization` header and shows it from a blob URL (`URL.createObjectURL`). A student can only fetch their own photos.

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
| `AUTHORIZATION_ADDR` | `localhost:9017` | authorization |
| `INTERNAL_NOTES_ADDR` | `localhost:9005` | internal-notes |
| `REPORT_HISTORY_ADDR` | `localhost:9003` | report-history |
| `PHOTO_MEDIA_ADDR` | `localhost:9002` | photo-media |
| `STATUS_LIFECYCLE_ADDR` | `localhost:9011` | status-lifecycle |

## Run

Start the services the routes need, each in its own terminal (identity needs a `JWT_SECRET`, never committed):

```bash
JWT_SECRET=<any long random string> make run SERVICE=identity
make run SERVICE=authorization
make run SERVICE=category-taxonomy
make run SERVICE=report-submission
make run SERVICE=internal-notes
make run SERVICE=report-history
make run SERVICE=photo-media
make run SERVICE=status-lifecycle
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
