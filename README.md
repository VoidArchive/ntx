# NTX

> NEPSE Stock Aggregator - Market data, analysis, and insights

Fast, offline-first stock aggregator for Nepal Stock Exchange (NEPSE). 

## Features (In Development)

### Market Data
- Real-time prices from NEPSE
- Market indices and overview
- Top gainers and losers
- Sector performance

### Stock Analysis
- Fundamentals (PE, PB, EPS, dividend yield)
- Technical signals
- Stock comparison
- Sector analysis

### Interfaces
- **CLI** - Fast terminal commands
- **Web** - SvelteKit dashboard
- **API** - ConnectRPC for integrations

## Installation

### Prerequisites
- Go 1.25 or later

### Build from Source
```bash
git clone https://github.com/voidarchive/ntx.git
cd ntx
go build -o ntx ./cmd/ntx
```

## Usage

```bash
# Show help
ntx --help

# Get stock price (coming soon)
ntx price NABIL

# Market overview (coming soon)
ntx market

# Stock analysis (coming soon)
ntx analyze NABIL
```

## Development

### Prerequisites
- Go 1.25+
- `buf` - Protobuf linter/generator
- `sqlc` - SQL to Go code generator
- Node.js + pnpm (for web)

### Build & Test
```bash
# Build
go build ./...

# Test
go test ./...

# Run web dev server
make web && pnpm dev

# Run server
make dev
```

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Language** | Go 1.25 |
| **CLI** | Kong |
| **Database** | SQLite + SQLC |
| **API** | ConnectRPC |
| **Market Data** | go-nepse |
| **Web** | SvelteKit + Tailwind + shadcn |

## License

MIT
