# location-taxonomy

Owner: Sonam Choki. gRPC port 9004. Contract: `proto/location-taxonomy/v1/location-taxonomy.proto`.

## What it does

Holds the list of valid locations for each category. The PWA uses it to fill the location drop-downs, and report-submission uses it to check that a submitted location exists. It never stores reports and never routes them.

## Location hierarchy

Each location is one row (category, building, wing, area), like the `locations` table in `docs/design/schema/location-taxonomy.sql`.

| Category | Shape | Example |
|---|---|---|
| `HOSTEL` | building, wing, area | Hostel A / Wing 1 / Bathroom 1 |
| `ACADEMIC_BLOCK` | building, wing, area | Academic Block 2 / Ground Floor / Lab 3 |
| `OTHERS` | flat list of areas under building `Campus`, never a wing | Campus / Wi-Fi |

**Sample data.** The design documents name no real CST buildings, so `adapters/out/memory/sample.go` holds a small sample set: two hostels, two academic blocks and the campus-wide areas (Wi-Fi, Estate, Canteen). Replace it with the real names when the team has them.

## RPCs

| RPC | Does | Errors |
|---|---|---|
| `ListLocations` | Returns the active locations of one category, in list order | `INVALID_ARGUMENT` when the category is missing or unknown |
| `ValidateLocation` | Returns `valid: true`, or `valid: false` with a `reason` | `INVALID_ARGUMENT` when the category is missing or unknown, or the building is empty |

An unknown location is a normal answer, not an error, so report-submission can show the reason to the student.

Rules (in `domain/location.go`, tested in `domain/location_test.go`):

| Rule | Detail |
|---|---|
| Category | Required: `HOSTEL`, `ACADEMIC_BLOCK` or `OTHERS` |
| Building | Required |
| Wing and area | Optional, but each one given must belong to the building, and the area to the wing |
| Case and spaces | Ignored, so `" hostel a "` matches `"Hostel A"` |
| Inactive rows | Kept in the list but never listed or matched |
| Reasons | `unknown building`, `unknown wing for this building`, `unknown area for this building` |

## Layout

```
domain/location.go               Category, Location, Entry, the hierarchy and validation rules
application/service.go           ListLocations and ValidateLocation use cases
ports/ports.go                   LocationRepository interface
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository and the sample list (sample.go)
cmd/server/main.go               wires everything and starts the server
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9004` | gRPC port |
| `SERVICE_NAME` | `location-taxonomy` | Name written in every log line |

Storage is in memory in Sprint 2. Sprint 3 adds a PostgreSQL adapter (`DATABASE_URL`) for `docs/design/schema/location-taxonomy.sql`.

## Run

```bash
make run SERVICE=location-taxonomy
```

In a second terminal:

```bash
grpcurl -plaintext localhost:9004 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"category":"HOSTEL"}' \
  localhost:9004 jurwa.locationtaxonomy.v1.LocationTaxonomyService/ListLocations

grpcurl -plaintext -d '{"category":"HOSTEL","location":{"building":"Hostel A","wing":"Wing 1","area":"Bathroom 1"}}' \
  localhost:9004 jurwa.locationtaxonomy.v1.LocationTaxonomyService/ValidateLocation
```

Expected: health `SERVING`; ListLocations returns the hostel rows; ValidateLocation returns `"valid": true`. For an unknown location grpcurl prints only `"reason"`, because protobuf JSON leaves out `valid` when it is `false`.

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

Coverage: 100% on `domain`, `application` and both adapters. Evidence: `docs/evidence/sonam-choki/scrum-23-tests.txt` and `scrum-23-grpcurl.txt`.
