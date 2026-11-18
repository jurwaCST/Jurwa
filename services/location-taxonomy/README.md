# location-taxonomy

Owner: Sonam Choki. gRPC port 9004. Contract: `proto/location-taxonomy/v1/location-taxonomy.proto`.

## What it does

Holds the list of valid locations for each category. The PWA uses it to fill the location drop-downs, and report-submission uses it to check that a submitted location exists. It never stores reports and never routes them.

## Location hierarchy

Each location is one row (category, building, wing, area), like the `locations` table in `docs/design/schema/location-taxonomy.sql`.

| Category | Shape | Example |
|---|---|---|
| `HOSTEL` | building, wing (the floor), area | HC / Floor 1 / Left wing bathroom |
| `ACADEMIC_BLOCK` | building and area, no wing | Academic Block / LH1 |
| `OTHERS` | flat list of areas under building `Campus`, never a wing | Campus / Library |

**The location list** (`adapters/out/memory/sample.go`) holds the real CST places from `docs/design/real-locations.md` (SCRUM-146, loaded in SCRUM-150): 552 places.

| Category | What | Places |
|---|---|---|
| `HOSTEL` | HA, HB, HC, HD, HE, HF, NK, RKA, RKB. Wing is the floor (`Floor 1`...); each floor has its rooms (`Room 101`...) and `Left/Right wing hallway`, `Left/Right wing bathroom` | 525 |
| `ACADEMIC_BLOCK` | Building `Academic Block`, no wing: CR1-5, IT1-5, Lab1-5, LH1-5 | 20 |
| `OTHERS` | Building `Campus`: Library, Mess, Football ground, Basketball court (indoor), Basketball court (outdoor), Volleyball court, Gym | 7 |

The list is built in code from the floor plans (rooms per floor), so a new room is one change to the `hostels` table there. Each place gets a fixed name-based id (UUID version 5 of category and location), the same after every restart. The first sample rows (Hostel A, Academic Block 1, Wi-Fi, Estate, Canteen...) are kept switched off with their old ids, so old reports that name them still make sense.

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
| Case and spaces | Ignored, so `" hc "` matches `"HC"` |
| Inactive rows | Kept in the list but never listed or matched |
| Reasons | `unknown building`, `unknown wing for this building`, `unknown area for this building` |

## Layout

```
domain/location.go               Category, Location, Entry, the hierarchy and validation rules
application/service.go           ListLocations and ValidateLocation use cases
ports/ports.go                   LocationRepository interface
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository and the location list (sample.go)
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

grpcurl -plaintext -d '{"category":"HOSTEL","location":{"building":"HC","wing":"Floor 1","area":"Room 101"}}' \
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
