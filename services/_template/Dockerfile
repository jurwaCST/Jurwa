# Build from the repo root, because the whole repo is one Go module:
#   docker build -f services/__SERVICE__/Dockerfile --build-arg SERVICE=__SERVICE__ -t jurwa/__SERVICE__ .
# or: make docker-build SERVICE=__SERVICE__

# Stage 1: compile a static binary.
FROM golang:1.25-alpine AS build
ARG SERVICE
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY proto/ proto/
COPY services/${SERVICE}/ services/${SERVICE}/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./services/${SERVICE}/cmd/server

# Stage 2: run it on a tiny image with no shell, as a non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
ARG SERVICE
ENV SERVICE_NAME=${SERVICE}
COPY --from=build /out/server /server
EXPOSE __PORT__
USER nonroot:nonroot
ENTRYPOINT ["/server"]
