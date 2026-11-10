# NTX sub-project 3: personal rules layer

Date: 2026-11-10. Status: approved after a grilling session; to be published as a spec issue with the six slices below as sub-issues. Parent record: `docs/superpowers/specs/2026-11-04-ntx-content-redesign-design.md`. Vocabulary: `GLOSSARY.md` (Personal section). Decisions: `docs/adr/0001`, `docs/adr/0002`, `docs/adr/0003`, `docs/adr/0006`.

## Why

The owner has run a NEPSE SIP for a year from one 900-line note: a thesis per holding, written break conditions, position and sector caps, a one-hydro-per-basin rule, a monthly order sheet and quarterly gate checks. The note's failures were all the same failure: a figure carried without its period and source (NMB's NPL, MEN's annualized EPS, KBL's NPL) drove decisions for months. NTX already holds the filings with provenance and evaluates published Checks against them. This sub-project lets the owner's own rules run against the same rows.

## Stance

NTX still makes no buy or sell call. A **Gate** is the owner's written rule, evaluated to Triggered, Holds or Needs data, with the value, threshold, period and source beside it. **Candidates** are the companies the owner's rules admit, never a ranking. The words buy, sell and signal do not appear on any surface.

## Scope

In: holdings and watchlist with transactions, gates and portfolio rules, a `/portfolio` page (gates, holdings against caps, order sheet, candidates, calendar), a "Your thesis" section on held company pages, manual bank-metric entry with provenance, export and import.

Out: portfolio snowflake, P&L or return charts, daily portfolio value history, news, dividend calendar from notices (needs the book-close parser), bank-metric scraping, notifications of any kind, fee arithmetic, multiple owners.

## 1. Access

One secret, `NTX_OWNER_TOKEN`, in the API environment. The personal RPCs require it in a header and compare with constant time. A missing or wrong token returns the same empty result with no detail. The web keeps the pasted token in local storage; without it the sidebar hides Portfolio and `/portfolio` shows only the token field. No users table, no session, no login (ADR 0003, ADR 0006).

## 2. Data

Additive migrations. Every personal row is owner-only; the public RPCs never read these tables.

- `holdings(id, symbol, status, target_shares, allocation_percent, carry_rs, drawdown_gate, thesis, frozen_reason, review_date)`. `status` is `build`, `hold` or `frozen`. A watched company is a holding with no transactions: no weight, not counted as a position. Mutual fund symbols are holdings too, valued from `nav.json`, included in weight denominators, excluded from the position count and from gates.
- `transactions(id, holding_id, day, kind, quantity, price)`. `kind` is `buy`, `sell`, `bonus`, `rights`, `ipo`. Price is per share as TMS shows it, fees included; mutual fund units use NAV as price. Quantity and WAC are derived: bonus is a zero-price lot so WAC falls; sells reduce quantity and leave WAC. Nothing is derived from `corporate_actions`; a declared bonus not yet in a transaction is shown as a hint only.
- `holding_notes(id, holding_id, day, text)`. The dated decision log, one line per entry.
- `gates(id, holding_id, kind, metric, direction, threshold, periods, text, checked_on)`. `kind` is `metric` or `event`. Metric gates: `metric` from a fixed list (eps, pe, cash_dividend_percent, bonus_percent, book_value, roe, promoter_percent, weight_percent, drawdown_percent, npl_percent, car_percent, loan_growth_percent, distributable_profit), `direction` is `above`, `below`, `not_below_prior` or `not_above_prior`, `threshold` (unused for the prior directions), and `periods` consecutive quarters or fiscal years (1 by default). The prior directions are the ratchet: MNBBL's "sell if Q2 NPL is not below the Q1 print" is `npl_percent not_below_prior`, and the threshold is always the previous period's own figure, so it moves with each filing without an edit. Event gates: `text` and `checked_on`, ticked at each review. No formula field.
- `rules(id = 1, soft_cap_percent 20, hard_cap_percent 25, sector_cap_percent 40, max_positions 8, min_entry_rs 15000, one_hydro_per_basin true, candidate_min_rating 'B', candidate_dividend_record true, monthly_budget_rs 12000, multiplier_below_median 1.5, multiplier_bottom_quarter 2.0)`. One editable row, these defaults from the owner's note. Mutual Fund counts as a sector under the sector cap.
- `bank_metrics(company_id, fiscal_year, quarter, npl_percent, car_percent, loan_growth_percent, distributable_profit, published_date, source_url)` as named in the parent spec, `UNIQUE(company_id, fiscal_year, quarter)`. Written only by the owner's form in v1; all four provenance fields required (ADR 0002). A later scraper writes the same rows.

## 3. Evaluation

All in Go, in the request, from the latest rows; nothing stored (ADR 0001). Every verdict carries the fiscal period it was evaluated against, so a verdict visibly turns to Needs data when a new quarter's fundamentals land without the owner's bank figures.

