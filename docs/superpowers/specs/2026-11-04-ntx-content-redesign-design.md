# NTX redesign: content first

Date: 2026-11-04. Status: approved; published as spec issue #20 (ready-for-agent). Tickets will be sub-issues of #20.

## Why

Every visual mock produced for NTX looked like a broker site or nepalstock.com. The research in `docs/research/2026-11-04-stock-page-survey.md` shows why: the distinctive stock pages worldwide (Screener.in, FireAnt, Kabutan, Simply Wall St) lead with a *sentence computed from the data*, show every ratio relative to sector and own history, and publish the rules behind any judgement. NEPSE incumbents lead with a price quote, a tab wall and technical verdicts. NTX's distinction has to come from what the page says, not the skin.

The owner's own investing notes add the personal half: a thesis per holding with written break conditions, position and sector caps, a one-hydro-per-river-basin rule, quarterly gate checks against published filings, and a rule born from being burned twice: no number without its quarter, fiscal year and source.

## Product thesis

Every page opens with what the data says, in prose, with provenance on every number. Beneath the prose, evidence: ratios relative to sector and own history, a bonus-adjusted fiscal-year record, quarterly results with record and streak labels, ownership over time. No tips, no scores, no buy/sell words.

## Decomposition

Three sub-projects sharing one database, built in this order, each usable alone:

1. **Content redesign** (this spec): story engine, provenance, company page, home page, removal of auth and portfolio.
   - **1b. Navigation, sections, rating, screener** (added 2026-11-05, see below): a sidebar, the company page split into one section per claim with a chart each, an A to D rating summarising the checks, and the companies list turned into the screener.
2. **Hydro map**: Nepal rivers and basins, every listed hydro's plants plotted from a curated table, daily rain and discharge per plant from Open-Meteo via the worker. Dual use: risk monitor for held basins, screener for cheap hydros in new basins.
3. **Personal rules layer**: holdings and watchlist, per-stock break conditions evaluated against new filings, rule checker (caps, minimum size, basin, dividend), dated gates calendar, monthly order sheet. Requires bank metrics (NPL, CAR, loan growth, distributable profit) which are not in go-nepse.

Sub-projects 2 and 3 get their own specs.

## Sub-project 1: content redesign

### 1. Data and provenance

Additive migrations only. Vocabulary per `GLOSSARY.md`; decisions in `docs/adr/0001` to `0003`.

**Found 2026-11-04:** the worker never ran the ownership and dividend syncs, so `ownership` and `corporate_actions` are empty. Stored `eps` and `pe_ratio` are annualized and `profit_amount` is year-to-date. go-nepse exposes each report's submitted date, filing path and the Bikram Sambat fiscal-year name (`FYNameNepali`), which NTX currently drops. Bank metrics do not exist in go-nepse.

- **First ticket:** schedule the existing ownership and dividend syncs as daily pre-market jobs. The ownership job writes `ownership_history` on change. The dividend job stores the notice headline, body and file path verbatim beside the corporate action (`notice_headline`, `notice_body`, `notice_url`). Book-close and AGM parsing is out of scope; "Upcoming" ships when a parser has a test corpus.
- New table `market_days(business_date, index_close, index_change, turnover, traded_shares, transactions, scrips_traded, market_cap, advances, declines, unchanged)` filled daily from go-nepse's market summary and index endpoints; advances/declines/unchanged derived from the day's price rows.

- `fundamentals` gains `published_date TEXT` (report submitted date), `source_url TEXT` (filing path) and `fiscal_year_bs TEXT` (`FYNameNepali`, e.g. `2083/84`). Backfilled by re-running the fundamentals sync, which upserts on the unique key. Rows the backfill cannot fill print "source not recorded". `quarter` is unchanged in the table; the engine normalizes it to `Q1`..`Q4` and `FY`.
- New table `ownership_history(company_id, as_of, promoter_percent, public_percent, promoter_shares, public_shares, listed_shares)` with `UNIQUE(company_id, as_of)`. The worker inserts a row whenever the fetched ownership differs from the latest stored row. The existing `ownership` table keeps the latest values.
- `corporate_actions` is unchanged. Shares outstanding per fiscal year is derived from `fundamentals.paid_up_capital / 100` (NEPSE par value Rs 100); bonus-adjusted price history is derived from `bonus_percentage` by fiscal year.
- Mutual fund holdings stay in the static `apps/web/src/lib/data/nav.json`, matched by normalized company name as today. Known weakness, deferred.
- New table `bank_metrics(company_id, fiscal_year, quarter, npl_percent, car_percent, loan_growth_percent, distributable_profit, published_date, source_url)` is created here so the schema is settled, but nothing writes to it until sub-project 3.

