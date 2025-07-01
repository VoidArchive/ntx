# Phase 3 Portfolio Management TUI - NTX Portfolio Terminal

## Problem Statement

Build a comprehensive, beautiful portfolio management interface for the NTX (NEPSE Power Terminal) that provides real-time portfolio tracking, transaction management, and financial analytics through an intuitive terminal user interface that combines functional excellence with stunning visual design.

**Key Context**: Phase 2 established the database foundation. Phase 3 focuses on creating the core portfolio management functionality that users interact with daily - transaction entry, holdings visualization, P/L tracking, and portfolio analytics.

## Solution Overview

Build a feature-rich portfolio management TUI that includes:

1. **Holdings Dashboard**: Real-time portfolio overview with key metrics
2. **Transaction Management**: Intuitive forms for buy/sell transaction entry
3. **Portfolio Analytics**: P/L calculations, performance metrics, allocation views
4. **CSV Import System**: Meroshare portfolio import with data mapping
5. **Multi-Portfolio Support**: Manage multiple portfolios with easy switching
6. **Responsive Design**: Beautiful, adaptive layouts for different terminal sizes

## Functional Requirements

### FR1: Holdings Dashboard & Overview

- **FR1.1**: Main dashboard showing portfolio summary:

  ```
  ┌─ Portfolio Overview ────────────────────────────────────────────────┐
  │ Total: Rs.2,45,670 (+1.8%) │ Today: +Rs.5,620 │ Unrealized: +Rs.12,340   │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR1.2**: Holdings table with sortable columns:

  ```
  ┌─ Holdings [4] ──────────────────────────────────────────────────────┐
  │ Symbol │Qty│ Avg Cost │  LTP   │ Value │  P/L  │ %Change │ Weight  │
  │►NABIL  │50 │  Rs.1,250  │ Rs.1,320 │ 66k   │ +3.5k │  +5.6%  │  28.5%  │
  │ EBL    │30 │    Rs.680  │   Rs.710 │ 21k   │ +0.9k │  +4.4%  │   9.1%  │
  │ HIDCL  │100│    Rs.420  │   Rs.445 │ 45k   │ +2.5k │  +6.0%  │  19.4%  │
  │ KTM    │25 │    Rs.890  │   Rs.920 │ 23k   │ +0.8k │  +3.4%  │   9.9%  │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR1.3**: Real-time portfolio metrics calculation:
  - Total portfolio value
  - Total cost basis
  - Unrealized P/L (absolute and percentage)
  - Day change (if price updates available)
  - Portfolio allocation by stock

- **FR1.4**: Color-coded indicators:
  - Green for gains, red for losses
  - Intensity based on percentage change
  - Consistent with theme color palette

- **FR1.5**: Keyboard navigation:
  - `hjkl` or arrow keys for navigation
  - `Enter` to view holding details
  - `Space` for multi-select operations

### FR2: Transaction Management System

- **FR2.1**: Transaction entry form with validation:

  ```
  ┌─ Add Transaction ───────────────────────────────────────────────────┐
  │                                                                     │
  │ Portfolio: [NEPSE Growth Portfolio    ▼]                           │
  │ Symbol:    [NABIL____________] (auto-complete from existing)        │
  │ Type:      [Buy ▼] [Sell]                                          │
  │ Quantity:  [100_______] shares                                      │
  │ Price:     [Rs.1,250.00_] per share                                   │
  │                                                                     │
  │ Commission: [Rs.25.00____] (auto-calc: 0.2% of value)               │
  │ Tax:        [Rs.1.50_____] (auto-calc: 0.015% of value)             │
  │ Date:       [2024-12-30] (today)                                   │
  │ Notes:      [________________________]                             │
  │                                                                     │
  │ Total Cost: Rs.125,026.50                                            │
  │                                                                     │
  │               [Save] [Cancel]                                       │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR2.2**: Auto-calculation features:
  - Commission: 0.2% of transaction value (configurable)
  - Tax: 0.015% of transaction value (NEPSE standard)
  - Total cost including all fees
  - Impact on average cost for existing holdings

- **FR2.3**: Transaction validation:
  - Positive quantities and prices
  - Valid symbol format (NEPSE symbols)
  - Sufficient holdings for sell transactions
  - Date validation (not future dates)

- **FR2.4**: Transaction history view:

  ```
  ┌─ Transaction History ───────────────────────────────────────────────┐
  │ Date       │ Symbol │ Type │ Qty │ Price   │ Total     │ Notes       │
  │ 2024-12-30 │ NABIL  │ Buy  │ 50  │ Rs.1,250  │ Rs.62,513   │ Initial buy │
  │ 2024-12-25 │ EBL    │ Buy  │ 30  │ Rs.680    │ Rs.20,408   │ Diversify   │
  │ 2024-12-20 │ HIDCL  │ Buy  │ 100 │ Rs.420    │ Rs.42,063   │ Hydro exp.  │
  │ 2024-12-15 │ NABIL  │ Sell │ 25  │ Rs.1,300  │ Rs.32,435   │ Profit take │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR2.5**: Transaction filtering and search:
  - Filter by symbol, type (buy/sell), date range
  - Search in notes field
  - Sort by date, symbol, value

