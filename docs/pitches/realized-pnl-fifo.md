# Pitch: Realized P&L with FIFO

## Problem

When I sell shares on NEPSE, I want to know my actual profit/loss per sale, so I can make informed selling decisions and track my trading performance.

Currently, NTX tracks unrealized P&L only. Users see their holdings' current value vs cost basis, but when they sell, there's no record of whether that sale was profitable. The `transactions` table has the data, but no calculation logic exists.

## Appetite

**2 weeks**

Why: Well-understood problem, clear implementation path. FIFO is standard, no algorithmic unknowns.

## Solution

### Overview

Implement FIFO (First-In-First-Out) cost basis matching when processing SELL transactions. Store realized P&L per transaction and surface it through CLI.

### Key Elements

**FIFO Cost Basis Matching**
When a SELL transaction is processed, match against oldest BUY lots first. Calculate realized gain/loss as: `(sell_price - cost_basis) * quantity`.

**Database Changes**
- Add `realized_pnl_paisa` column to transactions table
- Add `cost_basis_paisa` to track the matched cost per sell transaction

**CLI Commands**
- `ntx pnl` - Show realized gains/losses by symbol, with period filtering
- Update `ntx summary` to include total realized P&L alongside unrealized

### Flow

1. User imports Meroshare CSV with SELL transactions
2. System matches SELL against oldest BUY lots (FIFO)
3. Realized P&L calculated and stored per transaction
4. User runs `ntx pnl` to see trading performance
5. Summary shows combined realized + unrealized returns

## Rabbit Holes

- **LIFO/specific lot matching**: Defer to v2. FIFO is standard for Nepal tax purposes and covers 80% of use cases. Don't add complexity for edge cases.
- **Partial lot matching**: Keep simple. If selling 100 shares and oldest lot has 150, split the lot. Track remaining 50 at original cost.
- **Historical recalculation**: New imports trigger recalculation. Don't try to fix past data or support retroactive corrections.
- **Bonus/Rights cost basis**: These have zero cost. When sold, entire proceeds are gain. Already tracked separately.

## No-Gos

- Tax optimization suggestions (different job, regulatory complexity)
- Multiple accounting methods toggle (LIFO, average cost)
- Per-transaction cost basis override UI
- Wash sale rules (not applicable in Nepal)

## Open Questions

- Should `ntx pnl` default to current fiscal year or all-time?
- Format for P&L display: table vs summary cards?

## Success Criteria

- [ ] SELL transactions have realized P&L calculated and stored
- [ ] `ntx pnl` shows profit/loss breakdown by symbol
- [ ] `ntx summary` includes total realized gains
- [ ] Partial lot sales handled correctly
- [ ] Bonus/Rights shares show full proceeds as gain when sold
