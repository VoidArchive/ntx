# Contributing to NTX

Thanks for your interest in contributing to NTX! This document will help you get started.

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 20+ (for web frontend)
- SQLite
- Make

### Setup

```bash
# Clone the repository
git clone https://github.com/voidarchive/ntx.git
cd ntx

# Install Go tools
make tools

# Build the project
make build

# Run tests
make test
```

### Project Structure

```
ntx/
├── cmd/
│   ├── ntx/          # CLI application
│   │   ├── main.go   # Entry point
│   │   ├── cli/      # CLI commands
│   │   └── tui/      # TUI views (bubbletea)
│   └── ntxd/         # gRPC server daemon
├── internal/
│   ├── database/     # SQLite, migrations, queries
│   ├── portfolio/    # Portfolio management logic
│   ├── meroshare/    # Meroshare CSV parsers
│   ├── nepse/        # NEPSE API wrapper
│   └── analyzer/     # Stock analysis (AI-powered)
├── proto/            # Protocol Buffer definitions
├── gen/              # Generated code (protobuf)
└── web/              # SvelteKit frontend
```

## Development Workflow

### 1. Pick an Issue

- Check [ROADMAP.md](ROADMAP.md) for planned features
- Look for issues labeled `good first issue`
- Comment on the issue to claim it

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 3. Make Changes

Follow the code conventions below and make your changes.

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/portfolio/...

# Build to check for compilation errors
make build
```

### 5. Submit a Pull Request

- Write a clear PR description
- Reference any related issues
- Ensure CI passes

## Code Conventions

### Go

**Naming**
- No `Get` prefix on struct methods: `user.Name()` not `user.GetName()`
- Short, clear names: `srv` not `portfolioServiceInstance`

**Comments**
- Write **why**, not what
- Skip obvious comments

```go
// Bad
// parseQuantity parses the quantity string
func parseQuantity(s string) float64

// Good
// Meroshare uses "-" for zero quantities
func parseQuantity(s string) float64
```

**Error Handling**
- Wrap errors with context: `fmt.Errorf("failed to sync prices: %w", err)`
- Use `slog` for logging

**Money**
- Always store as `int64` paisa (1 NPR = 100 paisa)
- Never use `float64` for money calculations

### Proto/RPC

**Naming**
- Keep names short, follow Google API conventions
- Match request/response to RPC name

```proto
// Good
rpc Summary(SummaryRequest) returns (SummaryResponse);
rpc ListHoldings(ListHoldingsRequest) returns (ListHoldingsResponse);

// Bad
rpc GetPortfolioSummaryDetails(GetPortfolioSummaryDetailsRequest) returns (...);
```

### SQL

- Use SQLC for type-safe queries
- Migrations in `internal/database/migrations/`
- Query definitions in `internal/database/queries/`

```bash
# After modifying queries
make sqlc

# After adding migrations
make migrate
```

### CLI

- Use lipgloss for styling
- Keep commands short: `ntx sync` not `ntx synchronize-prices`
- Provide helpful error messages

## What to Contribute

### Good First Issues

- Add tests for existing functions
- Improve error messages
- Add new CLI output formats
- Documentation improvements
- Fix typos

### Medium Complexity

- New CLI commands
- Data parsers (TMS, broker formats)
- TUI components (bubbles)
- Database queries and reports

### Advanced

- AI/LLM integration for analysis
- Stock screener engine
- Charting and visualization
- Web dashboard features
- Performance optimizations

## Areas We Need Help

### High Priority

1. **TMS Parsers** - Different brokers have different export formats
2. **Test Coverage** - We need more tests!
3. **TUI Dashboard** - Interactive terminal interface
4. **Documentation** - Usage guides, API docs

### Domain Expertise Needed

- **NEPSE Knowledge** - Understanding of Nepal stock market specifics
- **Tax Rules** - Nepal capital gains tax calculations
- **Accounting** - Proper P&L and cost basis calculations

## Communication

- **Issues** - Bug reports, feature requests
- **Discussions** - General questions, ideas
- **Pull Requests** - Code contributions

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help newcomers get started
- Celebrate contributions of all sizes

## License

By contributing to NTX, you agree that your contributions will be licensed under the MIT License.

---

## Quick Reference

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Format
make fmt

# Generate protobuf
make proto

# Generate SQLC
make sqlc

# Run CLI
./bin/ntx --help

# Run server
./bin/ntxd
```

---

Thank you for contributing to NTX! Every contribution, no matter how small, helps make this project better for Nepali investors.
