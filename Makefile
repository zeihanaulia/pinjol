APP=pinjol

run:
	go run -ldflags "-X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" .

run-cli:
	go run -ldflags "-X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" . cli

cli-ontime:
	go run . cli --scenario ontime --repeat 10 --verbose

cli-skip2:
	go run . cli --scenario skip2 --verbose

cli-fullpay:
	go run . cli --scenario fullpay --verbose

test:
	go test ./...

test-unit:
	go test ./tests/unit/...

test-api:
	go test ./tests/api/...

test-verbose:
	go test -v ./...

test-coverage:
	go test -cover ./...

lint:
	golangci-lint run

docker:
	docker build -t $(APP):dev .

compose:
	docker compose up --build

clean:
	go clean
	rm -f $(APP)

build:
	go build -ldflags "-X main.version=$$(git rev-parse --short HEAD 2>/dev/null || echo 'dev') -X main.buildTime=$$(date -u +%FT%TZ)" -o $(APP) .

.PHONY: run run-cli cli-ontime cli-skip2 cli-fullpay test test-unit test-api test-verbose test-coverage lint docker compose clean build
