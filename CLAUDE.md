# CLAUDE.md - NTX Project Specification

USE sub-agent when appropriate.

## Project Overview

**NTX (NEPSE Power Terminal)** - A beautiful, professional terminal-based portfolio management tool for Nepal Stock Exchange. Built with Go and Bubbletea, designed for speed, accuracy, and visual excellence.

**Core Mission**: Create the perfect portfolio management TUI that combines functional excellence with stunning aesthetics - a tool you'll want to use every day.

## Vision Statement

A focused, beautiful portfolio management terminal that:

- Tracks your NEPSE holdings with precision (manual CSV import + transaction entry)
- Calculates P/L, unrealized gains, and portfolio analytics
- Scrapes Sharesansar for end-of-day prices (no real-time complexity)
- Generates data summaries for LLM analysis (Claude integration)
- Delivers a visually stunning experience with multiple themes (Tokyo Night, Rose Pine, etc.)

## Architecture Decisions

### Core Stack

- **Language**: Go 1.24+ (performance, concurrency, single binary)
- **UI Framework**: Bubbletea + Lipgloss (beautiful, responsive TUI)
- **Database**: SQLite with `modernc.org/sqlite` (pure Go, no CGO)
- **Security**: AES-256 encryption for sensitive data
- **Configuration**: Viper (flags > env vars > config file > defaults)
- **Money Handling**: Integer-based (paisa storage) for precision
- **Testing**: TDD focus on financial calculations and business logic

### Data Strategy

- **Primary Source**: Sharesansar scraping (end-of-day prices)
- **Portfolio Data**: Manual transaction entry + Meroshare CSV import
- **Offline-First**: Fully functional with stale price data
- **Update Frequency**: On-demand refresh, no real-time requirements

## Functional Requirements

### 1. Portfolio Management Core

```
🔄 CURRENT: Foundation Setup (Phase 1)
📋 NEXT: Database Tooling (Phase 2)
📋 NEXT: Portfolio Management (Phase 3)
```

**Essential Features**:

- **Transaction Tracking**: Manual buy/sell entry with validation
- **Holdings Management**: Current positions, quantities, average cost
- **P/L Calculations**: Realized gains, unrealized gains, total portfolio value
- **CSV Import**: Meroshare portfolio import for historical data
- **Portfolio Analytics**: Sector allocation, performance metrics, risk indicators

**Financial Accuracy Requirements**:

- Zero tolerance for floating-point errors (integer paisa storage)
- Precise dividend, bonus share, and rights issue handling
- Tax-loss harvesting calculations
- Portfolio-level and position-level metrics

### 2. Data Engine

**Price Data**: Sharesansar scraping with respect

- End-of-day price updates (no real-time complexity)
- Rate limiting and error handling
- Cached/offline operation when scraping fails
- Manual price entry as fallback

**Indicator Calculations**:

- Technical: RSI, MACD, Moving Averages
- Fundamental: P/E ratios, dividend yields, sector comparisons
- Portfolio: Allocation, risk metrics, performance attribution

**LLM Integration Point**:

- Export formatted portfolio summaries
- Key metrics and indicators for analysis
- Integration with Claude for decision support

### 3. Beautiful TUI Interface

#### Visual Design Goals

- **Professional aesthetic** inspired by btop and modern terminal tools
- **Information density** without clutter
- **Multiple color schemes** for personalization
- **Responsive layout** adapting to terminal size

#### Theme System

```go
// Supported themes
themes/
├── tokyo_night.go     // Dark purple/blue aesthetic
├── rose_pine.go       // Warm rose/pine colors  
├── gruvbox.go         // Retro warm colors
├── catppuccin.go      // Pastel elegance
├── nord.go            // Arctic blue tones
└── default.go         // Clean monochrome
```

**Theme Features**:

- Live theme switching (hotkey toggle)
- Persistent theme preferences
- Consistent color palette across all components
- Color-coded financial indicators (green gains, red losses)

#### Layout Architecture (Btop-Inspired)

