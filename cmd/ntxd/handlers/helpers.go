package handlers

import (
	"database/sql"
	"errors"
)

var errSymbolRequired = errors.New("symbol is required")

// nullFloat64 extracts value from sql.NullFloat64, returning 0 if null.
func nullFloat64(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}

// nullInt64 extracts value from sql.NullInt64, returning 0 if null.
func nullInt64(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}

// nullString extracts value from sql.NullString, returning empty string if null.
func nullString(n sql.NullString) string {
	if n.Valid {
		return n.String
	}
	return ""
}
