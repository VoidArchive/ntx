# Phase 2: Enhanced Features - Real-time Data & Analytics

## Learning Objectives
By the end of Phase 2, you will understand:
- Web scraping and API integration patterns
- Concurrent programming with goroutines
- Data caching strategies
- Advanced TUI components and navigation
- Financial calculations (P&L, allocation, etc.)

## Core Concept: What is NTX Phase 2?

**Building on Phase 1**: You now have a working portfolio tracker. Phase 2 transforms it into a **real-time financial dashboard** that automatically fetches current prices and provides detailed analytics.

## Socratic Questions for Deep Understanding

### Question 1: Data Freshness
**Q**: If you're fetching stock prices from ShareSansar every time you open NTX, what problems might you encounter?

**Think About**:
- How often do stock prices change?
- What happens if ShareSansar is down?
- Is it fair to their servers to scrape every second?
- What if you have 50 different stocks?

**Your Answer**: _______

**Follow-up**: How would you design a caching system that balances freshness with performance?

### Question 2: User Experience
**Q**: When prices are being fetched, what should the user see and experience?

**Think About**:
- Should the app freeze while fetching data?
- How do you show loading states?
- What if some prices fail to load?
- Should outdated prices be shown or hidden?

**Your Answer**: _______

**Follow-up**: How would you implement non-blocking price updates in Go?

### Question 3: Data Reliability
**Q**: What happens when ShareSansar changes their website structure?

**Think About**:
- How do you detect when scraping breaks?
- Should you have fallback data sources?
- How do you test scraping code?
- What's your monitoring strategy?

**Your Answer**: _______

**Follow-up**: How would you design a robust scraping system that adapts to changes?

### Question 4: Portfolio Analytics
**Q**: Beyond current value, what insights would be most valuable for a portfolio manager?

**Think About**:
- Which stocks are performing best/worst?
- How is your money allocated across sectors?
- What's your overall return since inception?
- Which purchases were most profitable?

**Your Answer**: _______

**Follow-up**: How would you calculate and display these metrics efficiently?

## Feature Specifications

### 1. Real-time Price Integration

#### ShareSansar Scraper
```go
type PriceProvider interface {
    GetPrice(scrip string) (float64, error)
    GetPrices(scrips []string) (map[string]float64, error)
}

type ShareSansarScraper struct {
    client   *http.Client
    cache    *PriceCache
    rateLimit *RateLimiter
}
```

**Key Design Questions**:
- How do you handle rate limiting?
- What's your retry strategy for failed requests?
- How do you parse HTML reliably?
- What's your fallback when scraping fails?

### 2. Concurrent Data Fetching

#### Goroutine Pool Pattern
```go
type PriceFetcher struct {
    workers    int
    jobs       chan string
    results    chan PriceResult
    semaphore  chan struct{}
}

func (pf *PriceFetcher) FetchPrices(scrips []string) {
    // Implementation using worker pool
}
```

**Key Design Questions**:
- How many concurrent requests are appropriate?
- How do you prevent overwhelming the target server?
- What's your timeout strategy?
- How do you handle partial failures?

### 3. Advanced Portfolio Calculations

#### Profit/Loss Calculations
```go
type PerformanceCalculator struct {
    transactions []Transaction
    currentPrices map[string]float64
}

func (pc *PerformanceCalculator) CalculateUnrealizedPL(scrip string) (float64, error) {
    // Current value vs WAC
}

func (pc *PerformanceCalculator) CalculateRealizedPL(scrip string) (float64, error) {
    // Actual gains from sales
}
```

**Key Design Questions**:
- How do you handle different lot sales for P&L calculation?
- What's the difference between realized and unrealized gains?
- How do you account for dividends and bonuses?
- What's your approach to currency conversion (if needed)?

### 4. Enhanced TUI Components

#### Multi-Panel Dashboard
```go
type DashboardModel struct {
    portfolioView   PortfolioView
    performanceView PerformanceView
    sectorView      SectorView
    transactionView TransactionView
    activePanel     Panel
}
```

**Key Design Questions**:
- How do users navigate between panels?
- What keyboard shortcuts make sense?
- How do you handle window resizing?
- What's your color scheme for gains/losses?

## Phase 2 Implementation Steps

