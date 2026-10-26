# Stage 1: build the Go program
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /hello .

# Stage 2: copy only the finished program into a tiny image
FROM alpine:3.20
COPY --from=build /hello /hello
EXPOSE 50051
ENTRYPOINT ["/hello"]
