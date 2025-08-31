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

test-verbose:
	CGO_ENABLED=1 go test -v ./...

test-coverage:
	CGO_ENABLED=1 go test -cover ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

# Simulation commands
simulation:
	@echo "🚀 Starting smoke test simulation..."
	@SIMULATION_DURATION=$$(SIMULATION_DURATION) \
	 SIMULATION_USERS=$$(SIMULATION_USERS) \
	 SIMULATION_MAX_REQUESTS=$$(SIMULATION_MAX_REQUESTS) \
	 PINJOL_URL=$$(PINJOL_URL) \
	 CGO_ENABLED=1 go run ./cmd/smoke/smoke_simulation.go

simulation-5m:
	@echo "🚀 Starting 30-second simulation with 3 users..."
	@SIMULATION_DURATION=30s SIMULATION_USERS=3 SIMULATION_MAX_REQUESTS=20 PINJOL_URL=http://localhost:8081 CGO_ENABLED=1 go run ./cmd/smoke/smoke_simulation.go

simulation-30m:
	@echo "🚀 Starting 30-minute simulation with 5 users..."
	@SIMULATION_DURATION=30m SIMULATION_USERS=5 SIMULATION_MAX_REQUESTS=50 PINJOL_URL=http://localhost:8081 CGO_ENABLED=1 go run ./cmd/smoke/smoke_simulation.go

simulation-1h:
	@echo "🚀 Starting 1-hour simulation with 10 users..."
	@SIMULATION_DURATION=1h SIMULATION_USERS=10 SIMULATION_MAX_REQUESTS=100 PINJOL_URL=http://localhost:8081 CGO_ENABLED=1 go run ./cmd/smoke/smoke_simulation.go

simulation-custom:
	@echo "🚀 Starting custom simulation..."
	@echo "Usage: make simulation SIMULATION_DURATION=10m SIMULATION_USERS=5 SIMULATION_MAX_REQUESTS=30 PINJOL_URL=http://localhost:8081"
	@SIMULATION_DURATION=$$(SIMULATION_DURATION) \
	 SIMULATION_USERS=$$(SIMULATION_USERS) \
	 SIMULATION_MAX_REQUESTS=$$(SIMULATION_MAX_REQUESTS) \
	 PINJOL_URL=$$(PINJOL_URL) \
	 CGO_ENABLED=1 go run ./cmd/smoke/smoke_simulation.go

# Alternative simulation using wrapper script
simulation-script:
	@echo "🚀 Starting simulation using wrapper script..."
	@./run_simulation.sh $(SIMULATION_ARGS)

simulation-script-5m:
	@echo "🚀 Starting 5-minute simulation..."
	@./run_simulation.sh -d 5m -u 3 -r 20

simulation-script-30m:
	@echo "🚀 Starting 30-minute simulation..."
	@./run_simulation.sh -d 30m -u 5 -r 50

simulation-script-1h:
	@echo "🚀 Starting 1-hour simulation..."
	@./run_simulation.sh -d 1h -u 10 -r 100

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

# Monitoring commands
monitoring-start:
	./scripts/monitoring.sh start

monitoring-stop:
	./scripts/monitoring.sh stop

monitoring-restart:
	./scripts/monitoring.sh restart

monitoring-status:
	./scripts/monitoring.sh status

monitoring-logs:
	./scripts/monitoring.sh logs $(service)

monitoring-test:
	./scripts/monitoring.sh test

monitoring-clean:
	./scripts/monitoring.sh clean

# Log aggregation commands
logs-test:
	./scripts/test-logs.sh

logs-query:
	@echo "Querying recent logs..."
	curl -s -G "http://localhost:3100/loki/api/v1/query" \
		--data-urlencode 'query={job="pinjol"}' \
		--data-urlencode 'limit=10' | jq '.data.result[0].values' 2>/dev/null || echo "No logs found or Loki not running"

logs-errors:
	@echo "Querying error logs..."
	curl -s -G "http://localhost:3100/loki/api/v1/query" \
		--data-urlencode 'query={job="pinjol", level="error"}' \
		--data-urlencode 'limit=10' | jq '.data.result[0].values' 2>/dev/null || echo "No error logs found"

# Full development environment
dev-full: dev-setup monitoring-start
	@echo "🚀 Full development environment started!"
	@echo "📊 Grafana: http://localhost:3000"
	@echo "🔍 Loki: http://localhost:3100"
	@echo "📈 Prometheus: http://localhost:9090"
	@echo "🟢 Pinjol App: http://localhost:8080"

dev-stop:
	-./scripts/monitoring.sh stop
	-docker compose down

# Health checks
health-check:
	@echo "🔍 Health Check Results:"
	@curl -s http://localhost:8080/healthz | grep -q "ok" && echo "✅ Pinjol App: Healthy" || echo "❌ Pinjol App: Unhealthy"
	@curl -s http://localhost:3100/ready | grep -q "ready" && echo "✅ Loki: Healthy" || echo "❌ Loki: Unhealthy"
	@curl -s http://localhost:9090/-/healthy >/dev/null && echo "✅ Prometheus: Healthy" || echo "❌ Prometheus: Unhealthy"
	@curl -s http://localhost:3000/api/health >/dev/null && echo "✅ Grafana: Healthy" || echo "❌ Grafana: Unhealthy"

# Development setup
dev-setup:
	go mod tidy
	go mod download

# Production build (includes all optimizations)
prod-build: lint test docker-build

.PHONY: run run-cli cli-ontime cli-skip2 cli-fullpay test test-unit test-api test-verbose test-coverage test-race lint fmt vet docker-build docker-push docker-run compose compose-detached compose-down compose-logs db-init db-migrate build build-static clean dev-setup prod-build monitoring-start monitoring-stop monitoring-restart monitoring-status monitoring-logs monitoring-test monitoring-clean logs-test logs-query logs-errors dev-full dev-stop health-check simulation simulation-5m simulation-30m simulation-1h simulation-custom simulation-script simulation-script-5m simulation-script-30m simulation-script-1h
