# internal-notes

Owner: Sonam Choki. gRPC port 9005. Contract: `proto/internal-notes/v1/internal-notes.proto`.

## What it does

Stores staff-only notes on a report, for example "Plumber booked for Monday". Notes are internal: students never see them (FR-21). It never changes a report or its status.

## RPCs

| RPC | Does | Errors |
|---|---|---|
| `AddNote` | Stores one note and returns it | `PERMISSION_DENIED` for any role except Category Head and Staff, `INVALID_ARGUMENT` for a bad id or a blank or too long body |
| `ListNotes` | Notes of one report, oldest first (empty when the report has none) | `PERMISSION_DENIED` for a student, the President or a missing role, `INVALID_ARGUMENT` when the report id is not a UUID |

## Caller role

The proto messages have no role field, so the caller's role travels in the gRPC metadata key `x-user-role`, with a `Role` enum name from `proto/common/v1` (for example `STAFF`). api-gateway sets it from the JWT, the same way it sets `author_id`. A call without it is denied.

| Role | Add | List |
|---|---|---|
| `STAFF`, `CATEGORY_HEAD` | yes | yes |
| `SYSTEM_ADMIN` | no | yes |
| `STUDENT`, `PRESIDENT`, missing | no | no |

These are the same rules as `docs/design/api-routes.md` and the authorization policy, so api-gateway and this service agree.

The role is checked before anything else, so a student learns nothing about the note rules.

Rules (in `domain/note.go`, tested in `domain/note_test.go`):

| Rule | Detail |
|---|---|
| Report id, author id | Required UUIDs |
| Body | Not blank; at most 1000 characters after trimming spaces. Characters are counted, not bytes, so Dzongkha text gets the same limit |
| Order | Oldest first; notes with the same time keep the order they were added |

## Layout

```
domain/note.go                   Note, roles, the rules and errors
application/service.go           AddNote and ListNotes use cases
ports/ports.go                   NoteRepository interface
adapters/in/grpc/handler.go      gRPC handler: role from metadata, protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository
cmd/server/main.go               wires everything and starts the server
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9005` | gRPC port |
| `SERVICE_NAME` | `internal-notes` | Name written in every log line |

Storage is in memory in Sprint 2. Sprint 3 adds a MongoDB adapter (collection `notes`, `docs/design/database-schema.md`).

## Run

```bash
make run SERVICE=internal-notes
```

In a second terminal:

```bash
grpcurl -plaintext localhost:9005 grpc.health.v1.Health/Check

grpcurl -plaintext -H 'x-user-role: STAFF' \
  -d '{"report_id":"6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f","author_id":"1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d","body":"Plumber booked for Monday."}' \
  localhost:9005 jurwa.internalnotes.v1.InternalNotesService/AddNote

grpcurl -plaintext -H 'x-user-role: CATEGORY_HEAD' -d '{"report_id":"6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"}' \
  localhost:9005 jurwa.internalnotes.v1.InternalNotesService/ListNotes
```

Expected: health `SERVING`; AddNote returns the note; ListNotes returns it. With `-H 'x-user-role: STUDENT'` both return `PermissionDenied`; with `SYSTEM_ADMIN` AddNote returns `PermissionDenied` and ListNotes works.

Docker:

```bash
make docker-build SERVICE=internal-notes
docker run --rm -p 9005:9005 jurwa/internal-notes
```

## Test

```bash
make test
go test ./services/internal-notes/... -cover
```

Coverage: 100% on `domain`, `application` and both adapters. Evidence: `docs/evidence/sonam-choki/scrum-21-tests.txt` and `scrum-21-grpcurl.txt`; the add rule for System Admin (SCRUM-116) in `scrum-116-tests.txt` and `scrum-116-grpcurl.txt`.
