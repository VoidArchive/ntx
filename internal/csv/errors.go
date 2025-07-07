package csv

import "errors"

// Custom error types for better error handling
var (
	ErrEmptyFile       = errors.New("CSV file is empty")
	ErrNoDataRows      = errors.New("CSV file has no data rows")
	ErrInvalidQuantity = errors.New("no valid quantity found")
	ErrRequiredField   = errors.New("required field is missing")
)
