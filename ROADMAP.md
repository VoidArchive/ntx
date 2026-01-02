# NTX Roadmap

> The greatest open-source portfolio manager & stock analyzer for NEPSE

## Vision

NTX aims to be the go-to tool for Nepali investors - combining **portfolio management** with **AI-powered stock analysis** in a fast, offline-first, privacy-respecting package that runs entirely on your machine.

**Portfolio Manager** - Track holdings, P&L, dividends, and taxes with data from Meroshare and your broker.

**Stock Analyzer** - Get AI-powered insights, fundamental analysis, and smart recommendations to make better investment decisions.

**Your Data, Your Machine** - No accounts, no cloud, no subscriptions. Everything runs locally.

---

## Current Status

### Done

- [x] Import transactions from Meroshare CSV
- [x] Import cost basis from WACC Report
- [x] Fetch live prices from NEPSE
- [x] Holdings view with unrealized P&L
- [x] Portfolio summary
- [x] Transaction history with filters
- [x] CLI with lipgloss styling

---

## Phase 1: Solid Foundation

*Goal: Complete portfolio tracking with accurate P&L*

### Realized P&L
- [ ] Track profit/loss when shares are sold
- [ ] FIFO/LIFO cost basis methods
- [ ] Sale transaction matching

### Dividend Tracking
- [ ] Record cash dividends
- [ ] Dividend yield per holding
- [ ] Annual dividend income report

### Tax Reports
- [ ] Capital gains summary (short-term vs long-term)
- [ ] Exportable tax report for filing
- [ ] Holding period tracking

### TMS Import
- [ ] Parse broker Trade Book Excel/CSV
- [ ] Import with actual buy/sell prices
- [ ] Reconcile with Meroshare transactions

### Data Integrity
- [ ] Backup/restore database
- [ ] Export portfolio to JSON/CSV
- [ ] Import from backup

---

## Phase 2: Market Intelligence

*Goal: Stay informed about the market*

### Watchlist
- [ ] Add/remove symbols to watch
- [ ] Quick price check for watchlist
- [ ] Notes per symbol

### Market Overview
- [ ] NEPSE index and sub-indices
- [ ] Market status (open/close)
- [ ] Today's turnover and volume
- [ ] Top gainers and losers
- [ ] Sector-wise performance

### Stock Details
- [ ] Company fundamentals (PE, PB, EPS)
- [ ] 52-week high/low
- [ ] Dividend history
- [ ] Price history (OHLC)
- [ ] Market depth

### Sector Analysis
- [ ] Holdings by sector breakdown
- [ ] Sector performance comparison
- [ ] Sector-wise P&L

---

## Phase 3: Smart Analysis

*Goal: Make better investment decisions*

### Stock Screener
- [ ] Filter by PE, PB, dividend yield
- [ ] Filter by sector, price range
- [ ] Sort by various metrics
- [ ] Save custom screens

### Compare Stocks
- [ ] Side-by-side comparison
- [ ] Key metrics comparison table
- [ ] Price performance comparison

### Alerts
- [ ] Price alerts (above/below threshold)
- [ ] Portfolio value alerts
- [ ] Volume spike alerts
- [ ] Store alerts locally, check on sync

### IPO Tracker
- [ ] Upcoming IPO calendar
- [ ] IPO application status
- [ ] IPO allotment results

---

## Phase 4: Beautiful Interfaces

*Goal: Delightful user experience*

### TUI Dashboard (bubbletea)
- [ ] Interactive home screen
- [ ] Real-time price updates
- [ ] Keyboard navigation
- [ ] Multiple views (portfolio, market, watchlist)
- [ ] Sparkline charts

### Charts & Visualization
- [ ] Portfolio allocation pie chart
- [ ] Holdings treemap
- [ ] Price candlestick charts (termui)
- [ ] P&L over time

### Web Dashboard (SvelteKit)
- [ ] Responsive web interface
- [ ] Portfolio overview
- [ ] Interactive charts
- [ ] Mobile-friendly

### Reports & Export
- [ ] PDF portfolio report
- [ ] Excel export
- [ ] Shareable portfolio summary

---

## Phase 5: AI-Powered Analysis

*Goal: Intelligent insights powered by AI*

### Stock Analysis
- [ ] AI-generated stock summaries
- [ ] Fundamental analysis interpretation
- [ ] Risk assessment per holding
- [ ] "What does this PE ratio mean?" explanations

### Portfolio Insights
- [ ] Portfolio health check
- [ ] Diversification recommendations
- [ ] Concentration risk warnings
- [ ] Rebalancing suggestions

### Market Commentary
- [ ] Daily market summary
- [ ] Sector trend analysis
- [ ] Notable price movements explained

### Natural Language Queries
- [ ] "How is my banking sector doing?"
- [ ] "Which stocks are undervalued?"
- [ ] "Show me my best performers this year"

### Research Assistant
- [ ] Company research summaries
- [ ] Peer comparison reports
- [ ] Investment thesis generation

---

## Phase 6: Power Features

*Goal: Advanced capabilities for power users*

### Multiple Portfolios
- [ ] Separate portfolios per account
- [ ] Family portfolio tracking
- [ ] Aggregate view across portfolios

### Background Daemon
- [ ] Auto-sync prices periodically
- [ ] Desktop notifications
- [ ] System tray integration

### Automation
- [ ] Hooks for custom scripts
- [ ] Webhook notifications
- [ ] API for external tools

### Historical Analysis
- [ ] Portfolio value over time
- [ ] Performance vs NEPSE index
- [ ] Monthly/yearly returns

---

## Technical Debt & Improvements

- [ ] Comprehensive test coverage
- [ ] CI/CD pipeline
- [ ] Performance benchmarks
- [ ] Documentation site
- [ ] Plugin architecture

---

## Contributing

We welcome contributions! Here's how you can help:

### Good First Issues
- Add a new CLI command
- Improve error messages
- Add tests for existing features
- Documentation improvements

### Medium Complexity
- Implement a Phase 1 feature
- Add new data parsers (TMS formats)
- TUI components

### Advanced
- Stock screener engine
- Charting library integration
- Web dashboard

### How to Contribute
1. Pick an unchecked item from this roadmap
2. Open an issue to discuss your approach
3. Submit a PR with your implementation
4. Get it reviewed and merged!

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.25 |
| CLI | Kong |
| TUI | Bubbletea + Lipgloss + Bubbles |
| Database | SQLite |
| API | ConnectRPC (Protocol Buffers) |
| Market Data | go-nepse |
| Web | SvelteKit + Tailwind + shadcn |

---

## Why NTX?

- **All-in-One**: Portfolio management + stock analysis in one tool
- **AI-Powered**: Intelligent insights, not just raw data
- **Fast**: Native Go binary, instant startup
- **Offline-first**: Your data stays on your machine
- **Private**: No accounts, no tracking, no cloud
- **Open Source**: MIT licensed, community-driven
- **Extensible**: Hooks, plugins, and API access
- **NEPSE-Native**: Built specifically for Nepal Stock Exchange

---

*Last updated: January 2026*
