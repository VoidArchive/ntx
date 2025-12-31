.PHONY: build run dev test lint fmt proto clean

build:
	go build -o bin/ntx ./cmd/ntx
	go build -o bin/ntxd ./cmd/ntxd

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