### 2. Story engine

Package `apps/api/internal/story`.

**Input:** one company's fundamentals, prices, corporate actions and ownership history, plus its sector's distributions (computed by SQL in the same request).

**Output:** an ordered `[]Claim` and the ratio strips.

```go
type Claim struct {
    Kind   string // earnings | valuation | payout | ownership | funds | record | check
    Text   string // the full sentence
    Value  float64
    Unit   string // "Rs", "%", "x", ""
    Period string // "Q1 2083/84" or "FY 2082/83"
    Source string // filing URL or "" (rendered as "source not recorded")
    Anchor string // id of the evidence section/row the sentence links to
}
```

**Year-to-date to quarter-only.** The exchange's quarterly `profit_amount` is cumulative from Shrawan 1 and `eps` is annualized. The engine derives quarter-only profit as this row's year-to-date minus the previous quarter's row in the same fiscal year (Q1 is itself), and quarter-only EPS as that over shares outstanding (`paid_up_capital / 100`, par Rs 100). Claims, record labels and streaks compare quarter-only figures. Year-to-date rows stay in the table untouched. Tested on NABIL's four quarters of 2082/83.

**Ratios.** Sector medians, not means. The PE median excludes companies with non-positive earnings and the strip states "of N profitable"; PB, EPS and yield medians include everyone. The company's own three-year range for PE and PB is low, median and high over its stored quarterly rows (about 12 points).

**No payout vs not recorded.** A fiscal year counts as "no payout" only if the company has a notice in a later fiscal year or a fundamentals row for that year; otherwise it is "not recorded" and dividend checks fall to "needs more history". Silence in the feed is never a fact.

**Bonus-adjusted return** per fiscal year: `(close_end × (1 + bonus%) + cash per share) / close_start − 1`, closes at fiscal-year boundaries; rights shown in the row but excluded from the return, with a footnote.

Claims computed, all from existing columns:

- **Earnings.** Latest quarter profit and EPS; change on the same quarter of the previous fiscal year; record and streak labels: "best Q1 in five years", "third consecutive year-on-year increase", "first loss in N quarters". For `sector = Hydro Power`, EPS is always stated three ways: quarter, annualized, last full year, with the monsoon quarter (Q1, Shrawan–Ashwin) flagged.
- **Valuation.** PE and PB at the latest close against the sector **median** and the company's own three-year median. Medians, not means: hydropower has 112 names.
- **Payout.** Last fiscal year's bonus, cash and rights; consecutive years with any payout; cash payout as a share of that year's profit.
- **Ownership.** Promoter and public share; changed or unchanged over the last four observations in `ownership_history`.
- **Funds.** Count of funds holding, total value, largest holder, from `nav.json` (passed in by the web layer; the engine does not read files).
- **Checks.** Eight rules in `rules.go`, each `{Name, Sentence, Evaluate(input) (Holds|Watch|NoData)}`. Initial set:
  1. EPS grew in at least 3 of the last 4 fiscal years.
  2. Net profit grew year-on-year in the latest quarter.
  3. PE is below the company's own three-year median.
  4. PB is below the sector median.
  5. Paid a dividend (bonus or cash) in each of the last 3 fiscal years.
  6. Cash payout is below 30% of last fiscal year's profit.
  7. Promoter holding unchanged over the last 4 observations.
  8. Book value per share grew over the last 3 fiscal years.
  Rendered as two lists, Holds and Watch, each item carrying its sentence, plus one quiet line "N checks need more history" for rules whose inputs do not reach back far enough. Never a score, never a buy/sell word. Rule text is in the repo and shown on the page.

