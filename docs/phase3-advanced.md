# Phase 3: Advanced Features - Analytics & Optimization

## Learning Objectives
By the end of Phase 3, you will understand:
- Historical data analysis and time series processing
- Portfolio optimization algorithms
- Advanced Go patterns (plugins, reflection, generics)
- Data visualization in terminal applications
- Export/import capabilities
- System architecture for scalability

## Core Concept: What is NTX Phase 3?

**Building on Phase 2**: You now have a real-time portfolio tracker. Phase 3 transforms it into a **professional-grade portfolio management system** with advanced analytics, optimization suggestions, and comprehensive reporting.

## Step-by-Step Implementation Guide

### Step 1: Historical Data Foundation

#### 1.1 Database Schema Extension
```sql
-- Add historical price tracking
CREATE TABLE price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scrip TEXT NOT NULL,
    price REAL NOT NULL,
    date TEXT NOT NULL,
    volume INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scrip, date)
);

-- Add portfolio snapshots
CREATE TABLE portfolio_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    snapshot_date TEXT NOT NULL,
    total_value REAL NOT NULL,
    total_cost REAL NOT NULL,
    holdings_json TEXT NOT NULL, -- JSON of holdings at that time
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.2 Historical Data Collection
```go
type HistoricalDataCollector struct {
    scraper     PriceProvider
    db          *sql.DB
    scheduler   *cron.Cron
}

func (hdc *HistoricalDataCollector) CollectDailyPrices() error {
    // Step 1: Get all unique scrips from portfolio
    scrips, err := hdc.getAllScrips()
    if err != nil {
        return err
    }
    
    // Step 2: Fetch current prices
    prices, err := hdc.scraper.GetPrices(scrips)
    if err != nil {
        return err
    }
    
    // Step 3: Store in price_history table
    return hdc.storePriceHistory(prices)
}
```

**Implementation Steps**:
1. **First**: Add migration for new tables
2. **Second**: Implement daily price collection
3. **Third**: Add backfill functionality for historical data
4. **Fourth**: Create scheduled jobs for automatic collection

### Step 2: Portfolio Analytics Engine

#### 2.1 Performance Metrics Calculator
```go
type PerformanceMetrics struct {
    TotalReturn         float64
    AnnualizedReturn    float64
    Volatility          float64
    SharpeRatio         float64
    MaxDrawdown         float64
    Alpha               float64
    Beta                float64
}

type AnalyticsEngine struct {
    db     *sql.DB
    market MarketDataProvider
}

func (ae *AnalyticsEngine) CalculateMetrics(startDate, endDate time.Time) (*PerformanceMetrics, error) {
    // Step 1: Get portfolio value history
    valueHistory, err := ae.getPortfolioValueHistory(startDate, endDate)
    if err != nil {
        return nil, err
    }
    
    // Step 2: Calculate returns
    returns := ae.calculateReturns(valueHistory)
    
    // Step 3: Calculate metrics
    metrics := &PerformanceMetrics{
        TotalReturn:      ae.calculateTotalReturn(returns),
        AnnualizedReturn: ae.calculateAnnualizedReturn(returns),
        Volatility:       ae.calculateVolatility(returns),
        SharpeRatio:      ae.calculateSharpeRatio(returns),
        MaxDrawdown:      ae.calculateMaxDrawdown(valueHistory),
    }
    
    return metrics, nil
}
```

**Implementation Steps**:
1. **First**: Implement basic return calculations
2. **Second**: Add volatility and risk metrics
3. **Third**: Implement benchmark comparisons
4. **Fourth**: Add sector-wise performance analysis

#### 2.2 Sector Allocation Analysis
```go
type SectorAllocation struct {
    Sector      string
    Value       float64
    Percentage  float64
    Count       int
    AvgReturn   float64
}

func (ae *AnalyticsEngine) CalculateSectorAllocation() ([]SectorAllocation, error) {
    // Step 1: Get current holdings
    holdings, err := ae.getCurrentHoldings()
    if err != nil {
        return nil, err
    }
    
    // Step 2: Map each scrip to sector
    sectorMap := make(map[string][]Holding)
    for _, holding := range holdings {
        sector := ae.getSectorForScrip(holding.Scrip)
        sectorMap[sector] = append(sectorMap[sector], holding)
    }
    
    // Step 3: Calculate allocation percentages
    totalValue := ae.calculateTotalValue(holdings)
    var allocations []SectorAllocation
    
    for sector, sectorHoldings := range sectorMap {
        sectorValue := ae.calculateSectorValue(sectorHoldings)
        allocation := SectorAllocation{
            Sector:     sector,
            Value:      sectorValue,
            Percentage: (sectorValue / totalValue) * 100,
            Count:      len(sectorHoldings),
        }
        allocations = append(allocations, allocation)
    }
    
    return allocations, nil
}
```

### Step 3: Advanced Visualization

#### 3.1 Terminal Charts
```go
type ChartRenderer struct {
    width  int
    height int
}

