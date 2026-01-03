# NTX Implementation Checklist

## Phase 1: Foundation

### Proto Schema
- [x] Define common.proto (Company, Price, Fundamentals, OHLCV, Report, Index, SectorSummary)
- [x] Define company.proto (CompanyService)
- [x] Define price.proto (PriceService)
- [x] Define market.proto (MarketService)
- [x] Define screener.proto (ScreenerService)
- [x] Generate Go + TypeScript clients

### Database Schema (market.db)
- [x] Create `companies` table (symbol, name, sector, description, logo_url)
- [x] Create `fundamentals` table (symbol, pe, pb, eps, book_value, market_cap, dividend_yield, roe)
- [x] Create `prices` table (symbol, date, open, high, low, close, volume, is_complete)
- [x] Create `reports` table (symbol, type, fiscal_year, quarter, revenue, net_income, eps, book_value)
- [x] Create `trading_days` table (date, status)
- [x] Write SQLC queries for all tables

### Market Hours Logic
- [x] Implement `internal/market` package
- [x] IsOpen() function (11:00-15:00 NPT, Sun-Thu)
- [x] IsTradingDay() function
- [x] NextOpen() function

---

## Phase 2: Server (ntxd)

### Background Worker
- [x] Company sync job (daily)
- [x] Fundamentals sync job (daily)
- [x] Price sync job (1/min during market hours)
- [x] Final snapshot job (after 15:00)
- [x] Historical price backfill (one-time)

### ConnectRPC Handlers
- [ ] CompanyService.ListCompanies
- [ ] CompanyService.GetCompany
- [ ] CompanyService.GetFundamentals
- [ ] CompanyService.ListReports
- [ ] PriceService.GetPrice
- [ ] PriceService.ListCandles
- [ ] MarketService.GetStatus
- [ ] MarketService.ListIndices
- [ ] MarketService.ListSectors
- [ ] ScreenerService.Screen
- [ ] ScreenerService.ListTopGainers
- [ ] ScreenerService.ListTopLosers

### Server Infrastructure
- [ ] HTTP server setup (port 8080)
- [ ] Health check endpoint (/health)
- [ ] Graceful shutdown
- [ ] Embed SvelteKit static files

---

## Phase 3: Web (SvelteKit)

### Setup
- [ ] Configure ConnectRPC client
- [ ] Create shared API client wrapper
- [ ] Setup TailwindCSS + shadcn components

### Pages
- [ ] `/` — Landing + market overview (indices, top gainers/losers)
- [ ] `/companies` — All companies grid/table with search
- [ ] `/company/[symbol]` — Company profile
  - [ ] Overview tab (price, chart)
  - [ ] Fundamentals tab (PE, EPS, ratios)
  - [ ] Financials tab (quarterly/annual reports)
  - [ ] History tab (price chart with timeframes)
- [ ] `/sectors` — Sector breakdown with aggregates
- [ ] `/screener` — Filter stocks by fundamentals

### Components
- [ ] Price display (with change indicator)
- [ ] Stock chart (OHLCV candlestick or line)
- [ ] Company card
- [ ] Sector badge
- [ ] Screener filter form

---

## Phase 4: CLI (ntx)

// Need to decide if i want to implement the portfolio management or not, it looks simple but is complex in nature. Just having a simple stock aggregator and screener is good. Let's not feature creep it. NTX will do one thing, and one thing really well. And that's stock screening with fundamentals.

### Database Schema (ntx.db)
- [ ] Create `portfolios` table
- [ ] Create `watchlists` table
- [ ] Create `price_cache` table
- [ ] Write SQLC queries

### Commands
- [ ] `ntx price <symbol>` — Get current price
- [ ] `ntx watch add/rm/list` — Manage watchlist
- [ ] `ntx portfolio add/rm/list` — Manage holdings
- [ ] `ntx portfolio summary` — Show P&L

### TUI
- [ ] Dashboard view (watchlist + portfolio summary)
- [ ] Price refresh (from go-nepse direct)
- [ ] Offline mode (from cache)

---

## Phase 5: Deployment

### Railway
- [ ] Create Dockerfile (multi-stage: Go build + SvelteKit build)
- [ ] Configure railway.toml
- [ ] Setup persistent volume for market.db
- [ ] Configure health checks
- [ ] Deploy and verify

### CLI Distribution
- [ ] Setup goreleaser
- [ ] GitHub Actions for releases
- [ ] Test `go install` path

---

## Phase 6: Future

- [ ] Historical data bootstrap (seed from CSV)
- [ ] Corporate actions (dividends, bonus) — requires go-nepse update
- [ ] WebSocket/SSE for live prices during market hours
- [ ] Report scraping (PDFs from NEPSE website)