```
┌─ Portfolio Overview ────────────────────────────────────────────────┐
│ Total: ₹2,45,670 (+1.8%) │ Today: +₹5,620 │ Unrealized: +₹12,340   │
└─────────────────────────────────────────────────────────────────────┘

┌─ Holdings [2] ──────────────────────┐ ┌─ Analysis [3] ─────────────┐
│ Symbol │Qty│ Avg  │ LTP  │Value│P/L │ │ Technical Indicators       │
│►NABIL  │50 │1,250 │1,320 │66k │+3.5k│ │ ───────────────────       │
│ EBL    │30 │680   │710   │21k │+0.9k│ │ Portfolio RSI: 45.2       │
│ HIDCL  │100│420   │445   │45k │+2.5k│ │ Avg P/E: 18.5             │
│ KTM    │25 │890   │920   │23k │+0.8k│ │ Sector Allocation:        │
│ ADBL   │40 │550   │565   │23k │+0.6k│ │ ├ Banking: 65%             │
│                                     │ │ ├ Hydro: 25%               │
│                                     │ │ └ Others: 10%              │
│                                     │ │                           │
│                                     │ │ Recent Activity           │
│                                     │ │ ───────────────           │
│                                     │ │ NABIL +10 @ ₹1,250       │
│                                     │ │ EBL -20 @ ₹700            │
└─────────────────────────────────────┘ └───────────────────────────┘

┌─ Status ──────────────────────────────────────────────────────────────┐
│[1]Overview [2]Holdings [3]Analysis│Tokyo Night│hjkl Move│r Refresh│q Quit│
└───────────────────────────────────────────────────────────────────────┘
```

#### UI Components

- **Portfolio Overview**: Top banner with key metrics, always visible
- **Holdings Table**: Main focus area (70% width) with selection indicator (►)
- **Analysis Sidebar**: Right panel (30% width) with technical indicators and recent activity
- **Status Bar**: Section navigation, theme indicator, key shortcuts

#### Section Navigation (Btop-Style)

**Main Sections:**

- **[1] Overview**: Portfolio summary and key statistics
- **[2] Holdings**: Current positions and holdings table (default focus)
- **[3] Analysis**: Technical indicators, sector allocation, metrics
- **[4] History**: Transaction history and performance tracking
- **[5] Market**: Market data, price updates, sector indices

#### Keyboard Navigation (Btop/Lazygit-Style)

```
Section Navigation:
- 1-5: Switch between main sections (Overview, Holdings, Analysis, History, Market)
- Tab/Shift+Tab: Alternative section switching

Vim-Style Movement (within sections):
- j/k: Move down/up in lists
- h/l: Move left/right between columns
- gg: Jump to top
- G: Jump to bottom
- /: Search/filter
- n/N: Next/previous search result

Actions:
- Enter: Select/drill down into item
- Space: Multi-select (for bulk operations)
- Esc: Go back/cancel/clear selection
- r: Refresh data
- R: Force refresh (ignore cache)
- t: Cycle themes
- T: Theme picker

Portfolio Operations:
- a: Add new transaction
- e: Edit selected transaction
- d: Delete selected transaction
- i: Import CSV (Meroshare)
- x: Export data
- b: Backup portfolio

View Options:
- s: Sort options
- f: Filter options
- v: Toggle view mode (compact/detailed)
- m: Toggle metrics sidebar

Help & Misc:
- ?: Show help/keybindings
- F1: Extended help
- q: Quit application
- Ctrl+C: Force quit
```

## Technical Implementation

### Database Schema (Enhanced)

```sql
-- Core tables (existing)
portfolios, holdings, transactions, corporate_actions

-- New tables for Phase 2.5+
price_history        -- Daily prices from Sharesansar
indicators          -- Calculated technical indicators  
market_summary      -- Sector indices, market stats
user_preferences    -- Theme, settings, watchlists
```

### Data Flow Architecture

```
CSV Import ──┐
             ├─► Portfolio DB ──► Analytics Engine ──► TUI Display
Manual Entry ┘                          │
                                        ▼
Sharesansar ────► Price Cache ──► LLM Export Format
```

### Package Structure

```
ntx/
├── cmd/ntx/main.go
├── internal/
│   ├── app/             # Application orchestration
│   ├── ui/
│   │   ├── dashboard/   # Main TUI components
│   │   ├── themes/      # Color schemes and styling
│   │   └── components/  # Reusable UI elements
│   ├── data/
│   │   ├── repository/  # Database layer (existing)
│   │   ├── migrations/  # Goose migrations
│   │   └── queries/     # SQLC generated code
│   ├── market/
│   │   ├── sharesansar/ # Price scraping
│   │   └── indicators/  # Technical calculations
│   ├── portfolio/       # Business logic
│   ├── security/        # Encryption (existing)
│   └── export/          # LLM data formatting
├── configs/
├── themes/              # Theme definitions
└── requirements/        # Generated specs
```

## Development Phases

### Phase 1: Foundation Setup (Current)

- **Project Structure**: Standard Go layout with cmd/ntx/main.go
- **Basic TUI**: Bubbletea skeleton with section navigation
- **Theme Foundation**: Basic theme system with Tokyo Night as default
- **Configuration**: Viper setup for config management
- **Navigation**: Implement btop-style keyboard navigation (1-5 sections, hjkl movement)
- **Layout**: Create responsive multi-pane layout with status bar

