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
- [ ] Frontend: Create `/company/[symbol]` page

**Done when:** Clicking a company shows its detail page.

---

## Cycle 3: Live Data

**Goal:** Real data from NEPSE via background worker.

- [ ] Backend: Implement company sync job (go-nepse -> SQLite)
- [ ] Backend: Schedule job to run daily
- [ ] Frontend: Verify real companies appear

**Done when:** Companies list shows real NEPSE companies.

---

## Cycle 4: Prices

**Goal:** Show current price on company page.

- [ ] DB: Create `prices` table migration
- [ ] DB: Write `GetLatestPrice`, `UpsertPrice` queries
- [ ] Proto: Define `PriceService.GetPrice`
- [ ] Backend: Implement handler
- [ ] Backend: Add price sync to background worker
- [ ] Frontend: Display price on company page

**Done when:** Company page shows today's price.

---

## Cycle 5: Fundamentals

**Goal:** Show PE, EPS, book value on company page.

- [ ] DB: Create `fundamentals` table migration
- [ ] DB: Write queries
- [ ] Proto: Define `GetFundamentals`
- [ ] Backend: Implement handler + sync job
- [ ] Frontend: Display fundamentals on company page

**Done when:** Company page shows key ratios.

---

## Cycle 6: Market Overview

**Goal:** Landing page with market status and top movers.

- [ ] DB: Add queries for top gainers/losers
- [ ] Proto: Define `MarketService.GetStatus`, `ListTopGainers`, `ListTopLosers`
- [ ] Backend: Implement handlers
- [ ] Frontend: Build landing page with indices + movers

**Done when:** Homepage shows market overview.

---

## Cycle 7: Screener

**Goal:** Filter stocks by fundamentals.

- [ ] DB: Write dynamic filter query
- [ ] Proto: Define `ScreenerService.Screen`
- [ ] Backend: Implement handler
- [ ] Frontend: Build screener page with filters

**Done when:** User can filter stocks by PE, market cap, etc.

---

## Cycle 8: Price History

**Goal:** Show price chart on company page.

- [ ] DB: Write `ListCandles` query
- [ ] Proto: Define `PriceService.ListCandles`
- [ ] Backend: Implement handler + historical backfill
- [ ] Frontend: Add chart component to company page

**Done when:** Company page shows price chart.

---

## Cycle 9: Polish & Deploy

**Goal:** Production-ready.

- [ ] Backend: Health check endpoint
- [ ] Backend: Graceful shutdown
- [ ] Backend: Embed SvelteKit build
- [ ] Deploy: Dockerfile
- [ ] Deploy: Railway config
- [ ] Deploy: Verify everything works

**Done when:** App is live on Railway.

---

## Future Cycles (if needed)

- Sectors page
- Financial reports (quarterly/annual)
- Corporate actions (dividends, bonus)
- User accounts + watchlists