### FR3: Portfolio Analytics & Calculations

- **FR3.1**: Portfolio performance metrics:

  ```
  ┌─ Portfolio Performance ─────────────────────────────────────────────┐
  │                                                                     │
  │ Total Investment:    Rs.2,30,050                                     │
  │ Current Value:       Rs.2,45,670                                     │
  │ Unrealized P/L:      +Rs.15,620 (+6.78%)                            │
  │ Realized P/L:        +Rs.2,150 (from 3 closed positions)            │
  │ Total Return:        +Rs.17,770 (+7.72%)                            │
  │                                                                     │
  │ Best Performer:      HIDCL (+6.0%)                                 │
  │ Worst Performer:     KTM (+3.4%)                                   │
  │ Most Weighted:       NABIL (28.5%)                                 │
  │                                                                     │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR3.2**: Sector allocation analysis:

  ```
  ┌─ Sector Allocation ─────────────────────────────────────────────────┐
  │                                                                     │
  │ Commercial Banking    ████████████████████████████ 65.2% Rs.1,60,170 │
  │ Hydropower           ████████████████ 25.8% Rs.63,480                │
  │ Manufacturing        ████ 9.0% Rs.22,020                             │
  │                                                                     │
  │ Diversification Score: 7.2/10 (Well Diversified)                   │
  │ Risk Level: Medium (based on allocation)                           │
  │                                                                     │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR3.3**: Individual holding analytics:

  ```
  ┌─ NABIL Details ─────────────────────────────────────────────────────┐
  │                                                                     │
  │ Current Holdings: 50 shares                                        │
  │ Average Cost:     Rs.1,250.00 per share                             │
  │ Current Price:    Rs.1,320.00 per share                             │
  │ Market Value:     Rs.66,000                                          │
  │ Total Cost:       Rs.62,500                                          │
  │ Unrealized P/L:   +Rs.3,500 (+5.60%)                                │
  │                                                                     │
  │ Purchase History:                                                   │
  │ 2024-12-30: +50 @ Rs.1,250                                          │
  │ 2024-12-01: +25 @ Rs.1,300                                          │
  │ 2024-11-15: -25 @ Rs.1,200 (sold)                                   │
  │                                                                     │
  │ Portfolio Weight: 28.5%                                            │
  │ Risk Contribution: Medium                                           │
  │                                                                     │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR3.4**: Average cost calculation algorithms:
  - FIFO (First In, First Out) for sell transactions
  - Weighted average cost for additional purchases
  - Accurate handling of partial sales
  - Corporate action adjustments (bonus, splits, dividends)

### FR4: CSV Import System (Meroshare Integration)

- **FR4.1**: CSV file import interface:

  ```
  ┌─ Import Portfolio Data ─────────────────────────────────────────────┐
  │                                                                     │
  │ Source: [Meroshare CSV ▼] [CDSC/TMS] [Manual CSV]                 │
  │ File:   [/home/user/meroshare_portfolio.csv] [Browse...]           │
  │                                                                     │
  │ Import Options:                                                     │
  │ ☑ Create new portfolio: "Imported Portfolio"                       │
  │ ☐ Merge with existing: [Select Portfolio ▼]                        │
  │ ☑ Import transactions (if available)                               │
  │ ☐ Import only current holdings                                      │
  │                                                                     │
  │ Data Mapping Preview:                                               │
  │ CSV Column        → NTX Field                                       │
  │ Symbol            → Symbol                                          │
  │ Quantity          → Quantity                                        │
  │ Rate              → Average Cost                                    │
  │ Amount            → Total Value                                     │
  │                                                                     │
  │               [Preview] [Import] [Cancel]                           │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR4.2**: Meroshare CSV format support:
  - Standard Meroshare portfolio export format
  - Automatic column detection and mapping
  - Handle Nepali number formats and symbols
  - Validation of imported data integrity

