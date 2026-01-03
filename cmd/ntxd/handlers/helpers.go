package handlers

import (
	"database/sql"
	"errors"
	"math"
	"time"
)

var errSymbolRequired = errors.New("symbol is required")

const (
	defaultLimit = 50
	maxLimit     = 500
)

// clampLimit ensures limit is within valid bounds.
func clampLimit(limit int64, defaultVal int64) int64 {
	if limit <= 0 {
		return defaultVal
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

// safeInt32 converts int64 to int32, clamping to int32 bounds to prevent overflow.
func safeInt32(n int64) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}

// safeIntToInt32 converts int to int32, clamping to int32 bounds to prevent overflow.
func safeIntToInt32(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}

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

// parseDate parses a date string in YYYY-MM-DD format, returning time.Time{} if parsing fails.
func parseDate(dateStr string) time.Time {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return date
}
