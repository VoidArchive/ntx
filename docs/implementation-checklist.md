# NTX Implementation Checklist

## Phase 1: Foundation (Week 1-2)

### Project Setup

- [ ] Initialize Go module: `go mod init ntx`
- [ ] Create directory structure: `internal/`, `cmd/`, `docs/`
- [ ] Setup dependencies: bubbletea, sqlx, goose, sqlite3
- [ ] Create `main.go` with basic CLI structure
- [ ] Setup `.gitignore` for Go projects

### CSV Parser (internal/csv/parser.go)

- [ ] Define `Transaction` struct with all required fields
- [ ] Implement `ParseCSV(filename string) ([]Transaction, error)`
- [ ] Handle transaction type detection (IPO, Bonus, Regular)
- [ ] Add validation for required fields
- [ ] Test with actual Meroshare CSV file
- [ ] Handle edge cases: empty rows, malformed data

### Database Layer (internal/db/)

- [ ] Create `models.go` with Transaction struct and database tags
- [ ] Setup SQLite connection with proper configuration
- [ ] Create goose migration for transactions table
- [ ] Use SQLC to generate the database from goose
- [ ] Implement `InsertTransaction(tx Transaction) error`
- [ ] Implement `GetAllTransactions() ([]Transaction, error)`
- [ ] Implement `GetTransactionsByScrip(scrip string) ([]Transaction, error)`
- [ ] Add proper error handling and logging

### WAC Calculator (internal/wac/calculator.go)

- [ ] Define `ShareLot` struct (Quantity, Price, Date)
- [ ] Define `Holding` struct (Scrip, TotalQuantity, WAC, Lots)
- [ ] Implement `CalculateHoldings(transactions []Transaction) ([]Holding, error)`
- [ ] Handle buy transactions: add to lots queue
- [ ] Handle sell transactions: remove from lots queue (FIFO)
- [ ] Handle partial lot consumption correctly
- [ ] Calculate WAC from remaining lots
- [ ] Test with complex scenarios (multiple buys/sells)

### Price Input TUI (internal/tui/price_input.go)

- [ ] Create price input form using bubbletea
- [ ] Group transactions by scrip for batch entry
- [ ] Display transaction details (date, quantity, type)
- [ ] Validate price inputs (positive numbers)
- [ ] Save entered prices to database
- [ ] Show progress indicator during batch entry

### Portfolio Display TUI (internal/tui/portfolio.go)

- [ ] Create table view for current holdings
- [ ] Display: Scrip, Quantity, WAC, Current Value
- [ ] Handle window resizing gracefully
- [ ] Add keyboard navigation (up/down arrows)
- [ ] Show total portfolio value
- [ ] Add refresh functionality

### Integration & Testing

- [ ] Connect all components in main.go
- [ ] Test complete workflow: CSV → DB → Price Input → Portfolio Display
- [ ] Add unit tests for WAC calculator
- [ ] Add integration tests with sample data
- [ ] Test error scenarios: missing files, invalid data
- [ ] Create sample CSV for testing

### Documentation & Polish

- [ ] Add command-line help and usage instructions
- [ ] Create README with installation and usage guide
- [ ] Add logging for debugging
- [ ] Handle graceful shutdown (Ctrl+C)
- [ ] Test on different terminal sizes

## Phase 2: Real-time Data (Week 3-4)

### Price Provider Interface (internal/price/)

- [ ] Define `PriceProvider` interface
- [ ] Implement `ShareSansarScraper` struct
- [ ] Add HTTP client with proper timeouts
- [ ] Implement rate limiting (avoid overwhelming server)
- [ ] Add error handling for network failures
- [ ] Create mock provider for testing

### ShareSansar Scraper

- [ ] Research ShareSansar HTML structure for prices
- [ ] Implement `GetPrice(scrip string) (float64, error)`
- [ ] Implement `GetPrices(scrips []string) (map[string]float64, error)`
- [ ] Add HTML parsing with goquery
- [ ] Handle scrips not found gracefully
- [ ] Add retry logic for failed requests

### Caching Layer (internal/cache/)

- [ ] Implement in-memory cache with expiration
- [ ] Cache prices for configurable duration (5-15 minutes)
- [ ] Add cache hit/miss metrics
- [ ] Implement cache invalidation
- [ ] Persist cache to disk (optional)

### Concurrent Price Fetching

- [ ] Implement worker pool pattern for concurrent requests
- [ ] Add semaphore to limit concurrent requests
- [ ] Use context for request cancellation
- [ ] Handle partial failures gracefully
- [ ] Add progress indicators for batch operations

### Enhanced Portfolio TUI

- [ ] Add real-time price updates
- [ ] Show current market value vs. cost
- [ ] Display profit/loss (absolute and percentage)
- [ ] Color-code gains (green) and losses (red)
- [ ] Add loading indicators during price fetching
- [ ] Implement auto-refresh functionality

### Multi-Panel Dashboard

- [ ] Create navigation between different views
- [ ] Implement portfolio overview panel
- [ ] Add transaction history panel
- [ ] Create performance summary panel
- [ ] Add keyboard shortcuts for navigation
- [ ] Handle panel switching smoothly

### Configuration System

- [ ] Create config file structure (YAML/JSON)
- [ ] Add scraping settings (intervals, timeouts)
- [ ] Add UI preferences (colors, refresh rates)
- [ ] Implement config validation
- [ ] Support config file in user home directory

### Error Handling & Resilience

