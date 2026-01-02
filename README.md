# NTX

> The greatest open-source portfolio manager & stock analyzer for NEPSE

**NTX** is a fast, offline-first portfolio manager and stock analyzer built specifically for Nepal Stock Exchange (NEPSE). Track your holdings, calculate P&L, analyze stocks, and make better investment decisions - all from your terminal.

**Your Data, Your Machine** - No accounts, no cloud, no subscriptions. Everything runs locally on your machine.

---

## Features

### Portfolio Management
- **Import transactions** from Meroshare CSV exports
- **Import cost basis** from WACC reports
- **Track holdings** with accurate cost calculations
- **Real-time prices** fetched from NEPSE
- **Unrealized P&L** tracking with FIFO cost basis
- **Portfolio summary** with aggregate metrics
- **Transaction history** with symbol and type filters

### Stock Analysis *(Coming Soon)*
- Stock fundamentals (PE, PB, EPS, dividend yield)
- Market overview (indices, top gainers/losers)
- Sector analysis and comparison
- AI-powered insights and recommendations

---

## Installation

### Prerequisites
- Go 1.25 or later

### Build from Source
```bash
git clone https://github.com/voidarchive/ntx.git
cd ntx
go build -o ntx ./cmd/ntx
sudo mv ntx /usr/local/bin/  # Optional: install globally
```

### Using Go Install
```bash
go install github.com/voidarchive/ntx/cmd/ntx@latest
```

---

## Quick Start

### 1. Import your transactions
Export your transaction history from Meroshare and import it:

```bash
ntx import ~/Downloads/meroshare_export.csv
```

### 2. Import cost basis (optional)
For accurate cost tracking, import your WACC report:

```bash
ntx import-wacc ~/Downloads/wacc_report.csv
```

### 3. Sync prices
Fetch live prices from NEPSE:

```bash
ntx sync
```

### 4. View your holdings
```bash
ntx holdings
```

Output:
```
Holdings (as of 2026-01-02)

Symbol    Qty    Avg Cost    Current    Value        P&L          %
----------------------------------------------------------------------
NABIL     100    1,250.00    1,450.00   145,000.00   +20,000.00   +16.00%
ADBL      50     750.00      680.00     34,000.00    -3,500.00    -4.67%
NICA      200    950.00      1,100.00   220,000.00   +30,000.00   +15.79%
```

### 5. Portfolio summary
```bash
ntx summary
```

Output:
```
Portfolio Summary

Total Investment:     485,000.00
Current Value:        512,500.00
Unrealized P&L:       +27,500.00 (+5.67%)
Holdings:             3 stocks
Last Updated:         2026-01-02 14:30:00
```

---

## Usage

### Import Commands
```bash
# Import Meroshare transactions
ntx import <file.csv>

# Import WACC cost basis
ntx import-wacc <file.csv>
```

### Portfolio Commands
```bash
# View all holdings
ntx holdings

# Portfolio summary
ntx summary

# Sync latest prices from NEPSE
ntx sync
```

### Transaction Commands
```bash
# List all transactions (default: 10 recent)
ntx transactions

# Filter by symbol
ntx transactions --symbol NABIL

# Filter by transaction type
ntx transactions --type buy

# Pagination
ntx transactions --limit 20 --offset 10
```

---

## Project Structure

```
ntx/
├── cmd/
│   ├── ntx/              # CLI application (main binary)
│   ├── ntxd/             # ConnectRPC server daemon (planned)
│   └── debug/            # Debug tools
├── internal/             # Business logic
│   ├── portfolio/        # Portfolio service
│   ├── meroshare/        # CSV parser
│   ├── nepse/            # Market data client
│   ├── database/         # SQLite + migrations + SQLC
│   └── analyzer/         # Stock analysis (planned)
├── proto/                # Protocol Buffer definitions
├── gen/                  # Generated code (Go + TypeScript)
├── web/                  # SvelteKit frontend (planned)
└── Makefile             # Build commands
```

---

## Development

### Prerequisites
- Go 1.25+
- `buf` - Protobuf linter/generator
- `sqlc` - SQL to Go code generator
- `golangci-lint` - Go linter
- `air` - Hot reload (optional)

### Build & Test
```bash
# Build binaries
make build

# Run tests
make test

# Lint code
make lint

# Format code
make fmt

# Generate protobuf code
make proto

# Generate SQLC queries
make sqlc

# Hot reload during development
make dev
```

### Database
Data is stored in `~/.local/share/ntx/ntx.db` (SQLite).

All money values are stored as **paisa** (1 NPR = 100 paisa) to avoid floating-point precision issues.

---

## Roadmap

NTX is under active development. See [ROADMAP.md](ROADMAP.md) for the complete feature plan.

### Next Up (Phase 1: Solid Foundation)
- [ ] Realized P&L tracking (FIFO/LIFO)
- [ ] Dividend tracking and tax reports
- [ ] TMS broker import support
- [ ] Data backup/export

### Future Phases
- **Phase 2**: Market intelligence (watchlist, market overview, stock details)
- **Phase 3**: Smart analysis (screener, comparisons, alerts, IPO tracker)
- **Phase 4**: Beautiful interfaces (TUI, charts, web dashboard)
- **Phase 5**: AI-powered analysis (insights, recommendations, NLP queries)
- **Phase 6**: Power features (multiple portfolios, daemon, automation)

---

## Architecture

### Tech Stack
| Layer | Technology |
|-------|------------|
| **Language** | Go 1.25 |
| **CLI** | Kong (argument parsing) |
| **Styling** | Lipgloss |
| **Database** | SQLite + SQLC + Goose |
| **API** | Protocol Buffers + ConnectRPC |
| **Market Data** | go-nepse |
| **TUI** (planned) | Bubbletea + Bubbles |
| **Web** (planned) | SvelteKit + Tailwind + shadcn |

### Key Design Decisions
- **CLI + TUI in one binary**: `ntx` (no args) launches TUI, `ntx <command>` runs CLI
- **Protobuf for everything**: Type-safe API definitions shared between Go and TypeScript
- **SQLC for queries**: Type-safe SQL without ORMs
- **Privacy-first**: Server never accepts file paths from client (CLI reads locally, sends bytes)
- **Offline-first**: All data stored locally, NEPSE sync on demand

See [CLAUDE.md](CLAUDE.md) for code conventions and development guidelines.

---

## Contributing

We welcome contributions! Whether you're fixing bugs, adding features, or improving docs - all contributions are appreciated.

### Getting Started
1. Check out [ROADMAP.md](ROADMAP.md) for feature ideas
2. Read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
3. Pick an issue or propose a feature
4. Submit a PR!

### Good First Issues
- Add CLI commands
- Improve error messages
- Add tests
- Documentation improvements

### Areas for Contribution
- TMS broker parsers (different export formats)
- TUI components (bubbletea)
- Market data features
- Stock analysis algorithms
- Web dashboard (SvelteKit)

---

## Why NTX?

- **All-in-One**: Portfolio management + stock analysis in one tool
- **Fast**: Native Go binary with instant startup
- **Offline-First**: Your data stays on your machine
- **Private**: No accounts, no tracking, no cloud sync
- **Open Source**: MIT licensed, community-driven
- **NEPSE-Native**: Built specifically for Nepal Stock Exchange
- **Extensible**: Protocol Buffer APIs for building integrations

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [go-nepse](https://github.com/samyak-jain/go-nepse) - NEPSE market data client
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Kong](https://github.com/alecthomas/kong) - CLI parser
- [SQLC](https://sqlc.dev) - Type-safe SQL

---

**Made for Nepali investors**
