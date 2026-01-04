# NTX Architecture

## Overview

NTX is a NEPSE stock screener. Single purpose: help users find and analyze stocks.

```
go-nepse --> Background Worker --> SQLite --> ConnectRPC --> SvelteKit
```

## Components

### Data Source (go-nepse)

External library that fetches from NEPSE:
- Company list and details
- Live and historical prices
- Fundamentals (PE, EPS, book value)
- Sectors and indices

### Background Worker

Runs inside the server. Fetches data on schedule:

| Schedule | Task |
|----------|------|
| Daily (before market) | Sync companies, fundamentals |
| Every minute (11:00-15:00 NPT) | Update prices |
| After market close | Final price snapshot |

### Database (SQLite)

Single source of truth. Tables:
- `companies` - symbol, name, sector
- `fundamentals` - PE, PB, EPS, market cap
- `prices` - daily OHLCV
- `trading_days` - market calendar

### API (ConnectRPC)

Go server using `net/http` (no router framework needed). ConnectRPC handlers are just `http.Handler`.

RPC endpoints for the frontend:
- `GetCompany`, `ListCompanies`
- `GetPrice`, `ListCandles`
- `GetMarketStatus`
- `Screen` (filter stocks)

### Frontend (SvelteKit)

Pages:
- `/` - Market overview
- `/company/[symbol]` - Company details
- `/screener` - Filter stocks

## Data Flow

### Write Path

```
go-nepse --> Background Worker --> SQLite
              (scheduled)
```

Only the worker writes to the database. Never during web requests.

### Read Path

```
Web Request --> ConnectRPC --> SQLite --> Response
```

All reads come from the database. Fast, predictable, no external API calls.

## Market Hours

NEPSE: 11:00-15:00 NPT, Sunday-Thursday

```
Before 11:00  --> Yesterday's data (frozen)
11:00-15:00   --> Live updates (1/min)
After 15:00   --> Today's final data (frozen)
```

After market close, data doesn't change until the next trading day.

## Deployment

Split deployment:

```
Cloudflare Pages          Railway
├── SvelteKit             ├── Go API (ConnectRPC)
├── Edge SSR              └── /data/market.db
└── Static asset cache
```

**Why split:**
- Cloudflare caches static assets at edge (fast for Nepal users)
- Railway only handles API calls (what Go is good at)
- Independent scaling
- Both have generous free tiers
