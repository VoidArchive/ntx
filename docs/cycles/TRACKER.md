# Cycle Tracker: 2026 Q1

**Cycle**: Jan 6 - Feb 16
**Current Week**: 2

---

## Hill Chart

```
                    ▲ Top of Hill
                   /|\
                  / | \
        UPHILL   /  |  \   DOWNHILL
       (figuring   |   (executing)
          out)     |
                   |
─────────────────────────────────────
```

### Bet Positions

| Bet | Position | Status |
|-----|----------|--------|
| Realized P&L | ◑ Downhill | Executing |
| TUI Dashboard | ● Top | Completed |
| Dividend Tracking | — | Descoped |

**Legend**: ○ Start → ◐ Uphill → ● Top → ◑ Downhill → ◉ Done

---

## Week 1 (Jan 6-12)

### Realized P&L with FIFO
- **Position**: Uphill - Figuring Out
- **Progress**:
  - [x] Design cost basis approach (simplified from FIFO to average cost)
  - [ ] Plan database schema changes
  - [ ] Spike: Test with sample sell transactions
- **Blockers**: None
- **Scope cuts**: FIFO → Average cost (simpler, matches Meroshare WACC)

### TUI Dashboard
- **Position**: Uphill - Figuring Out
- **Progress**:
  - [x] Review Bubbletea examples
  - [x] Sketch view layouts
  - [x] Decide sync behavior (wait for `s` keypress)
  - [x] Decide refresh behavior (auto-reload after sync)
- **Blockers**: None
- **Scope cuts**: None yet
- **Decisions**:
  - Sync on `s` keypress, not on startup
  - Auto-reload from DB after sync completes
  - Show spinner during sync
  - Default to Holdings view
  - Navigation: `1/2/3` keys

### Dividend Tracking
- **Position**: Descoped
- **Reason**: Cash dividends not in Meroshare CSV (goes directly to bank). Would require manual entry UI or scraping - different feature than originally pitched.

---

## Week 2 (Jan 13-19)

### Realized P&L with Average Cost
- **Position**: Downhill - Executing
- **Progress**:
  - [x] Add database migration for realized_pnl_paisa
  - [x] Update SQLC queries for realized P&L
  - [x] Add Holding.realized_pnl to proto
  - [x] Fix recalculateHoldings with average cost P&L logic
  - [x] Update Summary to include total realized P&L
  - [x] Add unit tests for P&L calculation
  - [x] Add `ntx pnl` CLI command
- **Blockers**: None
- **Scope cuts**: FIFO → Average cost (matches Meroshare WACC)

### TUI Dashboard
- **Position**:
- **Progress**:
  - [ ] Basic Bubbletea app structure
  - [ ] Holdings view implementation
- **Blockers**:
- **Scope cuts**:

### Dividend Tracking
- **Position**:
- **Progress**:
  - [ ] Add DIVIDEND_CASH transaction type
  - [ ] Update parser
- **Blockers**:
- **Scope cuts**:

---

## Week 3 (Jan 20-26)

### Realized P&L with Average Cost
- **Position**:
- **Progress**:
  - [ ] Integration tests with real Meroshare data
  - [ ] Edge case handling (partial sells, splits)
  - [ ] Polish CLI output formatting
- **Blockers**:
- **Scope cuts**:

### TUI Dashboard
- **Position**:
- **Progress**:
  - [ ] Transactions view
  - [ ] Summary view
  - [ ] View switching
- **Blockers**:
- **Scope cuts**:

### Dividend Tracking
- **Position**:
- **Progress**:
  - [ ] Holdings dividend aggregation
  - [ ] `ntx dividends` command
- **Blockers**:
- **Scope cuts**:

---

## Week 4 (Jan 27 - Feb 2)

### Realized P&L with Average Cost
- **Position**:
- **Progress**:
  - [ ] Documentation
  - [ ] Final polish
- **Blockers**:
- **Scope cuts**:

### TUI Dashboard
- **Position**:
- **Progress**:
  - [ ] Sync functionality
  - [ ] Error handling
  - [ ] Polish and styling
- **Blockers**:
- **Scope cuts**:

### Dividend Tracking
- **Position**:
- **Progress**:
  - [ ] Summary integration
  - [ ] Total return calculation
  - [ ] Testing with real data
- **Blockers**:
- **Scope cuts**:

---

## Week 5 (Feb 3-9)

### All Bets
- **Focus**: Ship, test, fix
- **Tasks**:
  - [ ] End-to-end testing with real Meroshare data
  - [ ] Fix critical bugs
  - [ ] README updates
  - [ ] Final polish

---

## Week 6 - Cooldown (Feb 10-16)

- [ ] Bug fixes from user testing
- [ ] Explore watchlist feature
- [ ] Research go-nepse historical data
- [ ] Shape next cycle pitches
- [ ] Retrospective

---

## Scope Cuts Log

Track what gets cut and why:

| Date | Bet | Cut | Reason |
|------|-----|-----|--------|
| Jan 6 | Dividend Tracking | Entire bet | Cash dividends not in Meroshare CSV - goes to bank directly. Would need manual entry or scraping. |
| Jan 6 | Realized P&L | FIFO lot tracking | Average cost sufficient and matches Meroshare WACC. Simpler implementation. |

---

## Blockers Log

| Date | Bet | Blocker | Resolution |
|------|-----|---------|------------|
| | | | |

---

## Notes

*Add observations, decisions, and learnings here*

---

## End of Cycle Checklist

- [ ] All bets shipped or killed with clear reason
- [ ] Scope cuts documented
- [ ] Learnings captured
- [ ] Next cycle pitches shaped
- [ ] Retrospective completed
