APP=pinjol

run:
	go run -ldflags "-X main.version=$$(git rev-parse --short HEAD) -X main.buildTime=$$(date -u +%FT%TZ)" .

test:
	go test ./...

docker:
	docker build -t $(APP):dev .

compose:
	docker compose up --build
