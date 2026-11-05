# Data gotchas

What the tables do not say about themselves. Verified 2026-11-04 against go-nepse v0.2.5 and the local database; re-verify when go-nepse is bumped.

- **`fundamentals.eps` and `pe_ratio` are annualized.** The exchange scales the quarterly EPS to a full year. For hydropower, the Q1 (monsoon) figure overstates the year. Quarter-only EPS is derived, never stored.
- **`fundamentals.profit_amount` is year-to-date**, cumulative from Shrawan 1. Quarter-only profit = this row minus the previous quarter's row in the same fiscal year.
- **Fiscal year strings.** `fiscal_year` holds the exchange's Gregorian span (`2026-2027`); the Nepali form (`2083/84`) is `FYNameNepali` in go-nepse. `quarter` holds `First Quarter` .. `Fourth Quarter`, empty for annual rows. Canonical display forms are defined in `GLOSSARY.md`.
- **Filing provenance exists in go-nepse** (report submitted date, filing path, document id) and is dropped until spec #20's provenance ticket lands.
- **Ownership is a snapshot** keyed by company; history only exists once `ownership_history` is written by the sync.
- **Dividend notices** carry headline and body as free text; book-close and AGM dates are not structured fields anywhere in go-nepse.
- **Which syncs run.** The scheduler runs companies, fundamentals and prices daily. Ownership and dividends were never scheduled before #9; a fresh database has zero rows in `ownership` and `corporate_actions` until it lands.
- **Delisted companies keep their row.** The company sync upserts by exchange id and sets `status = 'D'` on every company the feed stopped returning; list, search, sector counts, sector distributions and the per-company sync loops (fundamentals, ownership, dividends, prices) filter on `status = 'A'`, while `GetCompany` by symbol still serves the old page and its history. A ticker rename keeps the id, so the row simply gets the new symbol.
- **Prices come from the live-market feed**, so `open_price` is null and holiday rows can repeat the last session (#11).
- **Mutual fund NAV and holdings** are a static JSON snapshot in the web app produced by `scripts/`, not database data.
- **Bank metrics** (NPL, CAR, loan growth, distributable profit) do not exist in go-nepse.