Templates are Go string formatting, one per claim kind. A claim with any missing input is omitted; no placeholders.

**Tests:** one unit test per claim kind and per rule, fed real rows for NABIL (bank) and API (hydro) captured as fixtures, asserting the exact sentence and verdict. One test for the monsoon-flag path.

**RPCs** (proto `ntx/v1/story.proto`):

- `GetCompanyStory(symbol) → {claims, checks, strips}` where `strips` holds, for PE, PB, EPS and dividend yield: sector min/median/max, this company, own 3-year low/median/high.
- `GetSectorOverview() → {sectors[], filed_recently[], upcoming[]}`. Per sector: name, count, median PE, median PB, share of companies profitable in the latest quarter, aggregate profit change YoY, day change. `filed_recently`: companies with a fundamentals row whose `published_date` is within 30 days, each with its earnings claim; rows without `published_date` fall back to `created_at` and are labelled "seen by NTX on", never "published". `upcoming`: book-close and AGM dates are **not** in the current schema (`corporate_actions.submitted_date` is the filing date). The worker gains a `company_events(company_id, kind, event_date, source_url)` table filled from go-nepse announcements where it exposes them; if go-nepse has no such feed, `upcoming` is omitted from the home page and noted as a gap, not faked.

**Removed in the same cycle:** `auth.proto`, `portfolio.proto`, `internal/auth`, `internal/portfolio`, the auth interceptor, the `users`, `portfolios`, `transactions` tables (migration drops them), routes `/login`, `/register`, `/portfolio`, the `authStore`, and their components.

### 3. Company page `/company/[symbol]`

One scroll, no tabs. Every section renders for every company; a section with no data collapses to one line stating what is missing and since when.

1. **Opening.** Symbol, name, sector. If zero claims can be made, the opening reads "Not enough filings to say anything yet. Last filing: {period}." followed by the price line. Otherwise the paragraph: four to six claims (earnings, valuation, payout, ownership, funds) joined in order; each number is a link to its `Anchor`. Price, day change and business date on one small line *below* the paragraph. Checks beside the paragraph as Holds / Watch lists with rule text on hover or tap.
2. **Valuation strips.** PE, PB, EPS, dividend yield: two horizontal strips each (sector distribution with this company marked; own three-year range with today marked). Numbers at ends and median.
3. **Price and volume.** Existing chart; book-close dates, result publications and bonus ex-dates marked on the time axis. Range selector as today.
4. **Fiscal-year strip.** One row per fiscal year, newest first: EPS, net profit, book value, bonus %, cash %, rights %, shares outstanding, total return including bonus shares. Beside it, compounded growth for profit, EPS, book value and bonus-adjusted price over 1, 3, 5 years.
5. **Quarterly results.** Profit and EPS bars with record/streak labels; table with quarter, fiscal year, published date, source link. Hydro: the three EPS figures in the header.
6. **Ownership and funds.** Promoter/public as a thin stacked bar per observation; fund holdings list (units, value, share of fund).
7. **Peers.** Same sector sorted by PE, each on the shared sector strip.
8. **Data footer.** Last sync, age of newest filing, link to the raw rows. The AI Research action lives here; its prompt includes the claims.

### 4. Home page `/`

1. **Market line.** From `market_days`: business date in both calendars, index close and change, advances/declines/unchanged, turnover. Bikram Sambat dates come from a hand-checked Go table of Shrawan 1 Gregorian dates for 2075–2090 (with its source), used both for fiscal-year boundaries and for printing today's BS date by day count; no conversion library.
2. **Sectors.** Twelve rows from `GetSectorOverview`, sorted by count; a row links to its companies on the shared strips.
3. **Filed recently.** Each company's one-line earnings claim.
4. **Upcoming.** Book closes and AGMs with ex-date and a T+2 note.
5. **Movers.** Gainers, losers, most traded, largest market cap as four compact lists, last.

Mutual fund pages: add the opening sentence and provenance line only.

### 5. Visual, responsive, testing

