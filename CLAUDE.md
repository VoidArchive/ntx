# NTX - NEPSE Stock Aggregator

## Project Structure

```
cmd/
  ntx/              # CLI application
    main.go         # Entry point
  ntxd/             # ConnectRPC server
internal/           # Shared business logic
  nepse/            # Market data (go-nepse wrapper)
  database/         # SQLite + migrations + SQLC
gen/
  go/               # Generated Go (protobuf + connect)
  ts/               # Generated TypeScript (protobuf)
proto/              # Proto definitions
web/                # SvelteKit frontend
```

### Package Placement

- `cmd/ntx/` - CLI-specific code
- `cmd/ntxd/` - Server-specific code
- `internal/` - shared across binaries, not importable externally

## Conventions

### Comments
Write **why**, not what. Skip obvious comments.

### Go Naming
No `Get` prefix on struct getters:
```go
func (s *Stock) Symbol() string    // good
func (s *Stock) GetSymbol() string // bad
```

### Proto/RPC Naming
Keep names short. Follow Google API conventions:

| Pattern | Use Case | Example |
|---------|----------|---------|
| `Get*` | Single resource | `GetStock`, `GetPrice` |
| `List*` | Collection | `ListStocks`, `ListPrices` |
| Verb | Action | `Analyze`, `Compare` |
| Noun | Simple getter | `Status`, `Sector` |

## Commands

```bash
# Proto
cd proto && buf lint && buf generate

# Go
go build ./...
go test ./...

# Web
cd web && pnpm dev
cd web && pnpm check
```

## Stack

**Backend**
- Go 1.25
- ConnectRPC (proto)
- SQLite (data storage)
- go-nepse (market data)

**CLI**
- kong (CLI parsing)
- lipgloss (styling)

**Web**
- SvelteKit + Tailwind + shadcn

## Data Storage

```
~/.local/share/ntx/
    ntx.db              # SQLite - stocks, market data cache

~/.config/ntx/
    config.toml         # User settings
```

Use `adrg/xdg` for cross-platform paths.