- **FR4.3**: Import validation and preview:
  - Show preview of data to be imported
  - Detect and flag potential issues
  - Allow field mapping customization
  - Confirm before final import

- **FR4.4**: Import conflict resolution:
  - Handle duplicate holdings (merge vs. replace)
  - Average cost calculation for merged positions
  - Import transaction history if available
  - Maintain data integrity during import

### FR5: Multi-Portfolio Management

- **FR5.1**: Portfolio switcher interface:

  ```
  ┌─ Portfolio Selection ───────────────────────────────────────────────┐
  │                                                                     │
  │ Active Portfolios:                                                  │
  │                                                                     │
  │ ►NEPSE Growth Portfolio    │ Rs.2,45,670 │ +6.78% │ 4 holdings       │
  │  Conservative Holdings     │ Rs.1,85,420 │ +3.22% │ 6 holdings       │
  │  Speculative Plays         │   Rs.45,230 │ -2.15% │ 2 holdings       │
  │                                                                     │
  │ Combined Total: Rs.4,76,320  │ +4.85%                                │
  │                                                                     │
  │              [New Portfolio] [Settings]                             │
  └─────────────────────────────────────────────────────────────────────┘
  ```

- **FR5.2**: Portfolio creation and management:
  - Create new portfolios with custom names and descriptions
  - Set portfolio-specific settings (currency, risk tolerance)
  - Archive/delete empty portfolios
  - Portfolio-level performance tracking

- **FR5.3**: Cross-portfolio analytics:
  - Combined portfolio view
  - Asset allocation across all portfolios
  - Total portfolio correlation analysis
  - Consolidated P/L reporting

- **FR5.4**: Portfolio comparison:
  - Side-by-side performance comparison
  - Risk-adjusted returns
  - Sharpe ratio and other metrics
  - Allocation overlap analysis

### FR6: Advanced UI Features & Navigation

- **FR6.1**: Responsive layout system:
  - Automatic adjustment to terminal size
  - Minimum width requirements (80 columns)
  - Collapsible sidebars for small screens
  - Mobile-friendly key bindings

- **FR6.2**: Search and filtering:

  ```
  Filter: [/NABIL_______] [Type: All ▼] [Date: Last 30 days ▼]
  ```

  - Real-time search across symbols, notes
  - Advanced filtering options
  - Saved filter presets
  - Quick filter shortcuts

- **FR6.3**: Data export functionality:
  - Export portfolio to CSV
  - Generate PDF reports
  - Custom date range exports
  - Multiple format support (CSV, JSON, PDF)

- **FR6.4**: Keyboard shortcuts and efficiency:

  ```
  Portfolio Operations:
  a: Add transaction          e: Edit selected
  d: Delete transaction       i: Import CSV
  s: Sort options            f: Filter/search
  t: Theme switcher          r: Refresh data
  
  Navigation:
  1-5: Section switching     Tab: Next panel
  hjkl: Vim navigation       /: Quick search
  Enter: Select/edit         Esc: Cancel/back
  ```

### FR7: Section-Based Navigation Architecture

- **FR7.1**: Five-section application structure:

  ```
  1. Dashboard [1] - Portfolio command center with overview + key metrics
  2. Holdings [2] - Focused holdings table for position management  
  3. Analysis [3] - Technical indicators + portfolio analytics
  4. History [4] - Transaction history + performance tracking
  5. Market [5] - Market data + sector performance + news
  ```

- **FR7.2**: Consistent btop-style UI across all sections:
  - Uniform border system with integrated titles
  - Consistent color coding and theme integration
  - Responsive layout adapting to terminal size
  - Professional appearance matching btop aesthetics

