# NTX Development Guide

## Core rules
- Early returns, no `else` blocks
- `const` over `let` (except Svelte runes)
- No `any` type
- No `try/catch` unless at API boundaries
- No unnecessary destructuring
- Single word variables
- One function unless reusable elsewhere
- Comments explain why, not what

## Quick Commands

```bash
# Build
make build

# Run CLI
make run

# Run dev server (hot reload with Air)
make dev

# Web dev
make dev-web          # Start SvelteKit dev server
cd web && pnpm check  # Lint + typecheck

# Proto - always regenerate after changes
make proto

# Test
make test

# Database migrations
make migrate-create NAME=name
make migrate-up
make migrate-down
make migrate-status

# SQLC - regenerate queries
make sqlc

# Linting
make lint
make fmt
```

## Development Workflow

### Adding a new RPC endpoint

1. Define in `proto/ntx/v1/*.proto`
2. `make proto` → generates `gen/go/` and `gen/ts/`
3. Implement handler in `cmd/ntxd/main.go` (or new handler file)
4. Use `internal/database/` queries via SQLC for data access
5. Import generated types in `web/src/lib/api/client.ts`

### Adding a database table

1. Create migration: `make migrate-create NAME=name`
2. Write SQLC query in `internal/database/queries/<name>.sql`
3. `make sqlc` → generates `internal/database/sqlc/`
4. Use generated `*sqlc.Queries` methods in code

### Monorepo Flow

```
proto/*.proto
    │ buf generate
    ▼
gen/go/           gen/ts/
    │                │
    │ imports        │ imports
    ▼                ▼
cmd/ntxd/      web/src/
internal/          │
    └──────────────┘
           uses
```

## Key Patterns

### SQLC Usage

```bash
make sqlc  # Regenerate after query changes
```

```go
q := sqlc.New(db)
stocks, err := q.ListStocks(ctx)
```

### Database Queries

- Write queries in `internal/database/queries/*.sql`
- Use named params for clarity: `WHERE symbol = sqlc.narg('symbol')`
- Regenerate after changes

### ConnectRPC Server

```go
// Generated handler (gen/go/ntx/v1/ntxv1connect/*.connect.go)
mux := http.NewServeMux()
marketService := &MarketService{}
mux.Handle(marketconnect.NewMarketServiceHandler(marketService))
```

## Stack Choices

| Component | Library | Why |
|-----------|---------|-----|
| RPC | ConnectRPC | Simple, gRPC-compatible, better type safety |
| CLI | kong | Structured, flags + subcommands, minimal code |
| CLI styling | lipgloss | Terminal formatting, consistent styling |
| Web | SvelteKit | SSR + hydration, great DX |
| Web UI | shadcn | Copy-paste components, no lock-in |
| Database | SQLite | Embedded, single file, great for read-heavy |

## Testing

```bash
make test              # All Go tests
go test -v ./internal/database/  # Specific package with output

# Database tests use sqlite in-memory
# See internal/database/db_test.go for setup
```

## Data Paths

```bash
# CLI data
~/.local/share/ntx/ntx.db
~/.config/ntx/config.toml

# Server data (Railway)
/data/market.db  # Mounted volume
```

Use `adrg/xdg` for cross-platform paths:
```go
dataDir, _ := xdg.DataFile("ntx/ntx.db")
```

## Common Gotchas

- **Proto changes require regeneration** - TypeScript types won't update until `make proto`
- **SQLC queries must be named** - Use `-- name: ListStocks` comments
- **Monorepo imports** - Use absolute imports from repo root: `import "gen/go/ntx/v1"`
- **Development server** - Run `ntxd` on port 8080, `web` dev server proxies to it
