# Cycle Plan: 2026 Q1

**Dates**: Jan 6 - Feb 16 (5 weeks build + 1 week cooldown)
**Team**: Solo developer

---

## Bets

### Bet 1: Realized P&L with FIFO
- **Appetite**: 2 weeks
- **Pitch**: [realized-pnl-fifo.md](../pitches/realized-pnl-fifo.md)
- **Job**: Know actual profit/loss when selling shares

### Bet 2: TUI Dashboard
- **Appetite**: 2 weeks
- **Pitch**: [tui-dashboard.md](../pitches/tui-dashboard.md)
- **Job**: See portfolio at a glance in terminal

### Bet 3: Dividend Tracking
- **Appetite**: 2 weeks
- **Pitch**: [dividend-tracking.md](../pitches/dividend-tracking.md)
- **Job**: Track dividend income as part of total return

---

## Passed (Not This Cycle)

| Pitch | Reason |
|-------|--------|
| [Market Service](../pitches/market-service.md) | Lower priority than core portfolio features. Needs go-nepse API investigation. |
| [Web Dashboard](../pitches/web-dashboard.md) | Needs realized P&L and daily snapshots first. TUI covers immediate need. |

---

## Killed

| Idea | Reason |
|------|--------|
| TMS Broker Import | Job not validated. Different brokers have different formats. Meroshare covers most users. |
| Tax Reports | Too early. Need realized P&L and dividends working first. Regulatory requirements unclear. |

---

## Cooldown Plan (Week 6: Feb 10-16)

- [ ] Fix bugs from shipped features
- [ ] Explore watchlist feature for next cycle
- [ ] Research go-nepse historical data capabilities
- [ ] Shape web dashboard pitch with daily snapshots
- [ ] User feedback collection

---

## Success Criteria

| Bet | Definition of Done |
|-----|-------------------|
| Realized P&L | `ntx pnl` shows profit/loss per sale. Summary includes total realized gains. FIFO matching correct. |
| TUI Dashboard | `ntx` launches interactive UI. Three views work. Sync updates display. Clean exit. |
| Dividend Tracking | Cash dividends parsed. `ntx dividends` shows history. Summary includes dividend income. |

---

## Progress Tracking

Use the hill chart in [TRACKER.md](./TRACKER.md) for weekly updates.