- Visual world chosen after this spec through the Impeccable direction roll with the company page as a **Read** surface. Structure above is fixed; the roll chooses type, palette and the treatment of the strips. Hard constraints: dense; light and dark; no quote hero; no tabs; no scores; red/green never the only carrier of meaning (always sign and glyph).
- Phone (390): paragraph and checks first, strips stacked, tables scroll horizontally inside their section. Same order as desktop.
- Tests: Go unit tests per template and rule; one integration test per RPC against a fixture database; `svelte-check`; screenshot pass at 1440 and 390.

### Out of scope for sub-project 1

Hydro map, weather, holdings, break-condition evaluation, bank metrics ingestion, moving `nav.json` into SQLite. (A screener route was out of scope here and moved to 1b on 2026-11-05.)

## Sub-project 1b: navigation, sections, rating, screener

Added 2026-11-05 after sub-project 1 shipped. The pages read as an info dump and the site had no place to filter companies.

- **Sidebar**: one navigation surface replaces the top bar. Site links (Home, Companies, Mutual funds), then the search trigger and theme toggle; on a company page the section list follows, highlighting the section in view. On phones it collapses to a slim top bar with a menu button opening the same list. Market cap and most traded stop being pages; they are sort orders of the screener.
- **Company page as sections**: a two-line lede (earnings claim, rating, quote line), then one section per claim kind: Earnings, Valuation (strips with peers beneath), Payout, Ownership, Funds, Price and volume last, data footer. Each section opens with its claim sentence as the headline and shows its matching check's verdict and rule beside it. Payout gains a stacked bar of cash and bonus per fiscal year; every other section keeps its existing chart. The engine emits no new prose.
- **Rating**: computed in the API from the eight checks. A is 7 or 8 Holds of 8, B 5 or 6, C 3 or 4, D 0 to 2, scaled by the count of checks that could be evaluated; no rating when fewer than 4 evaluate. Shown once, beside the symbol, as a letter in the foreground colour with "n of m checks hold" on hover or tap. It summarises published checks and is not a recommendation; this amends the "no scores" constraint above and the glossary gains a Rating entry.
- **Screener**: the companies list at `/company` becomes a dense sortable table (symbol, name, sector, close, PE, PB, EPS, dividend yield, market cap, rating), default sort market cap descending. Filters: sector, PE range, PB range, minimum rating, text; all filter and sort state in the URL. The worker stores each company's rating after the daily sync so the screener reads a stored value; the company page keeps computing live.

## Sub-project 2: hydro map (summary; full spec in `2026-11-05-ntx-hydro-map-design.md`)

Superseded 2026-11-05 by the sub-project 2 spec: v1 is holdings-free, the basin check and mini-map move to sub-project 3, and the hydro company page gets a Plants section instead.

- Curated `data/hydro_plants.csv`: company symbol, plant, river, basin, district, MW, lat, lon, status, source URL. Built once with research help, edited by hand after.
- Map: Nepal border (geoBoundaries) and rivers (HydroRIVERS clipped) as GeoJSON rendered with d3-geo to static SVG. No tiles, no external map service.
- Weather: Open-Meteo forecast and flood (GloFAS discharge) endpoints, fetched daily by the worker per plant, stored in SQLite. Shown as readings with dates and a per-basin state (normal / heavy rain / high discharge). No alerts, no predictions.
- Surfaces: `/hydro` map with basin colouring, plant dots sized by MW, fundamentals on hover; hydro company pages gain a basin mini-map and the basin check against holdings.

## Sub-project 3: personal rules layer (summary, own spec later)

- Holdings and watchlist (storage decision open: local-first vs owner token).
- Per-stock break conditions as structured rules evaluated by the story engine against each new filing; gates shown with value vs threshold and verdict, as in the owner's 90-day table.
- Rule checker: max positions, per-stock and per-sector caps, minimum size, basin rule, dividend rule.
- Dated gates calendar filtered to holdings; monthly order sheet (budget, counter-cyclical NEPSE rule, ceilings, lots of 10).
- Bank metrics worker (NPL, CAR, loan growth, distributable profit) with provenance.
