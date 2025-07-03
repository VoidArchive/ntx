package validation

import (
	"fmt"
	"regexp"
	"time"

	"ntx/internal/portfolio/models"
)

// NEPSEValidator implements NEPSE-specific trading rules and constraints
type NEPSEValidator struct{}

// NewNEPSEValidator creates a new NEPSE validator
func NewNEPSEValidator() *NEPSEValidator {
	return &NEPSEValidator{}
}

// ValidateSymbol validates NEPSE stock symbol format and existence
func (v *NEPSEValidator) ValidateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	// NEPSE symbols: 3-6 characters, uppercase letters only
	if len(symbol) < 3 || len(symbol) > 6 {
		return fmt.Errorf("NEPSE symbol must be 3-6 characters long")
	}

	// Ensure uppercase and only letters
	symbolPattern := regexp.MustCompile(`^[A-Z]{3,6}$`)
	if !symbolPattern.MatchString(symbol) {
		return fmt.Errorf("NEPSE symbol must contain only uppercase letters (A-Z)")
	}

	// Check against known NEPSE symbols (basic validation)
	if !v.isKnownSymbol(symbol) {
		// NOTE: This is a warning, not an error - new symbols may be listed
		// Could be enhanced with real-time NEPSE API validation
	}

	return nil
}

// ValidateLotSize validates transaction quantity against NEPSE lot size requirements
func (v *NEPSEValidator) ValidateLotSize(symbol string, quantity int64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	lotSize := v.getSymbolLotSize(symbol)
	if quantity%lotSize != 0 {
		return fmt.Errorf("quantity must be in multiples of %d for %s (lot size: %d)", 
			lotSize, symbol, lotSize)
	}

	return nil
}

// ValidatePriceLimit validates price against NEPSE daily price band limits
func (v *NEPSEValidator) ValidatePriceLimit(symbol string, price models.Money, lastPrice models.Money) error {
	if price.IsZero() || price.IsNegative() {
		return fmt.Errorf("price must be positive")
	}

	// Skip validation if no previous price available
	if lastPrice.IsZero() {
		return nil
	}

	// NEPSE standard: ±10% daily price movement limit
	upperLimit := lastPrice.Multiply(1.10) // +10%
	lowerLimit := lastPrice.Multiply(0.90) // -10%

	if price.Compare(upperLimit) > 0 {
		return fmt.Errorf("price Rs.%s exceeds upper limit Rs.%s (+10%% from last price Rs.%s)", 
			price.FormatSimple(), upperLimit.FormatSimple(), lastPrice.FormatSimple())
	}

	if price.Compare(lowerLimit) < 0 {
		return fmt.Errorf("price Rs.%s below lower limit Rs.%s (-10%% from last price Rs.%s)", 
			price.FormatSimple(), lowerLimit.FormatSimple(), lastPrice.FormatSimple())
	}

	return nil
}

// ValidateTradingHours validates transaction time against NEPSE trading hours
func (v *NEPSEValidator) ValidateTradingHours(transactionDate time.Time) error {
	// NEPSE trading hours: Sunday-Thursday, 11:00-15:00 NPT
	
	// Check if date is in the future
	now := time.Now()
	if transactionDate.After(now) {
		return fmt.Errorf("transaction date cannot be in the future")
	}

	// Check day of week (Friday = 5, Saturday = 6 are non-trading days)
	weekday := transactionDate.Weekday()
	if weekday == time.Friday || weekday == time.Saturday {
		return fmt.Errorf("NEPSE is closed on %s (trading days: Sunday-Thursday)", weekday.String())
	}

	// Check trading hours for all transactions
	hour := transactionDate.Hour()
	if hour < 11 || hour >= 15 {
		return fmt.Errorf("transaction time %02d:00 outside NEPSE trading hours (11:00-15:00 NPT)", hour)
	}

	return nil
}

