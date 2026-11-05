# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

The maintainer (a solo developer who invests in NEPSE) and a small circle of people around them who also hold or consider NEPSE stocks. They use NTX to look up one company and understand it quickly, on a laptop or phone, usually outside market hours.

Confirmed 2026-11-04: "yourself and a small circle." Any broader audience is unconfirmed.

## Product Purpose

NTX is a company snapshot and screener for the Nepal Stock Exchange. The company page is the product: one dense page that tells the whole story of a stock (price, fundamentals, ownership, dividends, fund holdings, peers) without clicking around. Success is a user forming their own view of a company from a single page, faster than on any incumbent NEPSE site.

## Positioning

What a neighbouring NEPSE site (ShareSansar, MeroLagani, NepseAlpha) cannot truthfully copy, confirmed 2026-11-04:

- Density: one page tells the whole story.
- Speed and no clutter: reads from its own SQLite, no ads, no login walls for public data.
- Fundamentals-first, not tips-first: ratios, earnings history, and ratings so the user forms their own view. No buy/sell calls.
- Open source and free: MIT, one author, visible on GitHub. Trust comes from transparency.

## Operating Context

- NEPSE trades 11:00–15:05 NPT, Sunday–Thursday. Outside those hours all data is a frozen end-of-day snapshot.
- Data path: go-nepse → background worker → SQLite → ConnectRPC → SvelteKit. Web requests never hit NEPSE directly; every page is a database read.
- Currency is NPR, shown as "Rs." with Nepali grouping (Lakh, Crore) for large values.
- Fiscal periods follow the Nepali calendar (Bikram Sambat): fiscal years like 2080/81 and quarters Q1–Q4. Mutual fund NAV reports carry Nepali dates.
- Sectors are the NEPSE classification: Commercial Bank, Development Bank, Finance, Microfinance, Life Insurance, Non-Life Insurance, Hydropower, Manufacturing, Hotel, Trading, Investment, Others.

## Capabilities and Constraints

Confirmed data available per company (from proto and SQLC): daily OHLCV price history, latest price with change, quarterly and annual fundamentals (EPS, PE, PB, book value, market cap), promoter/public ownership, corporate actions (bonus and cash dividends by fiscal year), mutual fund holdings of the stock, sector statistics and peers. Mutual funds: NAV per unit, net assets, report date, holdings.

Current routes: home (market overview), company list, company detail, market-cap and most-traded lists, mutual funds list and detail.

The "AI Research" feature composes a research prompt from the company's data for the user to paste into an assistant; NTX does not call any model itself.

Decided 2026-11-04: the portfolio and auth features (login, register, portfolios, transactions) are removed. They are not part of the product going forward.

Decided 2026-11-04: every surface is in scope for redesign when it is needed. Nothing in the current look is protected.

Decided 2026-11-04: the product thesis is "every page opens with what the data says, in prose, with provenance on every number" (see docs/superpowers/specs/2026-11-04-ntx-content-redesign-design.md). Three sub-projects in order: content redesign (followed on 2026-11-05 by 1b: sidebar navigation, company page sections, A to D rating, screener), hydro map of Nepal's rivers and listed plants with daily weather and discharge (spec approved 2026-11-05: docs/superpowers/specs/2026-11-05-ntx-hydro-map-design.md; v1 is holdings-free), and a personal rules layer (holdings, break conditions, caps, river-basin rule, gates calendar).

Undecided: whether a screener with fundamental filters ships as its own route (the README names it; no route exists yet). Where personal holdings are stored (local-first vs owner token).

## Brand Commitments

- Name: NTX. Domain: ntx.anishshrestha.com. Repository: github.com/voidarchive/ntx. License: MIT.
- Favicon: apps/web/src/lib/assets/favicon.png.
- No visual constraint was pinned. The current Fira Code monospace typography and cream palette were an experiment, not a commitment.

## Evidence on Hand

- Real NEPSE market data in the local SQLite database, refreshed by the worker.
- No testimonials, user counts, press, or benchmarks exist. Future work must not invent them.

## Product Principles

1. One page, whole story. A company page answers the question without a second click.
2. Numbers first, opinions never. Show the data and the derived ratios; never a buy/sell call.
3. Fast and free with no walls. Public data stays public and loads instantly.
4. Nepali context is native. NPR, Lakh/Crore, Bikram Sambat fiscal periods, and NEPSE sectors are the defaults, not localisation.
5. Built in the open. The source is the proof.
