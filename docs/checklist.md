# NTX Development Checklist

Shape Up style: each cycle is a complete vertical slice (DB -> Proto -> Backend -> Frontend).

---

## Cycle 1: Hello World

**Goal:** Prove the stack works end-to-end.

- [x] DB: Create `companies` table migration
- [x] DB: Write SQLC query `ListCompanies`
- [x] Proto: Define `CompanyService.ListCompanies`
- [x] Backend: Implement handler (return hardcoded data first)
- [x] Frontend: Display company list on `/`

**Done when:** Browser shows a list of companies from the database.

---

## Cycle 2: Company Page

**Goal:** View a single company with basic info.

- [x] DB: Add `GetCompany` query
- [x] Proto: Define `CompanyService.GetCompany`
- [x] Backend: Implement handler
- [x] Frontend: Create `/company/[symbol]` page

**Done when:** Clicking a company shows its detail page.

---

## Cycle 3: Live Data

**Goal:** Real data from NEPSE via background worker.

- [x] Backend: Implement company sync job (go-nepse -> SQLite)
- [x] Backend: Schedule job to run daily
- [x] Frontend: Verify real companies appear

**Done when:** Companies list shows real NEPSE companies.

---
## Cycle 4: Fundamentals

**Goal:** Show PE, EPS, book value on company page.

- [x] DB: Create `fundamentals` table migration
- [x] DB: Write queries
- [x] Proto: Define `GetFundamentals`
- [x] Backend: Implement handler + sync job
- [x] Frontend: Display fundamentals on company page

**Done when:** Company page shows key ratios.

---

## Cycle 5: Prices

**Goal:** Show current price on company page.

- [x] DB: Create `prices` table migration
- [x] DB: Write `GetLatestPrice`, `UpsertPrice` queries
- [x] Proto: Define `PriceService.GetPrice`
- [x] Backend: Implement handler
- [x] Backend: Add price sync to background worker
- [x] Frontend: Display price on company page

**Done when:** Company page shows today's price.

---

## Cycle 6: Market Overview

**Goal:** Landing page with market status and top movers.

- [x] DB: Add queries for top gainers/losers
- [x] Proto: Define `MarketService.GetStatus`, `ListTopGainers`, `ListTopLosers`
- [x] Backend: Implement handlers
- [x] Frontend: Build landing page with indices + movers

**Done when:** Homepage shows market overview.

---

## Cycle 7: Screener

**Goal:** Filter stocks by fundamentals.

- [x] DB: Write dynamic filter query
- [x] Proto: Define `ScreenerService.Screen`
- [x] Backend: Implement handler
- [x] Frontend: Build screener page with filters

**Done when:** User can filter stocks by PE, market cap, etc.

---

## Cycle 8: Price History

**Goal:** Show price chart on company page.

- [x] DB: Write `ListCandles` query
- [x] Proto: Define `PriceService.ListCandles`
- [x] Backend: Implement handler + historical backfill
- [x] Frontend: Add chart component to company page

**Done when:** Company page shows price chart.

---

## Cycle 9: Polish & Deploy

**Goal:** Production-ready.

- [x] Backend: Health check endpoint
- [x] Backend: Graceful shutdown
- [x] Backend: Embed SvelteKit build
- [x] Deploy: Dockerfile
- [x] Deploy: Railway config
- [x] Deploy: Verify everything works

**Done when:** App is live on Railway.

---

## Future Cycles (if needed)

  ┌────────────────────┬─────────────────────────────────────────────────┐
  │      Feature       │                      Value                      │
  ├────────────────────┼─────────────────────────────────────────────────┤
  │ Peer Comparison    │ Compare to sector peers (same sector companies) │ done
  ├────────────────────┼─────────────────────────────────────────────────┤
  │ Volume Chart       │ Trading volume trends                           │ done
  ├────────────────────┼─────────────────────────────────────────────────┤
  │ News/Announcements │ Recent company news from NEPSE                  │
  ├────────────────────┼─────────────────────────────────────────────────┤
  │ Floorsheet Data    │ Large trades, broker activity                   │
  ├────────────────────┼─────────────────────────────────────────────────┤
  │ Book Closure Dates │ Upcoming dividends/AGM dates                    │
  ├────────────────────┼─────────────────────────────────────────────────┤
  │ Price Alerts       │ User notifications (requires auth)              │
  └────────────────────┴─────────────────────────────────────────────────┘
