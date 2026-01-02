# Pitch: Market Service Implementation

**Status: PASSED** - Not betting this cycle. Re-pitch after core portfolio features ship.

## Problem

When I want to understand market context for my holdings, I want to see sector performance, top gainers/losers, and market indices, so I can make decisions in context of the broader market.

Proto definitions exist (`market.proto` with 7 RPC methods) but no implementation. CLI is portfolio-focused but lacks market awareness.

## Appetite

**4 weeks** (could stretch to 6 if API issues arise)

## Solution

### Overview

Implement MarketService RPCs using go-nepse library. Add CLI commands for market data. Cache results in SQLite.

### Key Elements

**RPC Implementations**
- `ListStocks` - All NEPSE symbols with basic info
- `GetStock` - Single stock details (price, volume, 52w high/low)
- `GetIndex` - NEPSE index and sub-indices
- `TopGainers` / `TopLosers` - Daily movers
- `GetSector` - Sector-wise performance

**CLI Commands**
- `ntx market` - Show index value, top 5 gainers/losers
- `ntx stock NABIL` - Detailed stock info
- `ntx sectors` - Sector performance breakdown

**Caching**
Store market data in SQLite to reduce API calls. Cache TTL: 15 minutes for prices, 1 week for stock metadata.

## Why Passed

1. **Dependency on go-nepse**: Need to verify what data is actually available. Historical prices unclear.
2. **Lower priority**: Users need realized P&L and dividends first - those affect their actual money.
3. **Scope risk**: 7 RPC methods is ambitious. Could easily balloon.
4. **Standalone value limited**: Market data without portfolio context is just a worse TradingView.

## Re-pitch Criteria

- Core portfolio features (P&L, dividends) shipped and stable
- go-nepse API capabilities verified
- Clear job validation: Do users actually want this in their portfolio tool?

## Notes for Future Shaping

- Consider cutting to 3 RPCs: `GetIndex`, `TopMovers`, `GetStock`
- Watchlist feature might be better framing than generic market data
- Integration with TUI dashboard could be compelling
