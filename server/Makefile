.PHONY: run build test

run:
	@go run ./cmd/main.go

build:
	@go build -o ./bin/app ./cmd/main.go

test:
	@go test ./...

db-up:
	@docker compose -f compose.dev.yaml up -d 

db-down:
	@docker compose -f compose.dev.yaml down 