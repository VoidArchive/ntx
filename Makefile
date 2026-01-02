.PHONY: build run dev test lint fmt proto clean migrate-create migrate-up migrate-down migrate-status sqlc

build:
	go build -o bin/ntxd ./cmd/ntxd
	go build -o bin/ntx ./cmd/ntx

run:
	go run ./cmd/ntx

dev:
	air

dev-web:
	cd web && pnpm dev

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	golangci-lint run --fix

proto:
	cd proto && buf lint && buf generate

clean:
	rm -rf bin/ tmp/

tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

migrate-create:
	goose -dir internal/database/migrations create $(NAME) sql

migrate-up:
	goose -dir internal/database/migrations sqlite3 ~/.local/share/ntx/ntx.db up

migrate-down:
	goose -dir internal/database/migrations sqlite3 ~/.local/share/ntx/ntx.db down

migrate-status:
	goose -dir internal/database/migrations sqlite3 ~/.local/share/ntx/ntx.db status

sqlc:
	sqlc generate
