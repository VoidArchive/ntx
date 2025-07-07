# Phase 1: Foundation - CSV Import & Basic Portfolio Display

## Learning Objectives

By the end of Phase 1, you will understand:

- How to parse CSV files in Go
- SQLite database design for financial data
- FIFO WAC calculation algorithm
- Basic bubbletea TUI structure
- Single source of truth principle

## Core Concept: What is NTX Phase 1?

**Simple Answer**: A program that reads your Meroshare transaction CSV, asks you for missing prices, calculates your current holdings using FIFO method, and shows them in a terminal table.

**Explain Like I'm 5**:

- You have a CSV file of all your stock transactions
- Some transactions don't have prices (you need to enter them)
- The program calculates how much of each stock you own
- It shows you this in a nice table in your terminal

## The Data Flow (Step by Step)

### Step 1: CSV Import

```
Meroshare CSV → Parse → Validate → Store in SQLite
```

**Key Questions to Ask Yourself:**

- What happens if the CSV has invalid data?
- How do you handle different transaction types (IPO, Bonus, Regular)?
- What if a row is missing required fields?

### Step 2: Price Collection

```
Missing Prices → Group by Scrip → User Input → Update Database
```

**Key Questions to Ask Yourself:**

- How do you identify which transactions need prices?
- What's the most efficient way to collect prices from user?
- How do you validate price inputs?

### Step 3: FIFO WAC Calculation

```
Transactions → Sort by Date → Process Queue → Calculate WAC
```

**Key Questions to Ask Yourself:**

- What data structure represents your holdings?
- How do you handle partial lot sales?
- What happens when you sell more shares than you own?

### Step 4: Display

```
Holdings → Format → TUI Table → User Interaction
```

**Key Questions to Ask Yourself:**

- What information should be shown in the table?
- How do you handle window resizing?
- What keyboard shortcuts make sense?

## Database Schema

### Single Table Approach

```sql
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scrip TEXT NOT NULL,
    date TEXT NOT NULL,
    quantity INTEGER NOT NULL,  -- positive for buy, negative for sell
    price REAL,                 -- price per share (NULL if not entered)
    transaction_type TEXT NOT NULL, -- 'IPO', 'BONUS', 'REGULAR', 'RIGHTS'
    description TEXT,           -- original meroshare description
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Why This Design?**

- **Single source of truth**: All data lives in one place
- **Consistent**: No risk of holdings getting out of sync
- **Auditable**: You can trace every calculation back to transactions
- **Flexible**: Easy to add new transaction types

## FIFO Algorithm (Explained Simply)

### The Queue Concept

Think of your holdings as a queue at a store:

- **Buy shares** = People join the back of the queue
- **Sell shares** = People leave from the front of the queue
- **First In, First Out** = Oldest shares get sold first

### Example Implementation

```go
type ShareLot struct {
    Quantity int
    Price    float64
    Date     time.Time
}

// For each scrip, maintain a queue of lots
queue := []ShareLot{
    {41, 295.50, date1},  // Oldest purchase
    {20, 302.00, date2},
    {50, 310.00, date3},  // Newest purchase
}

// When selling 30 shares:
// 1. Take 30 from first lot (41 shares)
// 2. First lot now has 11 shares remaining
// 3. WAC = weighted average of remaining lots
```

## Phase 1 Implementation Steps

### 1. Project Setup

```bash
mkdir ntx
cd ntx
go mod init ntx
```

### 2. Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/pressly/goose/v3
go get github.com/mattn/go-sqlite3
```

### 3. File Structure

```
ntx/
├── main.go
├── internal/
│   ├── csv/
│   │   └── parser.go
│   ├── db/
│   │   ├── models.go
│   │   ├── queries.go
│   │   └── migrations/
│   ├── fifo/
│   │   └── calculator.go
│   └── tui/
│       └── app.go
└── docs/
    ├── phase1-foundation.md
    ├── phase2-enhancements.md
    └── phase3-realtime.md
```

### 4. Development Order

1. **CSV Parser** - Read and validate Meroshare CSV
2. **Database** - Setup SQLite with migrations
3. **Models** - Define transaction and holdings structs
4. **FIFO Calculator** - Implement WAC calculation
5. **TUI** - Basic table display with bubbletea
6. **Integration** - Connect all components

## Key Go Concepts You'll Learn

### 1. Error Handling

```go
// Go's explicit error handling
data, err := csv.ReadFile("transactions.csv")
if err != nil {
    return fmt.Errorf("failed to read CSV: %w", err)
}
```

### 2. Struct Methods

```go
type Transaction struct {
    Scrip    string
    Date     time.Time
    Quantity int
    Price    float64
}

func (t Transaction) IsBuy() bool {
    return t.Quantity > 0
}
```

### 3. Interfaces

```go
type Calculator interface {
    CalculateWAC(scrip string) (float64, error)
}

type FIFOCalculator struct {
    db *sql.DB
}

func (f *FIFOCalculator) CalculateWAC(scrip string) (float64, error) {
    // Implementation
}
```

## Success Criteria

At the end of Phase 1, you should be able to:

- [ ] Import your Meroshare CSV file
- [ ] Enter missing transaction prices
- [ ] See your current holdings in a terminal table
- [ ] Verify WAC calculations match your expectations
- [ ] Understand the FIFO algorithm conceptually

## Common Pitfalls to Avoid

1. **Don't create a separate holdings table** - Calculate from transactions
2. **Don't forget to handle partial lot sales** - Update lot quantities
3. **Don't hardcode CSV structure** - Make it flexible
4. **Don't ignore error handling** - Fail gracefully with clear messages
5. **Don't over-engineer** - Keep it simple for Phase 1

## Questions for Self-Assessment

Before moving to Phase 2, ask yourself:

1. Can you explain FIFO to someone else without looking at notes?
2. Do you understand why WAC is important for tax calculations?
3. Can you trace through the data flow from CSV to display?
4. Do you know what happens when you add a new transaction?
5. Can you modify the code to add a new transaction type?

## Next Steps

Once Phase 1 is complete, you'll be ready for Phase 2: Enhanced Features

- Real-time price integration
- Profit/Loss calculations
- Sector-wise allocation
- Portfolio analytics