- **FR7.3**: Section-specific functionality:

  **Dashboard Section [1]:**
  ```
  ┌─[1]Dashboard─────────────────────────────────────────────────────────┐
  │ Portfolio Overview                                                   │
  │ Total: Rs.2,45,670 (+1.8%) │ Today: +Rs.5,620 │ Unrealized: +Rs.12K │
  ├──────────────────────────────────────────────────────────────────────┤
  │ Quick Market Stats          │ Recent Activity                        │
  │ NEPSE: 2,089.5 (+0.8%)     │ NABIL +10 @ Rs.1,250                  │
  │ Banking: +1.2%              │ EBL -20 @ Rs.700                      │
  │ Hydro: +0.5%                │ HIDCL +50 @ Rs.445                    │
  └──────────────────────────────────────────────────────────────────────┘
  ```

  **Holdings Section [2]:**
  ```
  ┌─[2]Holdings─────────────────────────────────────────────────────────┐
  │ Symbol │ Qty │ Cost │ LTP │ Value │ Day P/L │ Total P/L │ %Chg │ RSI │
  │►NABIL  │ 50  │1,250 │1,320│ 66k   │ +3.5k   │ +Rs.850   │+4.9% │ 58  │
  │ EBL    │ 30  │ 680  │ 710 │ 21k   │ +0.9k   │ +Rs.900   │+4.4% │ 41  │
  └─────────────────────────────────────────────────────────────────────┘
  ```

  **Analysis/History/Market Sections [3-5]:**
  - Same border system and layout principles
  - Section-specific content with btop-style presentation
  - Consistent navigation and keyboard shortcuts

- **FR7.4**: Logical information architecture:
  - **Dashboard**: High-level overview and monitoring
  - **Holdings**: Detailed position management and transactions
  - **Analysis**: Deep-dive analytics and technical indicators
  - **History**: Historical data and performance tracking
  - **Market**: External market data and sector information

- **FR7.5**: Seamless section navigation:
  - Instant switching with 1-5 keys
  - Consistent state preservation across sections
  - Visual indicators for current section
  - Contextual help and shortcuts per section

## Technical Requirements

### TR1: Performance & Responsiveness

- **TR1.1**: Real-time calculations:
  - Portfolio metrics update in <100ms
  - Holdings table refresh in <50ms
  - Search results in <200ms
  - No blocking UI during calculations

- **TR1.2**: Memory efficiency:
  - Support portfolios with 100+ holdings
  - Efficient data structures for large transaction histories
  - Lazy loading for transaction history
  - Minimal memory footprint

- **TR1.3**: Database optimization:
  - Efficient queries with proper indexing
  - Connection pooling for concurrent operations
  - Transaction batching for bulk operations
  - Cache frequently accessed calculations

### TR2: Data Integrity & Validation

- **TR2.1**: Financial calculation accuracy:
  - Zero floating-point precision errors
  - Proper rounding strategies
  - Consistent handling of fractional shares
  - Audit trail for all calculations

- **TR2.2**: Input validation:
  - Real-time validation feedback
  - Prevent invalid data entry
  - Graceful error handling
  - User-friendly error messages

- **TR2.3**: Data consistency:
  - ACID compliance for all operations
  - Referential integrity enforcement
  - Automatic backup before major operations
  - Recovery mechanisms for data corruption

### TR3: User Experience & Accessibility

- **TR3.1**: Visual design excellence:
  - Consistent color coding across themes
  - Clear visual hierarchy
  - Intuitive iconography and symbols
  - Professional appearance

- **TR3.2**: Error handling and feedback:
  - Clear error messages with suggestions
  - Progress indicators for long operations
  - Undo functionality where possible
  - Graceful degradation on errors

- **TR3.3**: Help system integration:
  - Context-sensitive help
  - Keyboard shortcut reference
  - Interactive tutorials
  - Documentation integration

### TR4: Integration & Extensibility

- **TR4.1**: Theme system compatibility:
  - Full integration with existing theme system
  - Theme-aware color coding
  - Consistent styling across components
  - Dynamic theme switching

- **TR4.2**: Plugin architecture preparation:
  - Modular component design
  - Event system for extensibility
  - Configuration hooks
  - API preparation for future features

