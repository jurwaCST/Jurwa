# API routes (api-gateway)

Card: SCRUM-74. Owner: Sonam Zangmo.

The React PWA calls only the api-gateway over HTTP (port 8080, JSON). The gateway checks the JWT, applies rate limits, then calls the owning service over gRPC. The gateway has no business logic and no business RPCs of its own.

## Rules

1. Every route except `GET /health`, `POST /auth/register` and `POST /auth/login` needs `Authorization: Bearer <JWT>`. The gateway calls identity `ValidateToken`, then authorization `CheckPermission` for the user's role.
2. User ids such as `reporter_id`, `uploader_id`, `author_id` and `actor_id` are taken from the JWT, never from the request body.
3. JSON field names are the proto field names in lowerCamelCase (for example `photoId`). Enums are sent as their names (for example `"HOSTEL"`).
4. gRPC errors map to HTTP: INVALID_ARGUMENT 400, UNAUTHENTICATED 401, PERMISSION_DENIED 403, NOT_FOUND 404, ALREADY_EXISTS 409, ABORTED 409, FAILED_PRECONDITION 422, RESOURCE_EXHAUSTED 429, UNAVAILABLE 503, DEADLINE_EXCEEDED 504, anything else 500.
5. Rate limit: 20 requests a second per client IP (Sprint 2: in-memory token bucket in the gateway, `RATE_LIMIT_PER_SECOND`). Sprint 3 moves the counters to Redis (`ratelimit:<user or ip>:<minute>`) so all gateway pods share them.

## Routes

| Method and path | Service | RPC | Roles |
|---|---|---|---|
| `GET /health` | api-gateway | none (gateway health) | public |
| `POST /auth/register` | identity | Register | public |
| `POST /auth/login` | identity | Login | public |
| `GET /me` | identity | ValidateToken | any signed-in user |
| `GET /categories` | category-taxonomy | ListCategories | any signed-in user |
| `GET /categories/{category}/locations` | location-taxonomy | ListLocations | any signed-in user |
| `POST /photos` | photo-media | UploadPhoto | Student |
| `GET /photos/{photoId}` | photo-media | GetPhoto | Student (own), Category Head, Staff, System Administrator |
| `POST /reports` | report-submission | SubmitReport | Student |
| `GET /reports/{reportId}` | report-submission | GetReport | Student (own), Category Head, Staff, System Administrator |
| `GET /me/reports` | report-history | ListReportsByReporter | Student |
| `GET /reports/{reportId}/timeline` | report-history | GetReportTimeline | Student (own), Category Head, Staff, System Administrator |
| `GET /reports/{reportId}/status` | status-lifecycle | GetStatus | Student (own), Category Head, Staff, System Administrator |
| `POST /reports/{reportId}/status` | status-lifecycle | ChangeStatus | Category Head, Staff (assignee) |
| `POST /reports/{reportId}/reassign` | reassignment | ReassignReport | Category Head, Staff (assignee) |
| `GET /reports/{reportId}/assignments` | reassignment | ListReassignments | Category Head, Staff, System Administrator |
| `POST /reports/{reportId}/notes` | internal-notes | AddNote | Category Head, Staff |
| `GET /reports/{reportId}/notes` | internal-notes | ListNotes | Category Head, Staff, System Administrator |
| `POST /reports/{reportId}/feedback` | feedback | SubmitFeedback | Student (own report, after RESOLVED) |
| `GET /reports/{reportId}/feedback` | feedback | GetFeedback | Student (own), Category Head, System Administrator |
| `GET /reports/{reportId}/audit` | audit-log | ListByReport | System Administrator |
| `GET /me/notifications` | notification | ListNotifications | any signed-in user |
| `GET /departments` | department-directory | ListDepartments | Category Head, System Administrator |
| `GET /positions` | department-directory | ListPositions (held only) | Category Head, Staff (the reassign picker) |
| `GET /admin/routing-rules` | routing-rules | ListRules | System Administrator |
| `PUT /admin/routing-rules` | routing-rules | UpsertRule | System Administrator |
| `GET /admin/sla-policies/{category}` | sla-policy | GetPolicy | System Administrator |
| `PUT /admin/sla-policies/{category}` | sla-policy | UpsertPolicy | System Administrator |
| `GET /dashboard/summary` | analytics | GetSummary | System Administrator, College President |
| `GET /dashboard/categories` | analytics | GetCategoryBreakdown | System Administrator, College President |

## Not exposed to the PWA

These RPCs are called only between services or by the serverless functions:

| RPC | Called by |
|---|---|
| routing-engine RouteReport | after report.submitted (event) |
| routing-rules GetRule | routing-engine |
| department-directory GetHoD, GetStaff | routing-engine, reassignment |
| location-taxonomy ValidateLocation | report-submission |
| sla-tracking StartTimer, StopTimer, GetTimer, ListDueTimers | status-lifecycle events, sla-escalation-checker |
| audit-log Append | every service (or from events) |
| notification SendNotification | event consumers, digest-dispatcher |
| authorization CheckPermission | api-gateway and services |
| identity ValidateToken | api-gateway |

## Sprint 1 thin slice

Only `GET /health` and one forwarded route (`GET /categories` or `POST /reports`) are built in Sprint 1 (SCRUM-38). JWT checks and rate limiting come in Sprint 2.

## Sprint 2 slice

SCRUM-38 adds the JWT check (identity `ValidateToken`), the rate limit, `pkg/resilience` on every client and these routes: `GET /health`, `POST /auth/register`, `POST /auth/login`, `GET /me`, `GET /categories`, `GET /categories/{category}/locations`, `POST /reports` and `GET /reports/{reportId}`. Role checks call authorization `CheckPermission` (SCRUM-14) for every signed-in route. See `services/api-gateway/README.md`.
