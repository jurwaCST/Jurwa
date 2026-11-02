# photo-media

Owner: Sonam Choki. gRPC port 9002. Contract: `proto/photo-media/v1/photo-media.proto`.

## What it does

Stores the optional photo of a report. The student uploads the photo first and gets a `photo_id`, then sends that id with `SubmitReport`. It never stores reports and never decides who may see a photo; api-gateway checks that.

## RPCs

| RPC | Does | Errors |
|---|---|---|
| `UploadPhoto` | Checks and stores one photo, returns its `photo_id` | `INVALID_ARGUMENT` for a wrong type, empty or too large file, or a missing uploader |
| `GetPhoto` | Returns the bytes, content type, size and uploader of one photo | `INVALID_ARGUMENT` when the id is not a UUID, `NOT_FOUND` when it is unknown |

Rules (in `domain/photo.go`, tested in `domain/photo_test.go`):

| Rule | Detail |
|---|---|
| Content type | `image/jpeg` or `image/png` only (case and spaces ignored) |
| File bytes | Must start with the signature of the stated type: JPEG `FF D8 FF`, PNG `89 50 4E 47`. A GIF renamed to `.png` is rejected |
| Size | 1 byte to 5 MB (5 242 880 bytes). 0 bytes and 5 MB plus 1 byte are rejected |
| Uploader | Required UUID (api-gateway sets it from the JWT) |

The server starts with `grpc.MaxRecvMsgSize(6 MB)`, because gRPC accepts only 4 MB by default. api-gateway must allow the same size.

## Layout

```
domain/photo.go                  Photo, the upload rules and errors
application/service.go           UploadPhoto and GetPhoto use cases
ports/ports.go                   PhotoRepository interface
adapters/in/grpc/handler.go      gRPC handler: protobuf <-> domain, errors -> gRPC codes
adapters/out/memory/             in-memory repository
cmd/server/main.go               wires everything and starts the server
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `9002` | gRPC port |
| `SERVICE_NAME` | `photo-media` | Name written in every log line |

Storage is in memory in Sprint 2, so photos are lost on restart. Sprint 3 adds a MongoDB GridFS adapter (bucket `photos`, `docs/design/database-schema.md`).

## Run

```bash
make run SERVICE=photo-media
```

In a second terminal (grpcurl sends `bytes` fields as base64; `/9j/4AAQ` is the start of a JPEG):

```bash
grpcurl -plaintext localhost:9002 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"uploader_id":"1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d","content_type":"image/jpeg","data":"/9j/4AAQ","file_name":"tap.jpg"}' \
  localhost:9002 jurwa.photomedia.v1.PhotoMediaService/UploadPhoto

grpcurl -plaintext -d '{"photo_id":"<id from the upload>"}' \
  localhost:9002 jurwa.photomedia.v1.PhotoMediaService/GetPhoto
```

Expected: health `SERVING`; UploadPhoto returns a `photoId`; GetPhoto returns the same bytes. A GIF (`"content_type":"image/gif"`) returns `InvalidArgument`.

Docker:

```bash
make docker-build SERVICE=photo-media
docker run --rm -p 9002:9002 jurwa/photo-media
```

## Test

```bash
make test
go test ./services/photo-media/... -cover
```

Coverage: 100% on `domain`, `application` and both adapters. Evidence: `docs/evidence/sonam-choki/scrum-28-tests.txt` and `scrum-28-grpcurl.txt`.