- **TR4.3**: Future feature support:
  - Price update integration points
  - Market data visualization hooks
  - Technical indicator preparation
  - Report generation framework

## Implementation Plan

### Step 1: Core Holdings Display (Week 1) ✅ FULLY COMPLETE

1. ✅ Implement portfolio overview component
2. ✅ Create holdings table with sorting  
3. ✅ Build basic navigation system
4. ✅ Add real-time P/L calculations
5. ✅ Fix border alignment issues (Unicode character handling)

**Implementation Summary:**
- **Holdings Display Component**: `internal/ui/components/holdings/display.go` - Complete btop-style holdings management with navigation, sorting, and responsive layout
- **Table Renderer**: `internal/ui/components/holdings/table.go` - ASCII border table with integrated title, color-coded P/L, and footer shortcuts
- **Overview Integration**: `internal/ui/components/overview/overview.go` - Portfolio summary widget with perfect border alignment
- **Sample Data**: `internal/ui/components/holdings/sample_data.go` - Realistic NEPSE portfolio test data with various P/L scenarios
- **Application Integration**: Updated main application model to use new holdings component with theme switching and keyboard navigation
- **Critical Bug Fixes**: 
  - **Unicode Border Alignment**: Fixed Unicode character width calculation in border rendering (lines 109, 229, 391-392, 408-410)
  - **Component Spacing**: Corrected component spacing from double newline to single newline for seamless borders
  - **Footer Alignment**: Fixed footer content width calculations for perfect right border alignment

**Features Delivered**:
  - Btop-style integrated borders with component title and perfect alignment
  - Responsive column layout (120+, 100+, 80+, 60+ width breakpoints)
  - P/L color gradient system (green gains, red losses, gray neutral)
  - Vim-style navigation (hjkl, g/G) with selection highlighting
  - Portfolio totals calculation and display
  - Sort functionality across all columns (Symbol, Qty, Cost, LTP, Value, Day P/L, Total P/L, %Change, RSI)
  - Footer with contextual shortcuts and status information
  - Theme integration with live switching support
  - **Perfect border alignment** across all terminal sizes and themes

### Step 2: Section Restructure & UI Consistency (Week 2)

1. Restructure application sections according to new architecture
2. Move Portfolio Overview from Holdings to new Dashboard section
3. Apply btop-style UI consistently across all sections (Analysis, History, Market)
4. Implement focused, single-purpose section design
5. Ensure perfect border alignment across all components

### Step 3: Transaction Management System (Week 3)

1. Design transaction entry forms with btop-style UI
2. Implement validation and auto-calculations
3. Build transaction history viewer with consistent borders
4. Add edit/delete functionality with proper navigation

### Step 4: Portfolio Analytics (Week 4)

1. Implement portfolio metrics calculations
2. Create performance analytics views with btop-style layout
3. Build sector allocation analysis
4. Add individual holding details with consistent UI

### Step 5: CSV Import & Multi-Portfolio (Week 5)

1. Design import interface with btop-style forms
2. Implement Meroshare CSV parser and validation
3. Build portfolio switching and management
4. Add cross-portfolio analytics with consistent layout

### Step 6: Polish & Integration (Week 6)

1. Implement advanced UI features
2. Add search and filtering
3. Build export functionality
4. Complete testing and optimization

### Step 7: Code Cleanup & Architecture Refinement (Week 7)

1. Remove legacy rendering code (`renderMainInterface()` and related functions)
2. Clean up redundant components and dead code paths
3. Consolidate UI patterns and reduce code complexity
4. Optimize component structure and dependencies
5. Final code review and comprehensive documentation update
6. Performance optimization and memory usage analysis

## Acceptance Criteria

### AC1: Holdings Dashboard ✅ COMPLETE

- [x] Portfolio overview displays current metrics correctly
- [x] Holdings table shows all positions with accurate P/L
- [x] Color coding works consistently across themes
- [x] Navigation is smooth and responsive
- [x] Sorting and filtering work properly