func (cr *ChartRenderer) RenderLineChart(data []float64, title string) string {
    // Step 1: Normalize data to fit chart dimensions
    normalizedData := cr.normalizeData(data)
    
    // Step 2: Create ASCII line chart
    chart := make([][]rune, cr.height)
    for i := range chart {
        chart[i] = make([]rune, cr.width)
        for j := range chart[i] {
            chart[i][j] = ' '
        }
    }
    
    // Step 3: Plot data points
    for i, value := range normalizedData {
        x := int(float64(i) * float64(cr.width-1) / float64(len(normalizedData)-1))
        y := int(value * float64(cr.height-1))
        y = cr.height - 1 - y // Flip Y axis
        chart[y][x] = '●'
    }
    
    // Step 4: Draw connecting lines
    cr.drawLines(chart, normalizedData)
    
    // Step 5: Add title and axes
    return cr.formatChart(chart, title)
}
```

**Implementation Steps**:
1. **First**: Implement basic line charts for portfolio value over time
2. **Second**: Add bar charts for sector allocation
3. **Third**: Create sparklines for individual stock performance
4. **Fourth**: Add interactive chart navigation

#### 3.2 Enhanced TUI Dashboard
```go
type AdvancedDashboard struct {
    model          tea.Model
    views          map[string]tea.Model
    activeView     string
    chartRenderer  *ChartRenderer
    analytics      *AnalyticsEngine
}

func (ad *AdvancedDashboard) Init() tea.Cmd {
    return tea.Batch(
        ad.loadPortfolioData,
        ad.loadAnalytics,
        ad.loadCharts,
    )
}

func (ad *AdvancedDashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "1":
            ad.activeView = "portfolio"
        case "2":
            ad.activeView = "analytics"
        case "3":
            ad.activeView = "charts"
        case "4":
            ad.activeView = "sectors"
        case "r":
            return ad, ad.refreshAllData
        }
    }
    
    // Update active view
    if view, exists := ad.views[ad.activeView]; exists {
        updatedView, cmd := view.Update(msg)
        ad.views[ad.activeView] = updatedView
        return ad, cmd
    }
    
    return ad, nil
}
```

### Step 4: Portfolio Optimization

#### 4.1 Rebalancing Suggestions
```go
type RebalancingSuggestion struct {
    Scrip          string
    CurrentWeight  float64
    TargetWeight   float64
    Action         string // "BUY", "SELL", "HOLD"
    Quantity       int
    EstimatedCost  float64
}

type PortfolioOptimizer struct {
    analytics *AnalyticsEngine
    constraints *OptimizationConstraints
}

type OptimizationConstraints struct {
    MaxSectorAllocation float64
    MinDiversification  int
    RiskTolerance      float64
    InvestmentAmount   float64
}

func (po *PortfolioOptimizer) SuggestRebalancing(target TargetAllocation) ([]RebalancingSuggestion, error) {
    // Step 1: Get current portfolio state
    currentHoldings, err := po.analytics.getCurrentHoldings()
    if err != nil {
        return nil, err
    }
    
    // Step 2: Calculate current weights
    currentWeights := po.calculateCurrentWeights(currentHoldings)
    
    // Step 3: Compare with target allocation
    suggestions := []RebalancingSuggestion{}
    
    for scrip, targetWeight := range target {
        currentWeight := currentWeights[scrip]
        deviation := targetWeight - currentWeight
        
        if math.Abs(deviation) > 0.05 { // 5% threshold
            suggestion := RebalancingSuggestion{
                Scrip:         scrip,
                CurrentWeight: currentWeight,
                TargetWeight:  targetWeight,
            }
            
            if deviation > 0 {
                suggestion.Action = "BUY"
                suggestion.Quantity = po.calculateBuyQuantity(scrip, deviation)
            } else {
                suggestion.Action = "SELL"
                suggestion.Quantity = po.calculateSellQuantity(scrip, -deviation)
            }
            
            suggestions = append(suggestions, suggestion)
        }
    }
    
    return suggestions, nil
}
```

#### 4.2 Risk Analysis
```go
type RiskAnalysis struct {
    VaR95          float64 // Value at Risk at 95% confidence
    VaR99          float64 // Value at Risk at 99% confidence
    ConditionalVaR float64 // Expected Shortfall
    Correlation    map[string]map[string]float64
    Concentration  float64 // Concentration risk
}