- **Metric gate**: take the last `periods` rows of the metric's series (quarterly for eps, pe, roe, bank metrics; fiscal-year for dividends and book value; current for promoter, weight, drawdown). Triggered when every row is on the wrong side of the threshold, or for a prior direction when every row fails to improve on the row before it; Holds otherwise; Needs data when fewer rows exist than `periods`, or the metric's source row has no provenance. Drawdown is latest close against bonus-adjusted WAC and runs only when `drawdown_gate` is set (default on for Hydropower).
- **Event gate**: shows `text` and `checked_on`; no verdict beyond the date.
- **Caps**: a holding over the soft cap reads Watch, over the hard cap Triggered; a sector over its cap Triggered; positions over `max_positions` Triggered; two hydro holdings sharing a basin (from `hydro_plants`) Triggered. Weights use market value from the latest close, mutual funds from `nav.json`.
- **Order sheet**: budget is `monthly_budget_rs` times a multiplier from the latest `market_days` index close placed in its trailing 52-week range: bottom quarter 2.0, below the median 1.5, else 1.0. The sheet states the index, the range and the multiplier. Each `build` holding under its soft cap gets `allocation_percent` of the budget plus its `carry_rs`, buys whole lots of 10 at the latest close, and the remainder is its proposed new carry. Frozen holdings are listed with `frozen_reason` and `review_date` and get nothing. Marking the month executed writes the proposed carry; the owner may edit it first.
- **Candidates**: the screener's stored ratings filtered by: not a holding; rating at least `candidate_min_rating`; Dividend record Check holds when `candidate_dividend_record`; sector would stay under `sector_cap_percent` after a `min_entry_rs` entry; for Hydropower, no plant in a basin a holding has a plant in, and "basin not recorded" shown for a hydro with no plant row. Each row shows the entry: shares to reach `min_entry_rs` rounded up to a lot of 10, its cost, and months of budget. Sorted by rating, then PE against sector. The section's headline states the owner's standing rule from the note: a new scrip needs a sale first, and positions stand at N of `max_positions`.
- **Calendar**: for each held company, the next filing due date as fiscal quarter end plus 30 days from the Bikram Sambat table in Go; every `review_date`; every event gate's `checked_on` plus one quarter. Flat list, soonest first. No notices.

## 4. RPC

`PortfolioService` in `proto/ntx/v1/portfolio.proto`, every method token-gated: `GetPortfolio` (holdings with weights and verdicts, caps, order sheet, candidates, calendar), `GetHoldingStory(symbol)` (thesis, notes, gates with verdicts, bank metric rows), `UpsertHolding`, `DeleteHolding`, `AddTransaction`, `AddNote`, `UpsertGate`, `DeleteGate`, `UpdateRules`, `UpsertBankMetric`, `MarkMonthExecuted(carries)`, `ExportPersonal` (one JSON of every personal table), `ImportPersonal` (replaces them).

## 5. Pages

Same section pattern as the company page: a prose claim as headline, verdict beside it, table beneath. No new visual language.

**`/portfolio`.** Lede in prose, action first and value last: "Rs. 3.74 lakh across 7 positions and 3 funds. 1 gate triggered (MNBBL, NPA crosses 5%), 1 soft cap breached (MEN 21.3% of 20%). Order sheet for Mangsir: Rs. 12,000, index in the upper half of its year." Sections in order: Gates (every Triggered and Needs data first), Holdings and caps (symbol, shares, WAC, close, value, weight against soft and hard cap, status), Order sheet, Candidates, Calendar, then Rules (the editable row) and Export/Import.

**Company page, "Your thesis" section**, owner only, for a holding or watched company: the thesis paragraph as headline, gates with verdicts and periods, the bank-metric entry form for banks, dated notes, and the declared-but-uncredited bonus hint.

Phone: same order, tables scroll inside their section.

## 6. Tests

Go unit tests per gate kind and for the cap, order sheet and candidate rules, on fixtures taken from the owner's note (MNBBL NPL 4.77% against 5 holds; NMB 4.91% against 4.5 triggers; MEN 120 shares at 664.80 over a 20% soft cap; API lots at 343 from Rs. 6,000 plus carry; NTC at 910 from Rs. 6,000 buys nothing the first month and one lot of 10 the second from carry; MNBBL NPL 4.73% at Q1 2083/84 against a threshold of 4.77% holds; MNBBL `npl_percent not_below_prior` holds at Q1 against Q4's 4.77% and would trigger at Q2 on any print of 4.73% or more). One integration test per RPC against a fixture database with and without the token. `svelte-check`. Screenshot pass at 1440 and 390.

## Slices, each an issue

1. Owner token, holdings and transactions: migration, `PortfolioService` with the token check, `GetPortfolio` returning holdings with weights, the `/portfolio` route with the token field, holdings table and transaction entry.
2. Gates and rules: `gates` and `rules` tables, metric and event gate evaluation with periods, caps, the Gates section, the "Your thesis" section with gate entry and notes.
3. Order sheet and calendar: counter-cyclical budget, lots, carry, mark executed, filing due dates.
4. Candidates: rule-filtered screener with entry size and basin check.
5. Bank metrics: `bank_metrics` table, entry form with required provenance, bank gates read from it.
6. Export and import.

## Spec issue body

Title: `Spec: sub-project 3 — personal rules layer: holdings, gates, order sheet, candidates`

Body: this file's Why, Stance and Scope sections, a link to this file, and the six slice titles as sub-issues. Label `ready-for-agent`.
