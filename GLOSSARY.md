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
One letter, A to D, summarising a company's Checks: the share of Holds among the Checks that could be evaluated, scaled to eight (A for 7 or 8, B for 5 or 6, C for 3 or 4, D for 0 to 2). Unrated when fewer than four Checks evaluate. A summary of published rules, not a recommendation.
_Avoid_: score, grade, signal, buy, sell

**Strip**:
A horizontal range showing where one company's ratio sits, drawn twice: on the sector's distribution and on the company's own three-year range.
_Avoid_: gauge, meter, range bar

**Story**:
The ordered claims, checks and strips for one company, as returned by the story engine.
_Avoid_: narrative, analysis, report

**Provenance**:
The period, published date and source filing attached to a figure. A figure without provenance is not shown.
_Avoid_: citation, metadata

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

**Market day**:
One trading day's summary: index close, turnover, traded shares, transactions, and the advances, declines and unchanged derived from that day's price rows.

**Notice**:
A dividend or corporate-action announcement as published by the exchange: headline, body and filing, stored verbatim.
