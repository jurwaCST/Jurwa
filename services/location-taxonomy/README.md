# location-taxonomy

Owner: Sonam Choki. gRPC port 9004. Contract: `proto/location-taxonomy/v1/location-taxonomy.proto`.

## What it does

Holds the list of valid locations (building, wing, area) for each category. The PWA uses it to fill the location drop-downs, and report-submission uses it to check that a submitted location exists. It never stores reports and never routes them.

## RPCs

| RPC | Does | Errors |
|---|---|---|
| `ListLocations` | Returns the active locations of one category, in list order | `INVALID_ARGUMENT` when the category is missing or unknown |
| `ValidateLocation` | Returns `valid: true`, or `valid: false` with a `reason` | `INVALID_ARGUMENT` when the category is missing or unknown, or the building is blank |

An unknown location is a normal answer, not an error, so report-submission can show the reason to the student.

Rules (in `domain/location.go`, tested in `domain/location_test.go`):

| Rule | Detail |
|---|---|
| Category | Required: `HOSTEL`, `ACADEMIC_BLOCK` or `OTHERS` |
| Exact match | A location is valid only when building, wing and area all match one active entry. "Hostel B" alone is valid only because the list has a building-only "Hostel B" entry |
| Case and spaces | Ignored, so `" hostel b "` matches `"Hostel B"` |
| Inactive entries | Kept in the list but never listed or matched |
| Reasons | `unknown building`, or `unknown wing or area for this building` |

## Layout

```
domain/location.go               Category, Location, Entry, the rules above
application/service.go           ListLocations and ValidateLocation use cases
ports/ports.go                   LocationRepository interface
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository and the seed list (seed.go)
cmd/server/main.go               wires everything and starts the server
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9004` | gRPC port |
| `SERVICE_NAME` | `location-taxonomy` | Name written in every log line |

The list is in memory and comes from `adapters/out/memory/seed.go`. The names there are examples that match the wireframes; replace them with the real CST buildings. A PostgreSQL adapter for `docs/design/schema/location-taxonomy.sql` (`DATABASE_URL`) can replace the in-memory one later.

## Run

```bash
make run SERVICE=location-taxonomy
```

In a second terminal:

```bash
grpcurl -plaintext localhost:9004 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"category":"HOSTEL"}' \
  localhost:9004 jurwa.locationtaxonomy.v1.LocationTaxonomyService/ListLocations

grpcurl -plaintext -d '{"category":"HOSTEL","location":{"building":"Hostel B","wing":"Wing 2","area":"Bathroom 2"}}' \
  localhost:9004 jurwa.locationtaxonomy.v1.LocationTaxonomyService/ValidateLocation
```

Expected: health `SERVING`; ListLocations returns the hostel entries; ValidateLocation returns `"valid": true`. For an unknown location grpcurl prints only `"reason"`, because protobuf JSON leaves out `valid` when it is `false`.

Docker:

```bash
make docker-build SERVICE=location-taxonomy
docker run --rm -p 9004:9004 jurwa/location-taxonomy
```

## Test

```bash
make test
go test ./services/location-taxonomy/... -cover
```

Coverage: 100% on `domain`, `application` and the adapters.
