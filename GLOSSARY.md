# NTX

A company snapshot and screener for the Nepal Stock Exchange. Every page opens with what the data says, in prose, with the period and source beside every number.

## Language

### Story

**Claim**:
One sentence computed from a company's rows, carrying its value, period, source and a link to its evidence.
_Avoid_: statement, finding, insight, summary

**Check**:
A published rule with a threshold, evaluated for a company to a verdict of Holds, Watch or Needs more history.
_Avoid_: test, rule (for the verdict), score, pass/fail, signal

**Holds / Watch**:
The two verdicts of a Check. Holds means the rule is satisfied; Watch means it is not. Neither is a recommendation.
_Avoid_: pass, fail, buy, sell, bullish, bearish

**Rating**:
One letter, A to D, summarising a company's Checks: the share of Holds among all ten published Checks, so a Check without history counts as not holding (A at 85% or more, B at 60%, C at 35%, else D). Unrated when fewer than four Checks evaluate; never A while any Check lacks history. Shown with its coverage, "B · 6 of 10 hold, 9 evaluated". A summary of published rules, not a recommendation (ADR 0005).
_Avoid_: score, grade, signal, buy, sell

**Axis**:
One of the five groups a Check belongs to, named by the story engine: Value, Past, Health, Dividend, Ownership. Each is a company page section and one arm of the Snowflake, whose length is the axis's share of Holds among all its Checks.
_Avoid_: category, pillar, dimension

**Trailing EPS**:
The sum of a company's last four quarter-only profits over its current shares, when those four quarters are consecutive. Every PE the engine and the screener compute uses it where it exists and the exchange's annualized EPS otherwise; a figure priced this way names its span, "4Q to Q1 2083/84".
_Avoid_: TTM (in prose), run-rate

**Strip**:
The story engine's range for one ratio (low, median, high and count), given twice: over the sector and over the company's own three years. The web draws only the median, as a dashed line on the question chart, never the bar.
_Avoid_: gauge, meter, range bar

**Story**:
The ordered claims, checks and strips for one company, as returned by the story engine.
_Avoid_: narrative, analysis, report

**Provenance**:
The period, published date and source filing attached to a figure. A figure without provenance is not shown.
_Avoid_: citation, metadata

### Personal

**Gate**:
One of the owner's own written break conditions for a holding, evaluated against a filing to a verdict of Triggered, Holds or Needs data, shown with the value, the threshold and the source. A Gate is the owner's rule, never NTX's view, and is distinct from a published Check.
_Avoid_: signal, sell trigger (as the display word), alert, recommendation

**Candidates**:
The companies that pass the owner's rules and are not held: the screener filtered by sector headroom, basin, rating and minimum size. A list of what the rules admit, not a ranking of what to buy.
_Avoid_: picks, recommendations, ideas, best buys

**Triggered / Holds / Needs data**:
The three verdicts of a Gate. Triggered means the owner's condition is met for the required consecutive periods, against a fixed threshold or against the previous period's own figure (a ratchet); Holds means it is not; Needs data means the figure or its provenance is missing. Every verdict names the period it was evaluated against.
_Avoid_: fired, breached, sell, pass, fail

**Holding**:
A company or mutual fund the owner tracks, with its thesis, gates and dated notes, in one of three states: Build toward a target, Hold, or Frozen with a reason and a review date. A watched company is a Holding with no shares.
_Avoid_: position (as the stored concept), watchlist item, portfolio entry

**Order sheet**:
The month's proposed purchases: the budget scaled by where the index sits in its own year, split across Build holdings under their caps, in whole lots of ten, with the remainder carried. A proposal the owner confirms, never an order placed.
_Avoid_: trade list, buy list, plan (for the month's sheet)

### Periods

**Fiscal year**:
A Nepali financial year running Shrawan 1 to Ashadh end (mid-July to mid-July), written in Bikram Sambat as `2083/84`. The exchange's Gregorian-span form (`2026-2027`) is stored but never displayed.
_Avoid_: FY as a display form, calendar year

**Quarter**:
One of Q1 to Q4 of a fiscal year. Q1 is Shrawan to Ashwin, the monsoon quarter.
_Avoid_: First Quarter (display form)

**Year-to-date**:
The cumulative figure the exchange publishes for a quarter, from Shrawan 1 to that quarter's end. The exchange's quarterly profit and EPS are year-to-date and annualized respectively, never quarter-only.

**Quarter-only**:
A figure for a single quarter, derived as this quarter's year-to-date minus the previous quarter's in the same fiscal year.

**Annualized EPS**:
The exchange's published quarterly EPS, scaled to a full year. For hydropower this overstates after the monsoon quarter.

### Market

**Payout**:
A fiscal year's distribution to shareholders: bonus shares, cash dividend and rights, each as a percentage of par.
_Avoid_: dividend (when bonus is included)

**Bonus-adjusted return**:
A fiscal year's return counting bonus shares and cash, excluding rights.

**Sector median**:
The median of a ratio across a sector's listed companies. For PE, companies with non-positive earnings are excluded and the count is stated.

**ROE**:
Annualized EPS over book value per share, as a percentage; both from the same quarterly filing. Unset without book value.
_Avoid_: return on equity from the income statement (the filings do not carry total equity)

**Market day**:
One trading day's summary: index close, turnover, traded shares, transactions, and the advances, declines and unchanged derived from that day's price rows.

**Notice**:
A dividend or corporate-action announcement as published by the exchange: headline, body and filing, stored verbatim.

### Hydro

**Plant**:
One generating site of a listed hydropower company: its name, river, basin, district, capacity in MW, coordinates and status (operating or under construction). A company may have several.
_Avoid_: project (for a built plant), station, site

**River**:
The named river a plant's intake is on, as the licence names it (Tamakoshi, Trishuli, Kabeli). Two plants on different tributaries are on different rivers even when they share a basin.
_Avoid_: stream, khola (as a separate concept)

**Basin**:
One of nine drainage basins of Nepal: Koshi, Gandaki, Karnali, Mahakali, Bagmati, Kamala, West Rapti (including Tinau), Babai, Kankai (including Mechi). Every river belongs to exactly one basin; a river outside the list is an error in the plant table, not a tenth basin. Colouring, basin state and the one-hydro-per-basin rule work at this level, never at the river level.
_Avoid_: watershed, catchment, region

**Reading**:
One day's rain and river discharge at a plant's coordinates, carrying its date and source. Only past and current days are readings; a forecast is not a reading.
_Avoid_: forecast, observation, data point

**Plant state / Basin state**:
A plant's state is Normal, Heavy rain or High discharge, from two published rules: the day's rain on the DHM 24-hour scale (heavy from 50 mm), and the day's discharge against the plant's own record (high above its 95th percentile). Discharge is modelled, not gauged. A basin's state is the worst state among its plants. Neither is an alert or a prediction.
_Avoid_: alert, warning, risk level, flood