// ValidateMinimumTransaction validates minimum transaction value
func (v *NEPSEValidator) ValidateMinimumTransaction(quantity int64, price models.Money) error {
	transactionValue := price.MultiplyInt(quantity)
	minimumValue := models.NewMoney(500) // Rs. 500 minimum

	if transactionValue.Compare(minimumValue) < 0 {
		return fmt.Errorf("minimum transaction value is Rs.500 (current: Rs.%s)", 
			transactionValue.FormatSimple())
	}

	return nil
}

// ValidateCommissionAndTax validates broker commission and tax rates
func (v *NEPSEValidator) ValidateCommissionAndTax(transactionValue models.Money, commission, tax models.Money) error {
	// Standard NEPSE commission: 0.25% of transaction value (max Rs. 1,000)
	standardCommission := transactionValue.Multiply(0.0025) // 0.25%
	maxCommission := models.NewMoney(1000)                  // Rs. 1,000 cap

	if standardCommission.Compare(maxCommission) > 0 {
		standardCommission = maxCommission
	}

	// Allow some tolerance (±10%) for commission variations
	minCommission := standardCommission.Multiply(0.90)
	maxAllowedCommission := standardCommission.Multiply(1.10)

	if commission.Compare(minCommission) < 0 || commission.Compare(maxAllowedCommission) > 0 {
		return fmt.Errorf("commission Rs.%s outside expected range Rs.%s-Rs.%s (standard: Rs.%s)", 
			commission.FormatSimple(), 
			minCommission.FormatSimple(), 
			maxAllowedCommission.FormatSimple(),
			standardCommission.FormatSimple())
	}

	// Standard NEPSE SEBON fee: 0.015% of transaction value
	standardTax := transactionValue.Multiply(0.00015) // 0.015%
	toleranceTax := standardTax.Multiply(0.20)        // ±20% tolerance

	minTax := standardTax.Sub(toleranceTax)
	maxTax := standardTax.Add(toleranceTax)

	if tax.Compare(minTax) < 0 || tax.Compare(maxTax) > 0 {
		return fmt.Errorf("tax Rs.%s outside expected range Rs.%s-Rs.%s (standard SEBON fee: Rs.%s)", 
			tax.FormatSimple(),
			minTax.FormatSimple(),
			maxTax.FormatSimple(), 
			standardTax.FormatSimple())
	}

	return nil
}

// Helper methods

// getSymbolLotSize returns the lot size for a given symbol
func (v *NEPSEValidator) getSymbolLotSize(symbol string) int64 {
	// NEPSE lot sizes by category
	// Bank stocks: 10 shares
	// Insurance: 100 shares  
	// Others: 10 shares (default)
	
	bankSymbols := map[string]bool{
		"NABIL": true, "EBL": true, "KTM": true, "SBI": true, "HBL": true,
		"SANIMA": true, "MEGA": true, "PRVU": true, "SRBL": true, "GBIME": true,
	}
	
	insuranceSymbols := map[string]bool{
		"NICA": true, "PRIN": true, "SICL": true, "IGI": true, "UIC": true,
		"PICL": true, "NLICL": true, "LICN": true,
	}

	if bankSymbols[symbol] {
		return 10
	}
	if insuranceSymbols[symbol] {
		return 100
	}
	
	return 10 // Default lot size
}

// isKnownSymbol checks if symbol is in known NEPSE symbols list
func (v *NEPSEValidator) isKnownSymbol(symbol string) bool {
	// This is a basic list - in production, this would be fetched from NEPSE API
	knownSymbols := map[string]bool{
		// Banks
		"NABIL": true, "EBL": true, "KTM": true, "SBI": true, "HBL": true,
		"SANIMA": true, "MEGA": true, "PRVU": true, "SRBL": true, "GBIME": true,
		
		// Insurance
		"NICA": true, "PRIN": true, "SICL": true, "IGI": true, "UIC": true,
		"PICL": true, "NLICL": true, "LICN": true,
		
		// Hydro
		"HIDCL": true, "NHPC": true, "CHCL": true, "DHPL": true, "AKPL": true,
		
		// Others
		"ADBL": true, "NTC": true, "NMB": true, "PCBL": true, "LBL": true,
	}
	
	return knownSymbols[symbol]
}

// isSameDay checks if two times are on the same day
func (v *NEPSEValidator) isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}