### Phase 2: Database Tooling

- **Goose**: Database migrations for schema evolution
- **SQLC**: Type-safe SQL query generation  
- **Schema Design**: Core tables (portfolios, holdings, transactions)
- **Repository Pattern**: Clean data access layer
- **Goal**: Production-grade database foundation

### Phase 3: Portfolio Management

- **Holdings Display**: Beautiful table with sorting, filtering
- **Transaction Entry**: Clean forms with validation
- **CSV Import**: Meroshare data import with mapping
- **P/L Calculations**: Accurate financial metrics

### Phase 4: Theme System & UI Polish

- **Multi-theme Support**: Tokyo Night, Rose Pine, Gruvbox, etc.
- **Responsive Layout**: Adapt to terminal size gracefully
- **Visual Polish**: Progress bars, charts, styled components
- **Keyboard Navigation**: Vim-like efficiency

### Phase 5: Data Engine

- **Sharesansar Integration**: Respectful scraping with rate limiting
- **Indicator Calculations**: RSI, MACD, portfolio metrics
- **LLM Export**: Formatted data for Claude analysis
- **Offline Resilience**: Graceful degradation

### Phase 6: Advanced Features

- **Portfolio Analytics**: Risk metrics, performance attribution
- **Sector Analysis**: Industry allocation and comparison
- **Historical Performance**: Charts and trend analysis
- **Export/Backup**: Data portability and security

## Quality Standards

### Financial Accuracy

- **Zero float errors**: All money calculations in integer paisa
- **Precise corporate actions**: Bonus shares, dividends, splits
- **Validated transactions**: Impossible states prevented
- **Audit trail**: Complete transaction history

### Code Quality

- **TDD approach**: Test-driven development for business logic
- **Go best practices**: Standard project layout, error handling
- **Clean architecture**: Separation of concerns, testable components
- **Documentation**: Clear README, inline comments for complex logic

### User Experience

- **Sub-second startup**: Fast application launch
- **Responsive UI**: Smooth navigation and updates
- **Error resilience**: Graceful handling of network/data issues
- **Visual excellence**: Professional, beautiful interface

## Success Metrics

### Functional Goals

- **Portfolio Accuracy**: Perfect P/L calculations vs manual verification
- **Data Reliability**: Successful Sharesansar scraping >95% of attempts
- **Performance**: Handle 100+ stocks, 5+ years history smoothly
- **Usability**: Daily use without friction or errors

### Learning Objectives

- **Go Mastery**: Advanced patterns, concurrency, testing
- **TUI Excellence**: Beautiful, responsive terminal interfaces
- **Financial Domain**: Precise calculations, domain modeling
- **Professional Tooling**: Database migrations, type-safe queries

## Context for Claude Code Development

### Development Methodology

1. **Requirements Gathering**: Use `/requirements-start [feature]` for each phase
2. **Structured Questions**: Answer discovery + expert questions thoroughly  
3. **Specification Generation**: Create comprehensive specs with `/requirements-end`
4. **Implementation**: Focus on clean, testable, beautiful code

### Key Constraints

- **Simplicity over complexity**: Avoid over-engineering
- **Offline-first design**: Work without network connectivity
- **Single binary distribution**: No external dependencies
- **Beautiful by default**: Every UI component should be visually excellent
- **Financial precision**: Zero tolerance for calculation errors

### NEPSE Domain Context

- **Trading Hours**: 11:00-15:00 NPT, Sunday-Thursday
- **Common Symbols**: NABIL, EBL, KTM, HIDCL, ADBL, etc.
- **Corporate Actions**: Bonus shares (common), dividends, rights issues
- **Data Challenges**: Limited APIs, Nepali language news, small market
- **Target Users**: Intermediate to advanced Nepali investors

# Code Comment Standards: Why-Focused with todo-comments.nvim

Transform all codebase comments to explain WHY decisions were made, not WHAT the code does. Use todo-comments.nvim keywords for actionable items.

## Core Principle

Comments explain context, reasoning, and business logic. Code should be self-documenting for functionality.

## todo-comments.nvim Keywords

```go
// TODO: Implement real-time price updates via WebSocket
// HACK: Using string concatenation for SQL - replace with SQLC
// WARN: Floating point precision loss possible here
// PERF: O(n²) complexity - optimize for large portfolios  
// NOTE: NEPSE market closes at 15:00 NPT
// TEST: Add edge case for bonus share calculations
// FIX: Race condition in price update goroutine
// BUG: Incorrect P/L calculation for split-adjusted shares
```
