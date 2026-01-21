.PHONY: dev dev-api dev-web lint fmt proto tools migrate-create migrate-up migrate-down migrate-status sqlc

dev:
	make -j 2 dev-api dev-web

dev-api:
	cd apps/api && air

dev-web:
	cd apps/web && pnpm dev --host

test:
	cd apps/api && go test ./...

lint:
	cd apps/api && golangci-lint run

fmt:
	cd apps/api && golangci-lint run --fix

proto:
	cd proto && buf lint && buf generate

tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

migrate-create:
	goose -dir apps/api/internal/database/migrations create $(NAME) sql

migrate-up:
	goose -dir apps/api/internal/database/migrations sqlite3 ~/.local/share/ntx/market.db up

migrate-down:
	goose -dir apps/api/internal/database/migrations sqlite3 ~/.local/share/ntx/market.db down

migrate-status:
	goose -dir apps/api/internal/database/migrations sqlite3 ~/.local/share/ntx/market.db status

sqlc:
	cd apps/api && sqlc generate