func (po *PortfolioOptimizer) AnalyzeRisk(holdings []Holding) (*RiskAnalysis, error) {
    // Step 1: Get historical returns for all holdings
    returns := make(map[string][]float64)
    for _, holding := range holdings {
        stockReturns, err := po.analytics.getHistoricalReturns(holding.Scrip, 252) // 1 year
        if err != nil {
            return nil, err
        }
        returns[holding.Scrip] = stockReturns
    }
    
    // Step 2: Calculate portfolio returns
    portfolioReturns := po.calculatePortfolioReturns(holdings, returns)
    
    // Step 3: Calculate VaR
    var95 := po.calculateVaR(portfolioReturns, 0.05)
    var99 := po.calculateVaR(portfolioReturns, 0.01)
    
    // Step 4: Calculate correlations
    correlations := po.calculateCorrelationMatrix(returns)
    
    // Step 5: Calculate concentration risk
    concentration := po.calculateConcentrationRisk(holdings)
    
    return &RiskAnalysis{
        VaR95:          var95,
        VaR99:          var99,
        ConditionalVaR: po.calculateConditionalVaR(portfolioReturns, 0.05),
        Correlation:    correlations,
        Concentration:  concentration,
    }, nil
}
```

### Step 5: Export and Reporting

#### 5.1 Report Generation
```go
type ReportGenerator struct {
    analytics *AnalyticsEngine
    templates map[string]*template.Template
}

func (rg *ReportGenerator) GenerateMonthlyReport(month time.Month, year int) (*Report, error) {
    startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)
    
    report := &Report{
        Period: fmt.Sprintf("%s %d", month.String(), year),
        GeneratedAt: time.Now(),
    }
    
    // Step 1: Portfolio performance
    metrics, err := rg.analytics.CalculateMetrics(startDate, endDate)
    if err != nil {
        return nil, err
    }
    report.Performance = metrics
    
    // Step 2: Top performers
    topPerformers, err := rg.analytics.GetTopPerformers(startDate, endDate, 5)
    if err != nil {
        return nil, err
    }
    report.TopPerformers = topPerformers
    
    // Step 3: Sector allocation
    sectorAllocation, err := rg.analytics.CalculateSectorAllocation()
    if err != nil {
        return nil, err
    }
    report.SectorAllocation = sectorAllocation
    
    // Step 4: Risk metrics
    riskAnalysis, err := rg.analytics.AnalyzeRisk()
    if err != nil {
        return nil, err
    }
    report.RiskAnalysis = riskAnalysis
    
    return report, nil
}
```

#### 5.2 Data Export
```go
type DataExporter struct {
    db *sql.DB
}

func (de *DataExporter) ExportToCSV(filename string, dateRange DateRange) error {
    // Step 1: Query transactions within date range
    transactions, err := de.getTransactions(dateRange)
    if err != nil {
        return err
    }
    
    // Step 2: Create CSV file
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // Step 3: Write headers
    headers := []string{"Date", "Scrip", "Type", "Quantity", "Price", "Total", "Balance"}
    writer.Write(headers)
    
    // Step 4: Write data
    for _, transaction := range transactions {
        record := []string{
            transaction.Date.Format("2006-01-02"),
            transaction.Scrip,
            transaction.Type,
            fmt.Sprintf("%d", transaction.Quantity),
            fmt.Sprintf("%.2f", transaction.Price),
            fmt.Sprintf("%.2f", transaction.Total),
            fmt.Sprintf("%.2f", transaction.Balance),
        }
        writer.Write(record)
    }
    
    return nil
}
```

### Step 6: System Architecture Improvements

#### 6.1 Plugin System
```go
type Plugin interface {
    Name() string
    Version() string
    Initialize() error
    Shutdown() error
}

type PriceProvider interface {
    Plugin
    GetPrice(scrip string) (float64, error)
    GetPrices(scrips []string) (map[string]float64, error)
}

type PluginManager struct {
    plugins map[string]Plugin
    config  *Config
}

