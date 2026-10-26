# Jurwa developer commands. Run from the repo root. "make help" lists them.
# Recipe lines must start with a TAB, not spaces.

PROTO_FILES := $(shell find proto -name '*.proto' -not -path 'proto/gen/*' 2>/dev/null)

.PHONY: help proto test run build docker-build lint new-service check-service template-check

# Go skips folders starting with "_", so services/_template is never in ./...
# These targets do nothing (and pass) until the first real service exists.
HAS_GO := $(shell go list ./... 2>/dev/null | head -1)

help:
	@echo "make proto                          generate Go code from proto/**/*.proto into proto/gen/"
	@echo "make test                           run all unit tests with coverage"
	@echo "make lint                           gofmt check and go vet"
	@echo "make template-check                 build and test a temporary copy of services/_template"
	@echo "make run SERVICE=name               run a service locally"
	@echo "make build SERVICE=name             build a service binary into bin/"
	@echo "make docker-build SERVICE=name      build a service Docker image jurwa/name"
	@echo "make new-service NAME=x PORT=9000   create services/x from the template"

proto:
ifeq ($(strip $(PROTO_FILES)),)
	@echo "No .proto files found in proto/. Nothing to generate."
else
	@mkdir -p proto/gen
	protoc -I proto \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
endif

test:
ifeq ($(HAS_GO),)
	@echo "No Go packages yet. Nothing to test."
else
	go test ./... -cover
endif

lint:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed on:"; echo "$$out"; exit 1; fi
ifneq ($(HAS_GO),)
	go vet ./...
endif

template-check:
	@rm -rf services/template-check
	@bash scripts/new-service.sh template-check 9098 > /dev/null
	@status=0; go vet ./services/template-check/... && go test ./services/template-check/... -cover || status=$$?; \
		rm -rf services/template-check; exit $$status

check-service:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make $(MAKECMDGOALS) SERVICE=name"; exit 1; fi
	@if [ ! -d "services/$(SERVICE)" ]; then echo "services/$(SERVICE) does not exist"; exit 1; fi

run: check-service
	go run ./services/$(SERVICE)/cmd/server

build: check-service
	go build -o bin/$(SERVICE) ./services/$(SERVICE)/cmd/server

docker-build: check-service
	docker build -f services/$(SERVICE)/Dockerfile --build-arg SERVICE=$(SERVICE) -t jurwa/$(SERVICE) .

new-service:
	@if [ -z "$(NAME)" ] || [ -z "$(PORT)" ]; then echo "Usage: make new-service NAME=x PORT=9000"; exit 1; fi
	bash scripts/new-service.sh $(NAME) $(PORT)
