# Pitch: Dividend Tracking

## Problem

When NEPSE companies distribute cash dividends, I want to track that income as part of my portfolio returns, so I can see my true total return including dividends.

Dividend income is significant for NEPSE investors with yields of 5-20% common among banking and hydro stocks. This income is currently completely untracked, making total return calculations incomplete.

## Appetite

**2 weeks**

Why: Data already exists in Meroshare CSV. Parsing logic exists. Mainly needs new transaction type and aggregation.

## Solution

### Overview

Parse dividend credit transactions from Meroshare CSV, store as dedicated transaction type, and aggregate dividend income in summaries.

### Key Elements

**Transaction Type Detection**
Meroshare CSV shows cash dividends as credit entries with descriptions like "CASH DIVIDEND" or "DIV-CASH". Add `DIVIDEND_CASH` to TransactionType enum. Parser already handles description-based type detection.

**Database Changes**
- New TransactionType: `DIVIDEND_CASH`
- Add `total_dividend_paisa` to holdings table for per-symbol dividend tracking
- Dividend transactions store amount in `total_paisa` field

**CLI Command: `ntx dividends`**
Show dividend history:
```
Symbol    Date        Amount
────────────────────────────
NABIL     2024-12-15  Rs. 1,250.00
NICA      2024-11-20  Rs. 800.00
...
────────────────────────────
Total Dividends: Rs. 15,430.00
```

**Summary Integration**
Update `ntx summary` to show:
- Total Dividends Received
- Dividend Yield % (dividends / total invested)
- Total Return = Unrealized P&L + Realized P&L + Dividends

### Flow

1. User imports Meroshare CSV containing dividend entries
2. Parser detects "CASH DIVIDEND" in description
3. Transaction stored as DIVIDEND_CASH type
4. Holdings updated with cumulative dividend per symbol
5. User runs `ntx dividends` to see history
6. Summary shows total dividend income

## Rabbit Holes

- **Stock dividends vs cash**: Stock dividends are already handled as BONUS transactions. Focus only on cash dividends here.
- **Dividend reinvestment tracking**: Out of scope. If user reinvests dividends, they'll appear as separate BUY transactions.
- **Dividend per share rate**: Don't try to calculate or store dividend rate. Just track total received.
- **Withholding tax**: Nepal withholds 5% on dividends. Don't track gross vs net - just store what actually hit the account.

## No-Gos

- Tax withholding calculations or reporting
- Dividend calendar or payment date predictions
- Dividend alerts or notifications
- Historical dividend rate lookups
- Dividend forecasting

## Parser Changes

Current description patterns to detect:
- "CASH DIVIDEND"
- "DIV-CASH"
- "DIVIDEND"
- "CA-DIVIDEND" (corporate action dividend)

Amount comes from credit quantity * price or directly from description parsing.

## Open Questions

- Does Meroshare CSV include dividend amount directly or need calculation?
- Should dividends appear in transactions list or separate view only?

## Success Criteria

- [ ] Cash dividend transactions parsed from Meroshare CSV
- [ ] DIVIDEND_CASH transaction type stored correctly
- [ ] `ntx dividends` shows dividend history by symbol
- [ ] `ntx summary` includes total dividend income
- [ ] Total return calculation includes dividends
- [ ] Per-holding dividend total tracked