### 1. Price Data Layer
```bash
# Add new dependencies
go get github.com/PuerkitoBio/goquery  # HTML parsing
go get golang.org/x/time/rate          # Rate limiting
go get github.com/patrickmn/go-cache   # In-memory cache
```

### 2. Scraper Development
- **Start with**: Single stock price scraping
- **Ask**: How do you identify the price element on ShareSansar?
- **Test**: What happens when the stock is not found?
- **Extend**: Batch scraping with concurrency

### 3. Cache Implementation
- **Design**: When do you refresh cached prices?
- **Ask**: How do you handle cache misses?
- **Test**: What's your cache invalidation strategy?

### 4. TUI Enhancements
- **Start with**: Loading indicators
- **Ask**: How do you show real-time updates?
- **Test**: What happens with slow network?
- **Extend**: Multi-panel navigation

## Architecture Patterns

### 1. Repository Pattern
```go
type PriceRepository interface {
    GetPrice(scrip string) (Price, error)
    GetPrices(scrips []string) ([]Price, error)
    SavePrice(price Price) error
}

type CachedPriceRepository struct {
    scraper PriceProvider
    cache   Cache
    db      *sql.DB
}
```

### 2. Observer Pattern
```go
type PriceObserver interface {
    OnPriceUpdate(scrip string, price float64)
}

type PriceNotifier struct {
    observers []PriceObserver
}

func (pn *PriceNotifier) NotifyPriceUpdate(scrip string, price float64) {
    for _, observer := range pn.observers {
        observer.OnPriceUpdate(scrip, price)
    }
}
```

### 3. Command Pattern
```go
type Command interface {
    Execute() error
}

type RefreshPricesCommand struct {
    scrips []string
    repo   PriceRepository
}
```

## Error Handling Strategy

### 1. Graceful Degradation
```go
func (app *App) RefreshPrices() {
    for scrip := range app.holdings {
        price, err := app.priceRepo.GetPrice(scrip)
        if err != nil {
            // Log error but continue with cached/default price
            app.logger.Warn("Failed to fetch price", "scrip", scrip, "error", err)
            continue
        }
        app.updatePrice(scrip, price)
    }
}
```

### 2. Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    failureThreshold int
    resetTimeout     time.Duration
    state           State
    failures        int
}
```

## Testing Strategy

### 1. Mock Price Provider
```go
type MockPriceProvider struct {
    prices map[string]float64
    delay  time.Duration
}

func (m *MockPriceProvider) GetPrice(scrip string) (float64, error) {
    time.Sleep(m.delay)
    if price, exists := m.prices[scrip]; exists {
        return price, nil
    }
    return 0, errors.New("price not found")
}
```

### 2. Integration Tests
```go
func TestRealPriceFetching(t *testing.T) {
    scraper := NewShareSansarScraper()
    price, err := scraper.GetPrice("NABIL")
    assert.NoError(t, err)
    assert.Greater(t, price, 0.0)
}
```

## Performance Considerations

### 1. Concurrent Processing
**Question**: How many goroutines should you use for price fetching?
**Answer**: Consider server load, network latency, and your portfolio size

### 2. Memory Management
**Question**: How do you prevent memory leaks with long-running goroutines?
**Answer**: Use context cancellation and proper goroutine lifecycle management

### 3. Database Optimization
**Question**: How do you optimize database queries for large portfolios?
**Answer**: Use indexing, batch operations, and connection pooling

## Success Criteria

At the end of Phase 2, you should be able to:
- [ ] Automatically fetch current prices for all holdings
- [ ] See real-time portfolio value updates
- [ ] View profit/loss for individual holdings
- [ ] Navigate between different dashboard panels
- [ ] Handle network failures gracefully
- [ ] Understand concurrent programming patterns

## Self-Assessment Questions

Before moving to Phase 3, ask yourself:
1. How would you debug a failing price scraper?
2. What happens when you have 100 different stocks?
3. How do you ensure data consistency across goroutines?
4. Can you explain the observer pattern implementation?
5. What's your strategy for handling API rate limits?

## Next Steps

Once Phase 2 is complete, you'll be ready for Phase 3: Advanced Features
- Portfolio optimization algorithms
- Historical data analysis
- Advanced charting and visualization
- Export capabilities
- Multi-user support