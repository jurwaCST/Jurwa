# Build from the repo root, because the whole repo is one Go module:
#   docker build -f functions/sla-escalation-checker/Dockerfile -t jurwa/sla-escalation-checker .

# Stage 1: compile a static binary.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY proto/ proto/
COPY pkg/ pkg/
COPY functions/sla-escalation-checker/ functions/sla-escalation-checker/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/function ./functions/sla-escalation-checker

# Stage 2: run it on a tiny image with no shell, as a non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/function /function
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/function"]
