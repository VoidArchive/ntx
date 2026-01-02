# NTX - Portfolio Management & Stock Analyzer

## Project Structure

```
cmd/
  ntx/              # Main app
    main.go         # Entry point
    cli/            # CLI commands (package cli)
    tui/            # TUI views (package tui)
  ntxd/             # ConnectRPC server
  debug/            # Debug/test tools
internal/           # Shared business logic
  meroshare/        # Transaction parser
  portfolio/        # Holdings, P&L
  nepse/            # Market data (go-nepse wrapper)
  analyzer/         # Stock analysis
gen/
  go/               # Generated Go (protobuf + connect)
  ts/               # Generated TypeScript (protobuf)
proto/              # Proto definitions
web/                # SvelteKit frontend
```

### Package Placement

- `cmd/ntx/cli/` and `cmd/ntx/tui/` - code specific to this binary
- `internal/` - shared across binaries, not importable externally

### CLI + TUI in Same Binary

```bash
ntx              # no args → TUI
ntx holdings     # subcommand → CLI output
ntx price NABIL  # CLI for scripting/piping
```

## Conventions

### Comments
Write **why**, not what. Skip obvious comments.

```go
// Bad
// parseQuantity parses the quantity string
func parseQuantity(s string) float64

// Good
// Meroshare uses "-" for zero quantities
func parseQuantity(s string) float64
```

### Go Naming
No `Get` prefix on struct getters:
```go
func (u *User) Name() string    // good
func (u *User) GetName() string // bad
```

### Proto/RPC Naming
Keep names short. Follow Google API conventions:

| Pattern | Use Case | Example |
|---------|----------|---------|
| `Get*` | Single resource | `GetStock`, `GetHolding` |
| `List*` | Collection | `ListStocks`, `ListHoldings` |
| Verb | Action | `Import`, `Analyze`, `Compare` |
| Noun | Simple getter | `Summary`, `Status`, `Sector` |

Avoid verbose Java-style names:
```proto
// Bad
rpc GetPortfolioSummaryDetails(...)
rpc FetchAllTransactionHistory(...)

// Good
rpc Summary(...)
rpc ListTransactions(...)
```

### Request/Response Messages
Match the RPC name:
```proto
rpc ListHoldings(ListHoldingsRequest) returns (ListHoldingsResponse);
rpc Summary(SummaryRequest) returns (SummaryResponse);
rpc Import(ImportRequest) returns (ImportResponse);
```

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

**CLI/TUI**
- kong (CLI parsing)
- bubbletea (TUI framework)
- lipgloss (styling - shared across CLI/TUI)
- bubbles (TUI components)

**Web**
- SvelteKit + Tailwind + shadcn

## Data Storage

```
~/.local/share/ntx/
    ntx.db              # SQLite - transactions, holdings

~/.config/ntx/
    config.toml         # User settings
```

Use `adrg/xdg` for cross-platform paths.

## Security

### File Path Handling

Server never opens file paths from client input (path traversal risk).

```
CLI:  reads file locally → sends bytes
Web:  uploads file       → sends bytes
Server: receives bytes only
```

```proto
message ImportRequest {
  bytes csv_data = 1;  // No file_path field
}
```

G304 exclusion in golangci is for local CLI tools only, not server code.