- [ ] Implement circuit breaker pattern
- [ ] Add graceful degradation for network failures
- [ ] Create comprehensive error messages
- [ ] Add retry mechanisms with exponential backoff
- [ ] Log errors for debugging

### Testing & Performance

- [ ] Add unit tests for scraper with mock responses
- [ ] Test concurrent fetching performance
- [ ] Add integration tests with real ShareSansar
- [ ] Test error scenarios: network down, invalid responses
- [ ] Profile memory usage and optimize

## Phase 3: Advanced Analytics (Week 5-8)

### Historical Data Collection

- [ ] Extend database schema for price history
- [ ] Implement daily price collection job
- [ ] Add backfill functionality for historical data
- [ ] Create scheduled tasks for automatic collection
- [ ] Handle missing data points gracefully

### Portfolio Analytics Engine

- [ ] Implement performance metrics calculation
- [ ] Calculate total return, annualized return
- [ ] Add volatility and Sharpe ratio calculations
- [ ] Implement maximum drawdown calculation
- [ ] Create sector allocation analysis
- [ ] Add portfolio comparison features

### Advanced Visualization

- [ ] Implement ASCII chart rendering
- [ ] Create line charts for portfolio value over time
- [ ] Add bar charts for sector allocation
- [ ] Implement sparklines for individual stocks
- [ ] Create interactive chart navigation
- [ ] Add zoom and pan functionality

### Risk Analysis Tools

- [ ] Implement Value at Risk (VaR) calculations
- [ ] Calculate correlation matrix between holdings
- [ ] Add concentration risk analysis
- [ ] Implement portfolio beta calculation
- [ ] Create risk-adjusted return metrics

### Portfolio Optimization

- [ ] Implement rebalancing suggestions
- [ ] Add portfolio optimization algorithms
- [ ] Create target allocation comparisons
- [ ] Implement efficient frontier calculations
- [ ] Add diversification recommendations

### Reporting System

- [ ] Create monthly/quarterly report generation
- [ ] Implement PDF report export
- [ ] Add email reporting capabilities
- [ ] Create customizable report templates
- [ ] Add performance benchmarking

### Data Export/Import

- [ ] Implement CSV export for all data
- [ ] Add Excel export functionality
- [ ] Create backup/restore features
- [ ] Implement data migration tools
- [ ] Add API for third-party integrations

### Plugin Architecture

- [ ] Design plugin interface system
- [ ] Create sample plugins for data providers
- [ ] Implement plugin discovery and loading
- [ ] Add plugin configuration management
- [ ] Create plugin documentation

### Advanced TUI Features

- [ ] Implement tabbed interface
- [ ] Add search and filtering capabilities
- [ ] Create customizable dashboards
- [ ] Add drag-and-drop for panel arrangement
- [ ] Implement theme system

### Performance & Scalability

- [ ] Optimize database queries with indexes
- [ ] Implement connection pooling
- [ ] Add data streaming for large datasets
- [ ] Optimize memory usage
- [ ] Add performance monitoring

### Final Testing & Documentation

- [ ] Comprehensive integration testing
- [ ] Performance testing with large portfolios
- [ ] User acceptance testing
- [ ] Create complete user documentation
- [ ] Add developer documentation for plugins

## Quality Assurance Checklist

### Code Quality

- [ ] All functions have proper error handling
- [ ] Code follows Go best practices and conventions
- [ ] All public functions have documentation comments
- [ ] No hardcoded values (use configuration)
- [ ] Proper logging throughout the application

### Testing Coverage

- [ ] Unit tests for all business logic
- [ ] Integration tests for database operations
- [ ] End-to-end tests for complete workflows
- [ ] Test coverage above 80%
- [ ] Performance benchmarks for critical functions

### User Experience

- [ ] Intuitive keyboard navigation
- [ ] Clear error messages
- [ ] Responsive UI (handles window resizing)
- [ ] Consistent visual design
- [ ] Help documentation accessible within app

### Security & Privacy

- [ ] No sensitive data in logs
- [ ] Secure handling of financial data
- [ ] Input validation for all user inputs
- [ ] Safe file operations
- [ ] Proper cleanup of temporary files

### Deployment & Distribution

- [ ] Cross-platform compatibility (Linux, macOS, Windows)
- [ ] Build scripts for different architectures
- [ ] Installation instructions
- [ ] Version management and release notes
- [ ] Package for different distribution methods

## Success Metrics

### Phase 1 Success

- [ ] Can import Meroshare CSV without errors
- [ ] Successfully enter prices for all transactions
- [ ] Portfolio displays correct holdings and WAC
- [ ] FIFO calculations match manual verification

### Phase 2 Success

- [ ] Prices update automatically from ShareSansar
- [ ] UI remains responsive during price fetching
- [ ] Real-time portfolio value updates correctly
- [ ] Handles network failures gracefully

### Phase 3 Success

- [ ] Generates meaningful portfolio analytics
- [ ] Charts and visualizations display correctly
- [ ] Reports provide actionable insights
- [ ] System handles large portfolios efficiently

## Launch Readiness Checklist

### Pre-Launch

- [ ] All critical bugs fixed
- [ ] Performance meets requirements
- [ ] Documentation complete
- [ ] Installation tested on clean systems
- [ ] Backup and recovery tested

### Launch

- [ ] Create GitHub repository
- [ ] Write compelling README
- [ ] Create release packages
- [ ] Announce in relevant communities
- [ ] Gather initial user feedback

### Post-Launch

- [ ] Monitor for issues and bugs
- [ ] Respond to user feedback
- [ ] Plan future enhancements
- [ ] Maintain documentation
- [ ] Regular security updates