**Verification Results:**
- ✅ Btop-style table renders with proper borders and title integration
- ✅ All columns display correctly (Symbol, Qty, Cost, LTP, Value, Day P/L, Total P/L, %Change, RSI)
- ✅ Portfolio totals calculate accurately (Rs.214.7K total value, +Rs.7.9K total P/L, +3.8% return)
- ✅ P/L color coding implemented (green gains, red losses, gradient system)
- ✅ Vim navigation works (hjkl, g/G) with selection highlighting
- ✅ Sorting cycles through all columns with direction toggle
- ✅ Responsive layout adapts to terminal width (tested at 120x30)
- ✅ Footer displays contextual shortcuts and status information
- ✅ Theme switching updates colors immediately
- ✅ Component integrates seamlessly with main application

### AC2: Transaction Management

- [ ] Transaction entry form validates all inputs
- [ ] Auto-calculations for commission and tax work correctly
- [ ] Transaction history displays chronologically
- [ ] Edit/delete operations maintain data integrity
- [ ] Average cost calculations are accurate

### AC3: Portfolio Analytics

- [ ] P/L calculations match manual verification
- [ ] Sector allocation adds up to 100%
- [ ] Performance metrics are mathematically correct
- [ ] Individual holding details are accurate
- [ ] Risk assessments are reasonable

### AC4: CSV Import

- [ ] Meroshare CSV files import successfully
- [ ] Data mapping is intuitive and accurate
- [ ] Validation catches common errors
- [ ] Import maintains data integrity
- [ ] Conflict resolution works properly

### AC5: Multi-Portfolio

- [ ] Portfolio switching works seamlessly
- [ ] Cross-portfolio analytics are accurate
- [ ] Portfolio creation/deletion works properly
- [ ] Performance comparison is meaningful
- [ ] Data isolation between portfolios

### AC6: User Experience

- [ ] Interface is intuitive and beautiful
- [ ] Keyboard navigation is efficient
- [ ] Error handling is user-friendly
- [ ] Performance meets requirements
- [ ] Theme integration is seamless

## Success Metrics

- **Functional Completeness**: All portfolio management operations available across all sections
- **Section Architecture**: Logical, intuitive section organization with clear purpose separation
- **UI Consistency**: Uniform btop-style design across all sections with perfect border alignment
- **Calculation Accuracy**: 100% accuracy in financial calculations vs. manual verification
- **Performance**: Sub-100ms response times for all portfolio operations and section switching
- **User Experience**: Intuitive interface requiring minimal learning curve with efficient navigation
- **Code Quality**: Clean, maintainable codebase with minimal complexity and no dead code
- **Data Integrity**: Zero data loss or corruption incidents during all operations
- **Scalability**: Support for portfolios with 100+ holdings and 1000+ transactions

## Constraints & Assumptions

### Constraints

- Terminal-based interface only (no GUI dependencies)
- SQLite database backend (established in Phase 2)
- Integer-based financial calculations (paisa precision)
- NEPSE market focus (Nepali symbols and conventions)

### Assumptions

- Users understand basic financial concepts
- Portfolio sizes are reasonable (<100 holdings typically)
- Manual price updates acceptable (no real-time feed)
- Single-user application (no concurrent access)

## Future Phase Preparation

This Phase 3 implementation prepares for:

- **Phase 4**: Market data integration and price updates
- **Phase 5**: Technical indicators and advanced analytics
- **Phase 6**: Reporting and data visualization
- **Phase 7**: API integration and automation

The portfolio management foundation should be extensible for market data integration, technical analysis, and advanced reporting while maintaining the clean architecture and beautiful user experience established in Phase 3.

## Risk Mitigation

### Technical Risks

- **Calculation Complexity**: Extensive unit testing and validation against known scenarios
- **UI Complexity**: Incremental development with user feedback loops
- **Performance**: Early optimization and benchmarking

### Business Risks

- **User Adoption**: Focus on intuitive design and comprehensive help system
- **Data Accuracy**: Multiple validation layers and audit trails
- **Scalability**: Design for growth from the beginning

### Mitigation Strategies

- Test-driven development for all financial calculations
- Comprehensive error handling and recovery mechanisms
- Regular performance profiling and optimization
- User testing with real NEPSE portfolio data
- Extensive documentation and help system

This Phase 3 implementation will deliver a **production-ready portfolio management TUI** that combines **functional excellence** with **beautiful design**, providing NEPSE investors with a powerful tool for tracking and managing their investments with precision and confidence.

