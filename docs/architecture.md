# NTX Architecture

## Overview

NTX is a NEPSE stock aggregator with two independent components:

1. **CLI/TUI (`ntx`)** — Local, offline-first tool for personal portfolio tracking
2. **Web (`ntxd` + SvelteKit)** — Public funnel showing company profiles, fundamentals, and market data

Both share code but operate independently.

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           DATA SOURCES                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   go-nepse                         NEPSE Website (future?)          │
│   ├── Company list                 ├── Annual reports (PDF)         │
│   ├── Company details              └── AGM announcements            │
│   ├── Live prices                                                   │
│   ├── Historical prices                                             │
│   ├── Financials (PE, EPS, Book)                                    │
│   ├── Floorsheet                                                    │
│   └── Sectors/Indices                                               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         SERVER (ntxd)                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌──────────────────┐         ┌─────────────────────────────────┐  │
│   │ Background Worker│         │          market.db              │  │
│   │                  │         │                                 │  │
│   │ • During market: │────────►│ companies                       │  │
│   │   fetch/1min     │         │ ├── symbol, name, sector        │  │
│   │                  │         │ ├── description                 │  │
│   │ • After close:   │         │ └── logo_url                    │  │
│   │   mark complete  │         │                                 │  │
│   │                  │         │ fundamentals                    │  │
│   │ • Daily:         │         │ ├── pe, pb, eps                 │  │
│   │   sync companies │         │ ├── book_value, market_cap      │  │
│   │   sync reports   │         │ └── dividend_yield              │  │
│   └──────────────────┘         │                                 │  │
│                                │ prices                          │  │
│                                │ ├── daily OHLCV                 │  │
│                                │ └── is_complete flag            │  │
│                                │                                 │  │
│                                │ reports                         │  │
│                                │ ├── quarterly financials        │  │
│                                │ └── annual financials           │  │
│                                │                                 │  │
│                                │ trading_days                    │  │
│                                │ └── date, status                │  │
│                                └─────────────────────────────────┘  │
│                                                 │                   │
│                                                 ▼                   │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                    ConnectRPC API                           │   │
│   │                                                             │   │
│   │  CompanyService              MarketService                  │   │
│   │  ├── ListCompanies           ├── GetStatus                  │   │
│   │  ├── GetCompany              ├── ListIndices                │   │
│   │  ├── GetFundamentals         └── ListSectors                │   │
│   │  └── ListReports                                            │   │
│   │                                                             │   │
│   │  PriceService                ScreenerService                │   │
│   │  ├── GetPrice                ├── Screen (filters)           │   │
│   │  └── ListCandles             ├── ListTopGainers             │   │
│   │                              └── ListTopLosers              │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                   │                                 │
└───────────────────────────────────┼─────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                            WEB (SvelteKit)                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   /                        Landing + market overview                │
│   /companies               All companies grid/table                 │
│   /company/[symbol]        Company profile                          │
│                            ├── Overview (price, chart)              │
│                            ├── Fundamentals (PE, EPS, ratios)       │
│                            ├── Financials (quarterly/annual)        │
│                            └── Price history                        │
│   /sectors                 Sector breakdown                         │
│   /screener                Filter stocks by fundamentals            │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## CLI Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                            CLI (ntx)                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Completely independent from server                                │
│                                                                     │
│   ┌─────────────┐      ┌─────────────┐      ┌───────────────────┐   │
│   │  go-nepse   │─────►│  ntx.db     │◄─────│   TUI / CLI       │   │
│   │  (direct)   │      │  (local)    │      │                   │   │
│   └─────────────┘      │             │      │  • Portfolio      │   │
│                        │  personal:  │      │  • Watchlist      │   │
│                        │  ├ portfolio│      │  • Price check    │   │
│                        │  └ watchlist│      │  • Analysis       │   │
│                        │             │      │                   │   │
│                        │  cache:     │      └───────────────────┘   │
│                        │  └─ prices  │                              │
│                        └─────────────┘                              │
│                                                                     │
│   Works fully offline with cached data                              │
│   No dependency on ntxd                                             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Write Path (Server Only)

```
go-nepse ──► Background Worker ──► market.db
                   │
                   │ Schedule:
                   │ • 1/min during market hours (11:00-15:00)
                   │ • 1x after market close (final snapshot)
                   │ • 1x daily for company/fundamentals sync
                   │
                   ▼
              market.db
```

### Read Path

```
Web Request ──► ConnectRPC ──► market.db ──► Response
                    │
                    │ Never hits go-nepse
                    │ All reads from DB
                    ▼
               (unlimited throughput)
```

---

## Market Hours Logic

NEPSE operates 11:00-15:00 NPT, Sunday-Thursday.

```
◄─────── Market Closed ───────►◄── Open ──►◄─── Closed ────────►
00:00                         11:00      15:00                24:00
         data frozen            live        data frozen
         (yesterday's)                      (today's final)
```

**Key insight:** After 15:00, data is immutable until next trading day. No cache invalidation needed — SQLite IS the source of truth.

| Time | Data Source | Fetch Strategy |
|------|-------------|----------------|
| Before 11:00 | DB (previous day) | No fetch |
| 11:00-15:00 | DB (updating) | Fetch every minute |
| After 15:00 | DB (final) | One final fetch, then stop |

