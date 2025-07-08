package wac

import (
	"fmt"
	"sort"
	"time"

	"ntx/internal/csv"
	"ntx/internal/money"
)

// ShareLot represents a lot of shares purchased at a specific price and date.
// This is used to track individual purchase lots for FIFO calculation.
type ShareLot struct {
	Quantity int         // Number of shares in this lot
	Price    money.Money // Price per share when purchased
	Date     time.Time   // Date when purchased
}

// Holding represents the current holding for a specific scrip.
// It contains the total quantity, weighted average cost, and individual lots.
type Holding struct {
	Scrip         string      // Stock symbol (e.g., "API", "NMB")
	TotalQuantity int         // Total shares currently held
	WAC           money.Money // Weighted Average Cost per share
	Lots          []ShareLot  // Individual lots that make up this holding
}

// IsEmpty returns true if the holding has no shares
func (h Holding) IsEmpty() bool {
	return h.TotalQuantity == 0
}

// TotalValue returns the total value of the holding (quantity * WAC)
func (h Holding) TotalValue() money.Money {
	if h.TotalQuantity == 0 {
		return money.Money(0)
	}
	return h.WAC.MultiplyInt(h.TotalQuantity)
}

// Calculator implements the FIFO (First In, First Out) algorithm
// for calculating portfolio holdings and weighted average cost.
type Calculator struct {
	// Holdings maps scrip name to current holding
	holdings map[string]*Holding
}

// NewCalculator creates a new FIFO calculator
func NewCalculator() *Calculator {
	return &Calculator{
		holdings: make(map[string]*Holding),
	}
}

// CalculateHoldings processes a list of transactions and returns current holdings
// using the FIFO method for cost calculation.
func (c *Calculator) CalculateHoldings(transactions []csv.Transaction) ([]Holding, error) {
	// Reset holdings
	c.holdings = make(map[string]*Holding)
	
	// Sort transactions by date (oldest first) for proper FIFO processing
	sortedTransactions := make([]csv.Transaction, len(transactions))
	copy(sortedTransactions, transactions)
	sort.Slice(sortedTransactions, func(i, j int) bool {
		if sortedTransactions[i].Date.Equal(sortedTransactions[j].Date) {
			// If same date, process in consistent order (could use ID if available)
			return sortedTransactions[i].Scrip < sortedTransactions[j].Scrip
		}
		return sortedTransactions[i].Date.Before(sortedTransactions[j].Date)
	})
	
	// Process each transaction
	for _, tx := range sortedTransactions {
		if err := c.processTransaction(tx); err != nil {
			return nil, fmt.Errorf("error processing transaction %s: %w", tx.Scrip, err)
		}
	}
	
	// Convert holdings map to slice, excluding empty holdings
	result := make([]Holding, 0, len(c.holdings))
	for _, holding := range c.holdings {
		if !holding.IsEmpty() {
			result = append(result, *holding)
		}
	}
	
	// Sort result by scrip name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Scrip < result[j].Scrip
	})
	
	return result, nil
}

// processTransaction processes a single transaction (buy or sell)
func (c *Calculator) processTransaction(tx csv.Transaction) error {
	// Get or create holding for this scrip
	holding := c.getOrCreateHolding(tx.Scrip)
	
	if tx.IsBuy() {
		return c.processBuyTransaction(holding, tx)
	} else if tx.IsSell() {
		return c.processSellTransaction(holding, tx)
	}
	
	// Skip transactions with zero quantity (shouldn't happen if validated)
	return nil
}

// getOrCreateHolding gets existing holding or creates new one for scrip
func (c *Calculator) getOrCreateHolding(scrip string) *Holding {
	if holding, exists := c.holdings[scrip]; exists {
		return holding
	}
	
	holding := &Holding{
		Scrip:         scrip,
		TotalQuantity: 0,
		WAC:           money.Money(0),
		Lots:          []ShareLot{},
	}
	c.holdings[scrip] = holding
	return holding
}

// processBuyTransaction handles buy transactions by adding new lots
func (c *Calculator) processBuyTransaction(holding *Holding, tx csv.Transaction) error {
	// For bonus shares, rights, etc., price might be zero
	price := tx.Price
	
	// Create new lot
	newLot := ShareLot{
		Quantity: tx.AbsQuantity(),
		Price:    price,
		Date:     tx.Date,
	}
	
	// Add lot to the end of the queue (FIFO - newest goes to back)
	holding.Lots = append(holding.Lots, newLot)
	holding.TotalQuantity += newLot.Quantity
	
	// Recalculate WAC
	holding.WAC = c.calculateWAC(holding.Lots)
	
	return nil
}

// processSellTransaction handles sell transactions using FIFO method
func (c *Calculator) processSellTransaction(holding *Holding, tx csv.Transaction) error {
	sellQuantity := tx.AbsQuantity()
	
	// Check if we have enough shares to sell
	if holding.TotalQuantity < sellQuantity {
		return fmt.Errorf("insufficient shares: trying to sell %d but only have %d", 
			sellQuantity, holding.TotalQuantity)
	}
	
	// Remove shares from lots using FIFO (oldest first)
	remainingToSell := sellQuantity
	var newLots []ShareLot
	
	for _, lot := range holding.Lots {
		if remainingToSell == 0 {
			// Keep this lot as-is
			newLots = append(newLots, lot)
		} else if lot.Quantity <= remainingToSell {
			// Consume entire lot
			remainingToSell -= lot.Quantity
			// Don't add this lot to newLots (it's fully consumed)
		} else {
			// Partial consumption - split the lot
			consumedQuantity := remainingToSell
			remainingInLot := lot.Quantity - consumedQuantity
			
			// Keep the remaining portion of this lot
			newLots = append(newLots, ShareLot{
				Quantity: remainingInLot,
				Price:    lot.Price,
				Date:     lot.Date,
			})
			
			remainingToSell = 0
		}
	}
	
	// Update holding with new lots
	holding.Lots = newLots
	holding.TotalQuantity -= sellQuantity
	
	// Recalculate WAC from remaining lots
	holding.WAC = c.calculateWAC(holding.Lots)
	
	return nil
}

// calculateWAC calculates the weighted average cost from a list of lots
func (c *Calculator) calculateWAC(lots []ShareLot) money.Money {
	if len(lots) == 0 {
		return money.Money(0)
	}
	
	totalValue := money.Money(0)
	totalQuantity := 0
	
	for _, lot := range lots {
		// Only include lots with valid prices (non-zero)
		if !lot.Price.IsZero() {
			lotValue := lot.Price.MultiplyInt(lot.Quantity)
			totalValue = totalValue.Add(lotValue)
			totalQuantity += lot.Quantity
		}
	}
	
	if totalQuantity == 0 {
		return money.Money(0)
	}
	
	// Calculate weighted average
	return totalValue.DivideInt(totalQuantity)
}

// GetHolding returns the current holding for a specific scrip
func (c *Calculator) GetHolding(scrip string) (Holding, bool) {
	if holding, exists := c.holdings[scrip]; exists {
		return *holding, true
	}
	return Holding{}, false
}

// GetAllHoldings returns all current holdings
func (c *Calculator) GetAllHoldings() []Holding {
	holdings, _ := c.CalculateHoldings([]csv.Transaction{})
	return holdings
}