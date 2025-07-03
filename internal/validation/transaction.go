package validation

import (
	"fmt"
	"ntx/internal/portfolio/services"
)

// TransactionValidator provides comprehensive transaction validation
type TransactionValidator struct {
	nepseValidator *NEPSEValidator
}

// NewTransactionValidator creates a new transaction validator
func NewTransactionValidator() *TransactionValidator {
	return &TransactionValidator{
		nepseValidator: NewNEPSEValidator(),
	}
}

// ValidationLevel defines the strictness of validation
type ValidationLevel int

const (
	ValidationBasic  ValidationLevel = iota // Basic business rules only
	ValidationStrict                        // Full NEPSE compliance
	ValidationLenient                       // NEPSE rules with warnings
)

// ValidationResult holds validation outcome
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// IsValid returns true if validation passed without errors
func (r *ValidationResult) IsValid() bool {
	return r.Valid && len(r.Errors) == 0
}

// HasWarnings returns true if there are warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// AddError adds an error to the validation result
func (r *ValidationResult) AddError(err string) {
	r.Errors = append(r.Errors, err)
	r.Valid = false
}

// AddWarning adds a warning to the validation result
func (r *ValidationResult) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

// ValidateTransaction validates a complete transaction request
func (v *TransactionValidator) ValidateTransaction(req services.ExecuteTransactionRequest, level ValidationLevel) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Basic validation (always performed)
	v.validateBasicRules(req, result)

	// NEPSE-specific validation (if strict or lenient)
	if level == ValidationStrict || level == ValidationLenient {
		v.validateNEPSERules(req, result, level == ValidationLenient)
	}

	return result
}

// validateBasicRules performs basic business rule validation
func (v *TransactionValidator) validateBasicRules(req services.ExecuteTransactionRequest, result *ValidationResult) {
	// Portfolio ID validation
	if req.PortfolioID <= 0 {
		result.AddError("Portfolio ID is required and must be positive")
	}

	// Symbol validation
	if req.Symbol == "" {
		result.AddError("Stock symbol is required")
	}

	// Transaction type validation
	if req.TransactionType != "buy" && req.TransactionType != "sell" {
		result.AddError("Transaction type must be 'buy' or 'sell'")
	}

	// Quantity validation
	if req.Quantity <= 0 {
		result.AddError("Quantity must be positive")
	}

	// Price validation
	if req.Price.IsZero() || req.Price.IsNegative() {
		result.AddError("Price must be positive")
	}

	// Commission validation (if provided)
	if !req.Commission.IsZero() && req.Commission.IsNegative() {
		result.AddError("Commission cannot be negative")
	}

	// Tax validation (if provided)
	if !req.Tax.IsZero() && req.Tax.IsNegative() {
		result.AddError("Tax cannot be negative")
	}

	// Date validation (basic)
	if req.TransactionDate.IsZero() {
		result.AddError("Transaction date is required")
	}
}

// validateNEPSERules performs NEPSE-specific validation
func (v *TransactionValidator) validateNEPSERules(req services.ExecuteTransactionRequest, result *ValidationResult, lenient bool) {
	// Symbol format validation
	if err := v.nepseValidator.ValidateSymbol(req.Symbol); err != nil {
		if lenient {
			result.AddWarning(fmt.Sprintf("Symbol format: %s", err.Error()))
		} else {
			result.AddError(err.Error())
		}
	}

	// Lot size validation
	if err := v.nepseValidator.ValidateLotSize(req.Symbol, req.Quantity); err != nil {
		if lenient {
			result.AddWarning(fmt.Sprintf("Lot size: %s", err.Error()))
		} else {
			result.AddError(err.Error())
		}
	}

	// Trading hours validation
	if err := v.nepseValidator.ValidateTradingHours(req.TransactionDate); err != nil {
		if lenient {
			result.AddWarning(fmt.Sprintf("Trading hours: %s", err.Error()))
		} else {
			result.AddError(err.Error())
		}
	}

	// Minimum transaction value validation
	if err := v.nepseValidator.ValidateMinimumTransaction(req.Quantity, req.Price); err != nil {
		if lenient {
			result.AddWarning(fmt.Sprintf("Minimum value: %s", err.Error()))
		} else {
			result.AddError(err.Error())
		}
	}

	// Commission and tax validation (if provided)
	if !req.Commission.IsZero() || !req.Tax.IsZero() {
		transactionValue := req.Price.MultiplyInt(req.Quantity)
		if err := v.nepseValidator.ValidateCommissionAndTax(transactionValue, req.Commission, req.Tax); err != nil {
			if lenient {
				result.AddWarning(fmt.Sprintf("Fees: %s", err.Error()))
			} else {
				result.AddError(err.Error())
			}
		}
	}

	// Price limit validation (requires last price - skip for now)
	// TODO: Integrate with price data to validate against daily limits
}

// ValidateSymbolFormat validates just the symbol format (for autocomplete/input validation)
func (v *TransactionValidator) ValidateSymbolFormat(symbol string) error {
	return v.nepseValidator.ValidateSymbol(symbol)
}

// ValidateQuantityForSymbol validates quantity against lot size for a symbol
func (v *TransactionValidator) ValidateQuantityForSymbol(symbol string, quantity int64) error {
	return v.nepseValidator.ValidateLotSize(symbol, quantity)
}

// GetLotSize returns the lot size for a given symbol
func (v *TransactionValidator) GetLotSize(symbol string) int64 {
	return v.nepseValidator.getSymbolLotSize(symbol)
}

// SuggestCorrectQuantity suggests the nearest valid quantity for a symbol
func (v *TransactionValidator) SuggestCorrectQuantity(symbol string, quantity int64) int64 {
	lotSize := v.GetLotSize(symbol)
	if quantity%lotSize == 0 {
		return quantity // Already valid
	}
	
	// Round down to nearest lot size
	return (quantity / lotSize) * lotSize
}