---

## Data Model

### What Lives Where

| Data | Source | Refresh | Volatility |
|------|--------|---------|------------|
| Company list | go-nepse | Daily | Rare changes |
| Company details | go-nepse | Daily | Static |
| Fundamentals | go-nepse | Daily | Changes with quarterly reports |
| Live prices | go-nepse | 1/min (market hours) | Volatile |
| Historical prices | go-nepse | Once | Immutable |
| Financial reports | go-nepse | When published | Immutable |

### Database Tables

**Server (market.db):**
- `companies` — symbol, name, sector, description
- `fundamentals` — PE, PB, EPS, book value, market cap
- `prices` — daily OHLCV, is_complete flag
- `reports` — quarterly/annual financials
- `trading_days` — date, market status

**CLI (ntx.db):**
- `portfolios` — user's holdings
- `watchlists` — tracked symbols
- `price_cache` — local cache for offline use

---

## API Design (ConnectRPC)

```protobuf
service CompanyService {
  rpc ListCompanies(ListCompaniesRequest) returns (ListCompaniesResponse);
  rpc GetCompany(GetCompanyRequest) returns (GetCompanyResponse);
  rpc GetFundamentals(GetFundamentalsRequest) returns (GetFundamentalsResponse);
  rpc ListReports(ListReportsRequest) returns (ListReportsResponse);
}

service PriceService {
  rpc GetPrice(GetPriceRequest) returns (GetPriceResponse);
  rpc ListCandles(ListCandlesRequest) returns (ListCandlesResponse);
}

service MarketService {
  rpc GetStatus(GetStatusRequest) returns (GetStatusResponse);
  rpc ListIndices(ListIndicesRequest) returns (ListIndicesResponse);
  rpc ListSectors(ListSectorsRequest) returns (ListSectorsResponse);
}

service ScreenerService {
  rpc Screen(ScreenRequest) returns (ScreenResponse);
  rpc ListTopGainers(ListTopGainersRequest) returns (ListTopGainersResponse);
  rpc ListTopLosers(ListTopLosersRequest) returns (ListTopLosersResponse);
}
```

---

## Deployment

### Railway Setup

```
┌─────────────────────────────────────────┐
│              Railway                    │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │  ntxd service                   │    │
│  │  ├── Go binary                  │    │
│  │  ├── Background worker          │    │
│  │  ├── ConnectRPC server (:8080)  │    │
│  │  └── Embedded SvelteKit         │    │
│  └─────────────────────────────────┘    │
│                 │                       │
│                 ▼                       │
│  ┌─────────────────────────────────┐    │
│  │  Volume (persistent)            │    │
│  │  └── /data/market.db            │    │
│  └─────────────────────────────────┘    │
│                                         │
│  ntx.up.railway.app                     │
└─────────────────────────────────────────┘
```

**Railway config (railway.toml):**
```toml
[build]
builder = "dockerfile"

[deploy]
startCommand = "./ntxd"
healthcheckPath = "/health"
healthcheckTimeout = 5

[[mounts]]
source = "market_data"
destination = "/data"
```

### CLI Distribution

```
┌─────────────────────────────────────────┐
│         User's Machine                  │
│                                         │
│  ntx (single binary)                    │
│  └── ~/.local/share/ntx/ntx.db          │
│                                         │
│  Install via:                           │
│  • go install github.com/user/ntx/...   │
│  • brew install ntx                     │
│  • GitHub releases (goreleaser)         │
└─────────────────────────────────────────┘
```

---

## Monorepo Structure

```
ntx/
├── cmd/
│   ├── ntx/                 # CLI binary
│   │   └── main.go
│   └── ntxd/                # Server binary
│       └── main.go
├── internal/
│   ├── nepse/               # go-nepse wrapper, shared
│   ├── database/            # SQLC, migrations, shared
│   └── market/              # Market hours logic, shared
├── gen/
│   ├── go/                  # Generated protobuf (Go)
│   └── ts/                  # Generated protobuf (TS)
├── proto/
│   └── ntx/v1/              # Proto definitions
├── web/                     # SvelteKit frontend
├── docs/
│   └── architecture.md      
├── go.mod
└── railway.toml
```

**Why monorepo:**
- `internal/` shared between CLI and server
- Single proto source of truth
- One `go.mod` for dependencies
- Atomic cross-component changes

---

## Rate Limiting Strategy

go-nepse hits NEPSE's API. Rate limits apply per IP.

| Component | IP | Risk | Mitigation |
|-----------|-----|------|------------|
| CLI | User's IP | Low (personal use) | Local cache |
| Server | Railway IP | Medium | Background worker only |
| Web | N/A | None | Reads from DB only |

**Server strategy:** Web requests NEVER trigger go-nepse calls. Only the background worker fetches, on a controlled schedule (~240 calls/day during market hours).

---

## Future Considerations

1. **Historical data bootstrap** — NEPSE API doesn't provide years of history. May need to seed from external CSV source.

2. **Report scraping** — Annual reports (PDFs) may require scraping NEPSE website.

3. **Scaling** — If traffic grows, swap SQLite for PostgreSQL 

4. **Real-time updates** — Consider WebSocket/SSE for live price updates during market hours.
