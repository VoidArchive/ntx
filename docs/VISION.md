# NTX: Vision Document

> Building the greatest open-source portfolio manager for NEPSE

## The Problem

Every NEPSE investor faces the same struggle:

**Tracking portfolios is painful.**
- Meroshare shows holdings but no P/L calculations
- Broker TMS systems are clunky and fragmented
- Most resort to spreadsheets - manual, error-prone, tedious
- Web apps are slow, require accounts, and your data lives on someone else's server

**Making informed decisions is harder.**
- Fundamental data is scattered across multiple sites
- No unified view of your portfolio health
- Analysis requires jumping between ShareSansar, NepseAlpha, and random Excel sheets

**Nothing works offline.**
- NepalStock API breaks regularly
- Load shedding is still reality in many areas
- Mobile data isn't always reliable

## The Vision

**NTX is the tool every serious NEPSE investor deserves but doesn't exist yet.**

A fast, offline-first, privacy-respecting portfolio manager and stock analyzer that:

1. **Runs entirely on your machine** - No accounts, no cloud, no subscriptions
2. **Works offline** - Your data is always accessible, syncs when you want
3. **Does the math correctly** - FIFO cost basis, corporate actions, dividends, taxes
4. **Helps you think** - AI-powered insights that explain, not predict
5. **Gets out of your way** - Fast CLI for power users, beautiful TUI for everyone else

## What NTX Is

- **Portfolio Manager**: Import from Meroshare/TMS, track holdings, calculate P/L
- **Stock Analyzer**: Fundamentals, comparisons, screeners, sector analysis
- **Decision Support**: AI explanations of your portfolio health and market conditions
- **Local-First Software**: SQLite database on your machine, you own your data

## What NTX Is NOT

- **Not a trading platform** - We don't execute trades
- **Not a prediction engine** - We analyze, we don't forecast
- **Not a social network** - No feeds, no follows, no engagement metrics
- **Not a subscription service** - Open source, forever free

## Why Open Source?

Nepal's developer community is growing. Our capital markets are maturing. But we lack foundational tools built by us, for us.

NTX is MIT licensed because:

1. **Trust** - You can audit exactly what the code does with your data
2. **Longevity** - The tool survives even if the original author disappears
3. **Community** - Better tools emerge when smart people collaborate
4. **Learning** - Nepali developers can study, fork, and build on real-world Go code

## The Ten-Year Goal

**NTX used by every serious NEPSE investor.**

Not through marketing. Not through hype. Through being genuinely useful.

When someone asks "what do you use to track your NEPSE portfolio?", the answer should be obvious: "NTX, obviously."

## Current State (January 2026)

**Working:**
- Import transactions from Meroshare CSV
- Import cost basis from WACC reports
- Fetch live prices from NEPSE
- Holdings view with unrealized P/L
- Portfolio summary
- Transaction history with filters
- Clean CLI with lipgloss styling

**In Progress:**
- Realized P/L tracking (FIFO)
- Dividend tracking
- TUI dashboard

**Planned:**
- Market overview and stock details
- Stock screener
- Price alerts
- AI-powered analysis
- Web dashboard

See [ROADMAP.md](../ROADMAP.md) for the full feature plan.

## How You Can Help

### Use It
The best feedback comes from real usage. Try NTX with your actual portfolio.

### Report Issues
Found a bug? CSV import failed? P/L calculation seems wrong? Open an issue.

### Contribute Code
We especially need help with:
- **TMS Parsers** - Every broker has a different export format
- **TUI Components** - Interactive terminal interface (bubbletea)
- **Test Coverage** - More tests = more confidence
- **Domain Knowledge** - Nepal tax rules, NEPSE specifics

### Spread the Word
If NTX helps you, tell other investors. Write about it. Share your workflow.

## Technical Philosophy

**Boring technology.**
Go, SQLite, Protocol Buffers. Proven, stable, maintainable. No framework churn.

**Local-first architecture.**
Remote APIs (NEPSE, ShareSansar) are enhancements, not dependencies. Everything works offline.

**UI is disposable.**
Core logic lives in services. CLI, TUI, and Web are thin clients. Rewrite any interface without touching business logic.

**Typed boundaries.**
ConnectRPC for API contracts. SQLC for database queries. Catch errors at compile time, not runtime.

## The Name

**NTX** = **N**EPSE **T**racking e**X**perience

Or maybe it's **N**epal **T**rader's e**X**pert.

Or just a short, memorable name that's easy to type: `ntx holdings`

## Join Us

NTX is built in the open. Every decision, every line of code, every roadmap item is public.

- **GitHub**: [github.com/voidarchive/ntx](https://github.com/voidarchive/ntx)
- **Issues**: Bug reports, feature requests, questions
- **PRs**: Code contributions welcome

---

*Built for Nepali investors, by Nepali developers.*
