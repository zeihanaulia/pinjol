APP=pinjol
DOCKER_IMAGE=pinjol
DOCKER_TAG?=latest

# Development commands
run:
	CGO_ENABLED=1 go run -ldflags "-X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" .

run-cli:
	CGO_ENABLED=1 go run -ldflags "-X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" . cli

cli-ontime:
	CGO_ENABLED=1 go run . cli --scenario ontime --repeat 10 --verbose

cli-skip2:
	CGO_ENABLED=1 go run . cli --scenario skip2 --verbose

cli-fullpay:
	CGO_ENABLED=1 go run . cli --scenario fullpay --verbose

# Testing commands
test:
	CGO_ENABLED=1 go test ./...

test-unit:
	CGO_ENABLED=1 go test ./tests/unit/...

test-api:
	CGO_ENABLED=1 go test ./tests/api/...

test-verbose:
	CGO_ENABLED=1 go test -v ./...

test-coverage:
	CGO_ENABLED=1 go test -cover ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

# Code quality
lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	CGO_ENABLED=1 go vet ./...

# Docker commands (optimized)
docker-build:
	DOCKER_BUILDKIT=1 docker build \
		--build-arg GIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.

docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

docker-run:
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose
compose:
	docker compose up --build

compose-detached:
	docker compose up -d --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f

# Database commands
db-init:
	CGO_ENABLED=1 go run . db-init

db-migrate:
	CGO_ENABLED=1 go run . db-migrate

# Build commands
build:
	CGO_ENABLED=1 go build \
		-ldflags "-s -w -X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" \
		-o $(APP) \
		.

build-static:
	CGO_ENABLED=1 GOOS=linux go build \
		-a \
		-installsuffix cgo \
		-ldflags="-s -w -extldflags '-static' -X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" \
		-o $(APP)-static \
		.

# Cleanup
clean:
	go clean
	rm -f $(APP)
	rm -f $(APP)-static
	rm -f *.db
	docker system prune -f

# Development setup
dev-setup:
	go mod tidy
	go mod download

# Production build (includes all optimizations)
prod-build: lint test docker-build

.PHONY: run run-cli cli-ontime cli-skip2 cli-fullpay test test-unit test-api test-verbose test-coverage test-race lint fmt vet docker-build docker-push docker-run compose compose-detached compose-down compose-logs db-init db-migrate build build-static clean dev-setup prod-build