func (pm *PluginManager) LoadPlugin(path string) error {
    // Step 1: Load plugin from file
    plugin, err := plugin.Open(path)
    if err != nil {
        return err
    }
    
    // Step 2: Look for required symbols
    symPlugin, err := plugin.Lookup("Plugin")
    if err != nil {
        return err
    }
    
    // Step 3: Type assertion
    pluginImpl, ok := symPlugin.(Plugin)
    if !ok {
        return errors.New("invalid plugin interface")
    }
    
    // Step 4: Initialize and register
    err = pluginImpl.Initialize()
    if err != nil {
        return err
    }
    
    pm.plugins[pluginImpl.Name()] = pluginImpl
    return nil
}
```

#### 6.2 Configuration Management
```go
type Config struct {
    Database struct {
        Path     string `yaml:"path"`
        Timeout  int    `yaml:"timeout"`
    } `yaml:"database"`
    
    Scraping struct {
        Enabled     bool          `yaml:"enabled"`
        Interval    time.Duration `yaml:"interval"`
        Timeout     time.Duration `yaml:"timeout"`
        RateLimit   int           `yaml:"rate_limit"`
        UserAgent   string        `yaml:"user_agent"`
    } `yaml:"scraping"`
    
    UI struct {
        Theme       string `yaml:"theme"`
        RefreshRate int    `yaml:"refresh_rate"`
    } `yaml:"ui"`
    
    Plugins struct {
        Directory string   `yaml:"directory"`
        Enabled   []string `yaml:"enabled"`
    } `yaml:"plugins"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## Implementation Checklist

### Phase 3.1: Historical Data & Analytics
- [ ] Extend database schema for historical data
- [ ] Implement daily price collection
- [ ] Create performance metrics calculator
- [ ] Add sector allocation analysis
- [ ] Implement risk analysis calculations

### Phase 3.2: Advanced Visualization
- [ ] Create terminal chart renderer
- [ ] Implement line charts for portfolio value
- [ ] Add bar charts for sector allocation
- [ ] Create sparklines for individual stocks
- [ ] Enhance TUI with multiple views

### Phase 3.3: Portfolio Optimization
- [ ] Implement rebalancing suggestions
- [ ] Add risk analysis tools
- [ ] Create optimization constraints
- [ ] Implement portfolio comparison features

### Phase 3.4: Reporting & Export
- [ ] Create report generation system
- [ ] Implement CSV export functionality
- [ ] Add PDF report generation
- [ ] Create email reporting (optional)

### Phase 3.5: System Architecture
- [ ] Implement plugin system
- [ ] Add configuration management
- [ ] Create proper logging system
- [ ] Implement graceful shutdown

## Testing Strategy

### Unit Tests
```go
func TestPerformanceCalculator(t *testing.T) {
    calc := NewPerformanceCalculator()
    
    // Test total return calculation
    returns := []float64{0.1, -0.05, 0.15, -0.02, 0.08}
    totalReturn := calc.calculateTotalReturn(returns)
    
    expected := 0.1 * 0.95 * 1.15 * 0.98 * 1.08 - 1
    assert.InDelta(t, expected, totalReturn, 0.001)
}
```

### Integration Tests
```go
func TestReportGeneration(t *testing.T) {
    // Setup test database with sample data
    db := setupTestDB()
    defer db.Close()
    
    generator := NewReportGenerator(db)
    report, err := generator.GenerateMonthlyReport(time.January, 2024)
    
    assert.NoError(t, err)
    assert.NotNil(t, report)
    assert.NotEmpty(t, report.Performance)
}
```

## Performance Optimization

### 1. Database Optimization
- Use indexes on frequently queried columns
- Implement connection pooling
- Use batch operations for bulk inserts
- Consider read replicas for analytics

### 2. Memory Management
- Implement data streaming for large datasets
- Use object pooling for frequent allocations
- Profile memory usage with pprof
- Consider memory-mapped files for large historical data

### 3. Concurrent Processing
- Use worker pools for CPU-intensive calculations
- Implement proper context cancellation
- Use sync.Pool for reusable objects
- Consider pipeline patterns for data processing

## Success Criteria

At the end of Phase 3, you should be able to:
- [ ] View comprehensive portfolio analytics
- [ ] Generate detailed performance reports
- [ ] Visualize data with terminal charts
- [ ] Get rebalancing suggestions
- [ ] Export data in multiple formats
- [ ] Analyze portfolio risk metrics
- [ ] Use plugin system for extensibility

## Final Assessment

Before considering NTX complete, ensure you can:
1. Explain the entire system architecture
2. Implement a new analytics metric
3. Create a custom visualization
4. Write a plugin for a new data source
5. Debug performance issues
6. Scale the system for larger portfolios

## Beyond Phase 3

Consider these advanced features:
- **Web Interface**: Browser-based dashboard
- **Mobile App**: React Native or Flutter companion
- **API Server**: RESTful API for third-party integrations
- **Machine Learning**: Predictive analytics and recommendations
- **Multi-Currency**: Support for international investments
- **Collaborative Features**: Portfolio sharing and discussions